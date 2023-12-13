package outbox

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	SpanNameRelay     = "thunder.outbox.relayer.Relay"
	TracerNameRelayer = "thunder.outbox.relayer"

	SpanNamePoll   = "thunder.outbox.relayer.Poll"
	TracerNamePoll = "thunder.outbox.poller"
)

type withTracingRelayer struct {
	next   Relayer
	prop   propagation.TextMapPropagator
	tracer trace.Tracer
}

// Close implements Relayer.
func (w *withTracingRelayer) Close(ctx context.Context) error {
	return w.next.Close(ctx)
}

// Start implements Relayer.
func (w *withTracingRelayer) Start(ctx context.Context) error {
	return w.next.Start(ctx)
}

// prepareMessages implements Relayer.
func (w *withTracingRelayer) prepareMessages(msgPack []*Message) []*message.Message {
	return w.next.prepareMessages(msgPack)
}

// relay implements Relayer.
func (w *withTracingRelayer) relay(ctx context.Context, msgPack []*Message) error {
	spans := make([]trace.Span, len(msgPack))

	for i := range msgPack {
		tctx := w.prop.Extract(ctx, propagation.MapCarrier(msgPack[i].Headers))
		ntctx, span := w.tracer.Start(
			tctx, SpanNameStore,
			trace.WithSpanKind(trace.SpanKindProducer))
		defer span.End()
		spans[i] = span

		w.prop.Inject(ntctx, propagation.MapCarrier(msgPack[i].Headers))
	}

	err := w.next.relay(ctx, msgPack)
	if err != nil {
		for i := range spans {
			spans[i].RecordError(err)
		}
	}

	return err
}

func WrapRelayerWithTracing(next Relayer) Relayer {
	return &withTracingRelayer{
		next: next,
		prop: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
		tracer: otel.Tracer(TracerNameRelayer)}
}

type withTracingMessagePoller struct {
	next   MessagePoller
	prop   propagation.TextMapPropagator
	tracer trace.Tracer
}

// Close implements MessagePoller.
func (w *withTracingMessagePoller) Close() error {
	return w.next.Close()
}

// Poll implements MessagePoller.
func (w *withTracingMessagePoller) Poll(ctx context.Context) (<-chan []*Message, func(), error) {
	tMessageChan := make(chan []*Message)

	messages, next, err := w.next.Poll(ctx)
	if err != nil {
		return messages, next, err
	}

	go func() {
		defer close(tMessageChan)

		endSpans := func() {}
		defer endSpans()

		for msgPack := range messages {
			endSpans()
			spans := make([]trace.Span, len(msgPack))
			for i := range msgPack {
				tctx := w.prop.Extract(ctx, propagation.MapCarrier(msgPack[i].Headers))
				ntctx, span := w.tracer.Start(
					tctx, SpanNamePoll,
					trace.WithSpanKind(trace.SpanKindConsumer))
				spans[i] = span
				w.prop.Inject(ntctx, propagation.MapCarrier(msgPack[i].Headers))

				endSpans = func() {
					for i := range spans {
						spans[i].End()
					}
				}
			}

			tMessageChan <- msgPack
		}

	}()

	return tMessageChan, next, err
}

func WrapPollerWithTracing(next MessagePoller) MessagePoller {
	return &withTracingMessagePoller{
		next: next,
		prop: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
		tracer: otel.Tracer(TracerNamePoll)}
}
