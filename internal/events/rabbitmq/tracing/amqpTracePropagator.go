package tracing

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/propagation"
)

type amqpMapCarrier map[string]interface{}

func (c amqpMapCarrier) Get(key string) string {
	if v, ok := c[key]; ok {
		if strv, ok := v.(string); ok {
			return strv
		}
	}
	return ""
}

func (c amqpMapCarrier) Set(key, value string) {
	c[key] = value
}

func (c amqpMapCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

type AmqpTracePropagator struct {
	propagator propagation.TextMapPropagator
}

func NewAmqpTracing() *AmqpTracePropagator {
	return &AmqpTracePropagator{
		propagator: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	}
}

func (r *AmqpTracePropagator) WithTrace(ctx context.Context, msg *amqp091.Publishing) *amqp091.Publishing {
	r.propagator.Inject(ctx, amqpMapCarrier(msg.Headers))
	return msg
}

func (r *AmqpTracePropagator) ExtractTrace(ctx context.Context, msg *amqp091.Delivery) context.Context {
	return r.propagator.Extract(ctx, amqpMapCarrier(msg.Headers))
}
