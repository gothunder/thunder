package consumer

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (r *rabbitmqConsumer) retryBackoff(ctx context.Context, msg amqp.Delivery) {
	r.backoffWg.Add(1)
	defer r.backoffWg.Done()
	ctx, span := otel.Tracer(scope).Start(ctx, "rabbitmqConsumer.retryBackoff",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	logger := log.Ctx(ctx).With().Stack().Ctx(ctx).Logger()

	// Get the current retry count
	attempts, ok := msg.Headers["x-delivery-count"]
	if !ok {
		attempts = int64(0)
	}

	logger.Info().Msgf("message has been attempted %d times", attempts.(int64))
	span.SetAttributes(attribute.Key("message.attempts.count").Int64(attempts.(int64)))

	if attempts.(int64) >= int64(r.config.MaxRetries) {
		logger.Info().Msg("message has reached max retries")
		// We should stop processing the message
		err := msg.Nack(false, false)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to put message in dead letter queue")
			logger.Error().Err(err).Msg("failed to put message in dead letter queue")
		}
		return
	}

	backOff := backoff.NewExponentialBackOff()
	backOff.InitialInterval = r.config.InitialInterval
	backOff.RandomizationFactor = r.config.RandomizationFactor
	backOff.Multiplier = r.config.Multiplier
	backOff.MaxInterval = r.config.MaxInterval

	interval := backOff.NextBackOff()
	for i := 0; int64(i) < attempts.(int64); i++ {
		interval = backOff.NextBackOff()
	}

	if interval == backoff.Stop {
		// We should stop processing the message
		err := msg.Nack(false, false)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to put message in dead letter queue")
			logger.Error().Err(err).Msg("failed to put message in dead letter queue")
		}
		return
	}

	time.Sleep(interval)

	err := msg.Nack(false, true)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to requeue message")
		logger.Error().Err(err).Msg("failed to requeue message")
	}
}
