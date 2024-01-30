package tracing

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"

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
	logger     *zerolog.Logger
}

func NewAmqpTracing(logger *zerolog.Logger) *AmqpTracePropagator {
	log := logger.With().Str("component", "amqp-trace-propagator").Logger()
	return &AmqpTracePropagator{
		propagator: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
		logger: &log,
	}
}

func (r *AmqpTracePropagator) WithTrace(ctx context.Context, msg *amqp091.Publishing) *amqp091.Publishing {
	r.propagator.Inject(ctx, amqpMapCarrier(msg.Headers))
	if traceparent, ok := msg.Headers["traceparent"]; ok {
		r.logger.Debug().Ctx(ctx).Str("msg-headers", traceparent.(string)).Msg("injecting trace context into message headers")
	}
	defer r.logger.Debug().Ctx(ctx).Str("traceparent", amqpMapCarrier(msg.Headers).Get("traceparent")).Msg("trace context injected")
	return msg
}

func (r *AmqpTracePropagator) ExtractTrace(ctx context.Context, msg *amqp091.Delivery) context.Context {
	if traceparent, ok := msg.Headers["traceparent"]; ok {
		r.logger.Debug().Ctx(ctx).Str("msg-headers", traceparent.(string)).Msg("extracting trace context into message headers")
	}
	defer r.logger.Debug().Ctx(ctx).Str("traceparent", amqpMapCarrier(msg.Headers).Get("traceparent")).Msg("trace context extracted")
	return r.propagator.Extract(ctx, amqpMapCarrier(msg.Headers))
}
