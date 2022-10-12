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

func registerConsumer(topics []string) interface{} {
	fn := func(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, handler events.Handler) {
		consumer, err := NewRabbitMQConsumer(logger)
		if err != nil {
			logger.Err(err).Msg("failed to create consumer")
			err = s.Shutdown()
			if err != nil {
				logger.Err(err).Msg("failed to shutdown")
			}
		}

		lc.Append(
			fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						err := consumer.Subscribe(ctx, topics, handler.Handle)
						if err != nil {
							logger.Err(err).Msg("failed to subscribe to topics")
							err = s.Shutdown()
							if err != nil {
								logger.Err(err).Msg("failed to shutdown")
							}
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

// A module that provides a RabbitMQ consumer.
// The consumer will be automatically started and stopped gracefully.
// The consumer will subscribe to the provided topics.
// The handler will be called when a message is received.
// The handler will be called concurrently
// The application will shutdown if the consumer fails to start or reconnect.
func InvokeConsumer(topics []string) fx.Option {
	return fx.Invoke(registerConsumer(topics))
}
