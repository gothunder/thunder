package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	outboxpublisher "github.com/gothunder/thunder/internal/events/rabbitmq/outboxPublisher"
	"github.com/gothunder/thunder/internal/events/rabbitmq/publisher"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

// shutdown tears the application down after the RabbitMQ transport is gone.
// Both the publisher and the consumer use it, so a service never keeps
// running with nothing attached to the broker.
func shutdown(logger *zerolog.Logger, s fx.Shutdowner) {
	if err := s.Shutdown(fx.ExitCode(1)); err != nil {
		logger.Error().Err(err).Msg("failed to shutdown")
	}
}

func NewRabbitMQPublisher(logger *zerolog.Logger) (events.EventPublisher, error) {
	return publisher.NewPublisher(amqp091.Config{}, logger)
}

func provideRabbitMQPublisher(logger *zerolog.Logger) (events.EventPublisher, error) {
	publisher, err := NewRabbitMQPublisher(logger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create publisher")
		return nil, fmt.Errorf("create rabbitmq publisher: %w", err)
	}

	return publisher, nil
}

func provideRabbitMQOutboxPublisher[T outboxpublisher.OutboxPublisherFactory](
	logger *zerolog.Logger,
	forwardFactory outboxpublisher.ForwarderFactory,
	outboxPublisherFactoryCtxExtractor outboxpublisher.OutboxPublisherFactoryCtxExtractor[T],
) (events.EventPublisher, error) {
	publisher, err := outboxpublisher.NewRabbitMQOutboxPublisher(logger, forwardFactory, outboxPublisherFactoryCtxExtractor)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create publisher")
		// Shutting down from inside a provider still lets startup succeed and
		// hands the nil publisher to whoever depends on it; returning the
		// error refuses to build the graph at all, and matches the plain
		// publisher path.
		return nil, fmt.Errorf("create rabbitmq outbox publisher: %w", err)
	}

	return publisher, nil
}

func startPublisher(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, publisher events.EventPublisher) {
	var stopping atomic.Bool

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					err := publisher.StartPublisher(context.Background())
					if err == nil || errors.Is(err, context.Canceled) || stopping.Load() {
						return
					}

					logger.Error().Err(err).Msg("failed to start publisher")
					shutdown(logger, s)
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info().Msg("stopping publisher")
				stopping.Store(true)

				// Create a new context with a timeout of 5 seconds
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

				err := publisher.Close(ctx)
				cancel()
				if err != nil {
					logger.Error().Err(err).Msg("error closing publisher")
					return err
				}

				logger.Info().Msg("publisher stopped")
				return nil
			},
		},
	)
}

// A module that provides a RabbitMQ publisher.
// The publisher will be provided to the application.
// The publisher is automatically started and stopped gracefully.
// The application will shutdown if the publisher fails to start or reconnect.
var PublisherModule = fx.Options(
	fx.Provide(provideRabbitMQPublisher),
	fx.Invoke(startPublisher),
)

func OutboxPublisherModule[T outboxpublisher.OutboxPublisherFactory](
	outboxPublisherFactoryCtxExtractor outboxpublisher.OutboxPublisherFactoryCtxExtractor[T],
) fx.Option {
	return fx.Options(
		fx.Provide(provideRabbitMQOutboxPublisher[T]),
		fx.Supply(outboxPublisherFactoryCtxExtractor),
		fx.Invoke(startPublisher),
	)
}

func UseForwarderFactory(factory interface{}) fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(func(client outboxpublisher.ForwarderFactory) outboxpublisher.ForwarderFactory {
				return client
			}, fx.From(factory)),
		),
	)
}
