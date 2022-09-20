package rabbitmq

import (
	"context"

	"github.com/gothunder/thunder/internal/events/rabbitmq/consumer"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

func NewRabbitMQConsumer(logger *zerolog.Logger) (events.EventConsumer, error) {
	return consumer.NewConsumer(amqp091.Config{}, logger)
}

func registerConsumer(topics []string, handler events.HandlerFunc) interface{} {
	fn := func(lc fx.Lifecycle, logger *zerolog.Logger) {
		consumer, err := NewRabbitMQConsumer(logger)
		if err != nil {
			// TODO shutdown
			logger.Err(err).Msg("failed to create consumer")
			panic(err)
		}

		lc.Append(
			fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						err := consumer.Subscribe(ctx, topics, handler)
						if err != nil {
							// TODO shutdown
							logger.Err(err).Msg("failed to subscribe to topics")
						}
					}()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info().Msg("stopping consumer")

					err := consumer.Close(ctx)
					if err != nil {
						logger.Err(err).Msg("error closing consumer")
						return err
					}

					logger.Info().Msg("consumer stopped")
					return nil
				},
			},
		)
	}

	return fn
}

func InvokeConsumer(topics []string, handler events.HandlerFunc) fx.Option {
	return fx.Invoke(registerConsumer(topics, handler))
}
