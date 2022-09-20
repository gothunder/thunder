package rabbitmq

import (
	"context"

	"github.com/gothunder/thunder/internal/events/rabbitmq/publisher"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

func NewRabbitMQPublisher(logger *zerolog.Logger) (events.EventPublisher, error) {
	return publisher.NewPublisher(amqp091.Config{}, logger)
}

func provideRabbitMQPublisher(logger *zerolog.Logger) events.EventPublisher {
	publisher, err := NewRabbitMQPublisher(logger)
	if err != nil {
		logger.Err(err).Msg("failed to create publisher")
		panic(err)
	}

	return publisher
}

func startPublisher(lc fx.Lifecycle, logger *zerolog.Logger, publisher events.EventPublisher) {
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					err := publisher.StartPublisher(ctx)
					if err != nil {
						// TODO shutdown
						logger.Err(err).Msg("failed to start publisher")
					}
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info().Msg("stopping publisher")

				err := publisher.Close(ctx)
				if err != nil {
					logger.Err(err).Msg("error closing publisher")
					return err
				}

				logger.Info().Msg("publisher stopped")
				return nil
			},
		},
	)
}

var PublisherModule = fx.Options(
	fx.Provide(provideRabbitMQPublisher),
	fx.Invoke(startPublisher),
)
