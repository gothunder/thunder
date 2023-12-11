package outbox

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	SpanNameStore = "thunder.outbox.storer.Store"
	TracerName    = "thunder.outbox.storer"
)

type withTracingStorer struct {
	next   Storer
	prop   propagation.TextMapPropagator
	tracer trace.Tracer
}

// Store implements Storer.
func (wts withTracingStorer) Store(ctx context.Context, tx interface{}, messages []Message) error {
	tctx, span := wts.tracer.Start(
		ctx, SpanNameStore,
		trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	for i := range messages {
		wts.prop.Inject(tctx, propagation.MapCarrier(messages[i].Headers))
	}

	err := wts.next.Store(tctx, tx, messages)
	if err != nil {
		span.RecordError(err)
	}

	return err
}

// WithTxClient implements Storer.
func (wts *withTracingStorer) WithTxClient(tx interface{}) (TransactionalStorer, error) {
	// needs to be reimplemented here or otherwise TransactionalStorer
	// will not have tracing
	return newTransactionalStorer(wts, tx)
}

func WrapStorerWithTracing(next Storer) Storer {
	return &withTracingStorer{
		next: next,
		prop: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
		tracer: otel.Tracer(TracerName)}
}
