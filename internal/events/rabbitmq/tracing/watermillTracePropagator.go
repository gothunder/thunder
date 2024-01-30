package tracing

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel/propagation"
)

type WatermillTracePropagator struct {
	propagator propagation.TextMapPropagator
}

// Creates a new watermill trace propagator
func NewWatermillTracePropagator() *WatermillTracePropagator {
	return &WatermillTracePropagator{
		propagator: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	}
}

// WithTrace injects the trace context into the message metadata
// It will be mapped to the message headers depending on the message broker
// As we are using watermill, it automatically handle the mapping
// Considering the broker is rabbitmq, the metadata will be mapped to the message headers
func (r *WatermillTracePropagator) WithTrace(ctx context.Context, msg *message.Message) *message.Message {
	r.propagator.Inject(ctx, propagation.MapCarrier(msg.Metadata))
	return msg
}

// As opposed to WithTrace, ExtractTrace extracts the trace context from the message metadata
// The metadata is mapped from the message headers depending on the message broker
// As we are using watermill, it automatically handle the mapping
// Considering the broker is rabbitmq, the metadata will be mapped from the message headers
func (r *WatermillTracePropagator) ExtractTrace(ctx context.Context, msg *message.Message) context.Context {
	return r.propagator.Extract(ctx, propagation.MapCarrier(msg.Metadata))
}
