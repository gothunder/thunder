package consumer

import (
	"context"
	"time"

	"github.com/TheRafaBonin/roxy"
	"github.com/cenkalti/backoff/v4"
	"github.com/gothunder/thunder/internal/events/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	confirmTimeout = 5 * time.Second

	deliveryCountHeader = "x-delivery-count"
)

func (r *rabbitmqConsumer) retryBackoff(ctx context.Context, msg amqp.Delivery) {
	r.backoffWg.Add(1)
	defer r.backoffWg.Done()
	ctx, span := otel.Tracer(scope).Start(ctx, "rabbitmqConsumer.retryBackoff",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	logger := log.Ctx(ctx).With().Stack().Ctx(ctx).Logger()

	attempts := deliveryCount(msg)

	logger.Info().Msgf("message has been attempted %d times", attempts)
	span.SetAttributes(attribute.Key("message.attempts.count").Int64(attempts))

	if attempts >= int64(r.config.MaxRetries) {
		logger.Info().Msg("message has reached max retries")
		// We should stop processing the message
		err := msg.Nack(false, false)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to put message in dead letter queue")
			logger.Error().Err(err).Msg("failed to put message in dead letter queue")
			return
		}
		logger.Info().Msg("message has been put in dead letter queue")
		return
	}

	backOff := newBackoff(r.config)
	interval := currentInterval(backOff, attempts)

	if interval == backoff.Stop {
		// We should stop processing the message
		logger.Info().Msg("backoff has reached max interval")
		err := msg.Nack(false, false)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to put message in dead letter queue")
			logger.Error().Err(err).Msg("failed to put message in dead letter queue")
			return
		}
		logger.Info().Msg("message has been put in dead letter queue")
		return
	}

	setDeliveryCount(&msg, attempts+1)
	logger.Info().Msgf("requeueing message in %f seconds", interval.Seconds())
	time.Sleep(interval)

	err := requeue(ctx, r.chManager.Channel, r.config.QueueName, msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to requeue message")
		logger.Error().Err(err).Msg("failed to requeue message")
		return
	}

	logger.Info().Msg("message has been requeued")
}

func deliveryCount(msg amqp.Delivery) int64 {
	attempts, ok := msg.Headers[deliveryCountHeader]
	if !ok {
		return 0
	}

	return attempts.(int64)
}

func setDeliveryCount(msg *amqp.Delivery, attempts int64) {
	msg.Headers[deliveryCountHeader] = attempts
}

// newBackoff creates a new backoff policy given the config.
func newBackoff(config rabbitmq.Config) backoff.BackOff {
	boff := backoff.NewExponentialBackOff()

	boff.InitialInterval = config.InitialInterval
	boff.RandomizationFactor = config.RandomizationFactor
	boff.Multiplier = config.Multiplier
	boff.MaxInterval = config.MaxInterval

	return boff
}

// currentInterval returns the current interval given the backoff policy and the number of attempts.
func currentInterval(backOff backoff.BackOff, attempts int64) time.Duration {
	interval := backOff.NextBackOff()
	for i := 0; int64(i) < attempts; i++ {
		interval = backOff.NextBackOff()
	}

	return interval
}

func requeue(ctx context.Context, channel *amqp091.Channel, queueName string, msg amqp.Delivery) error {
	ctx, cancel := context.WithTimeout(ctx, confirmTimeout)
	defer cancel()

	err := channel.PublishWithContext(
		ctx,
		"",
		queueName,
		true,
		false,
		amqp091.Publishing{
			ContentType:  msg.ContentType,
			DeliveryMode: msg.DeliveryMode,
			Body:         msg.Body,
			Headers:      msg.Headers,
		},
	)
	if err != nil {
		return roxy.Wrap(err, "publishing message")
	}

	err = msg.Ack(false)
	if err != nil {
		return roxy.Wrap(err, "acknowledging message")
	}

	return nil
}
