package rabbitmq

import (
	"context"
	"time"

	outboxpublisher "github.com/gothunder/thunder/internal/events/rabbitmq/outboxPublisher"
	"github.com/gothunder/thunder/internal/events/rabbitmq/publisher"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

func NewRabbitMQPublisher(logger *zerolog.Logger) (events.EventPublisher, error) {
	return publisher.NewPublisher(amqp091.Config{}, logger)
}

func provideRabbitMQPublisher(logger *zerolog.Logger, s fx.Shutdowner) events.EventPublisher {
	publisher, err := NewRabbitMQPublisher(logger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create publisher")
		err = s.Shutdown()
		if err != nil {
			logger.Error().Err(err).Msg("failed to shutdown")
		}
	}

	return publisher
}

func provideRabbitMQOutboxPublisher[T outboxpublisher.OutboxPublisherFactory](
	logger *zerolog.Logger,
	s fx.Shutdowner,
	forwardFactory outboxpublisher.ForwarderFactory,
	outboxPublisherFactoryCtxExtractor outboxpublisher.OutboxPublisherFactoryCtxExtractor[T],
) events.EventPublisher {
	publisher, err := outboxpublisher.NewRabbitMQOutboxPublisher(logger, forwardFactory, outboxPublisherFactoryCtxExtractor)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create publisher")
		err = s.Shutdown()
		if err != nil {
			logger.Error().Err(err).Msg("failed to shutdown")
		}
	}

	return publisher
}

func startPublisher(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, publisher events.EventPublisher) {
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					err := publisher.StartPublisher(context.Background())
					if err != nil {
						logger.Error().Err(err).Msg("failed to start publisher")
						err = s.Shutdown()
						if err != nil {
							logger.Error().Err(err).Msg("failed to shutdown")
						}
					}
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info().Msg("stopping publisher")

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
