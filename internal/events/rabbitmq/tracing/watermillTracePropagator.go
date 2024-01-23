package tracing

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel/propagation"
)

type WatermillTracePropagator struct {
	propagator propagation.TextMapPropagator
}

func NewWatermillTracePropagator() *WatermillTracePropagator {
	return &WatermillTracePropagator{
		propagator: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	}
}

func (r *WatermillTracePropagator) WithTrace(ctx context.Context, msg *message.Message) *message.Message {
	r.propagator.Inject(ctx, propagation.MapCarrier(msg.Metadata))
	return msg
}

func (r *WatermillTracePropagator) ExtractTrace(ctx context.Context, msg *message.Message) context.Context {
	return r.propagator.Extract(ctx, propagation.MapCarrier(msg.Metadata))
}
