package rabbitmq

import (
	"context"

	"github.com/gothunder/thunder/internal/events/rabbitmq"
	"github.com/gothunder/thunder/internal/events/rabbitmq/consumer"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

type namedHandlerParams struct {
	fx.In
	NamedHandlers []events.NamedHandler `group:"named_handlers"`
}

func registerNamedConsumers(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, params namedHandlerParams) {
	for _, namedHandler := range params.NamedHandlers {
		registerNamedConsumer(lc, s, logger, namedHandler)
	}
}

func registerNamedConsumer(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, namedHandler events.NamedHandler) {
	consumer, err := NewRabbitMQConsumer(logger, WithQueueNamePosfix(namedHandler.QueuePosfix()))
	if err != nil {
		logger.Error().Err(err).Msg("failed to create consumer")
		err = s.Shutdown()
		if err != nil {
			logger.Error().Err(err).Msg("failed to shutdown")
		}
	}

	registerProvidedConsumer(lc, s, logger, namedHandler, consumer)
}

func registerConsumer(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, handler events.Handler) {
	consumer, err := NewRabbitMQConsumer(logger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create consumer")
		err = s.Shutdown()
		if err != nil {
			logger.Error().Err(err).Msg("failed to shutdown")
		}
	}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					err := consumer.Subscribe(ctx, handler)
					if err != nil {
						logger.Error().Err(err).Msg("failed to subscribe to topics")
						err = s.Shutdown()
						if err != nil {
							logger.Error().Err(err).Msg("failed to shutdown")
						}
					}
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info().Msg("stopping consumer")

				err := consumer.Close(ctx)
				if err != nil {
					logger.Error().Err(err).Msg("error closing consumer")
					return err
				}

				logger.Info().Msg("consumer stopped")
				return nil
			},
		},
	)
}

func registerProvidedConsumer(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, handler events.Handler, consumer events.EventConsumer) {
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					err := consumer.Subscribe(ctx, handler)
					if err != nil {
						logger.Error().Err(err).Msg("failed to subscribe to topics")
						err = s.Shutdown()
						if err != nil {
							logger.Error().Err(err).Msg("failed to shutdown")
						}
					}
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info().Msg("stopping consumer")

				err := consumer.Close(ctx)
				if err != nil {
					logger.Error().Err(err).Msg("error closing consumer")
					return err
				}

				logger.Info().Msg("consumer stopped")
				return nil
			},
		},
	)
}

func NewRabbitMQConsumer(logger *zerolog.Logger, opts ...rabbitmq.RabbitmqConfigOption) (events.EventConsumer, error) {
	return consumer.NewConsumer(amqp091.Config{}, logger, opts...)
}

// A module that provides a RabbitMQ consumer.
// The consumer will be automatically started and stopped gracefully.
// The consumer will subscribe to the provided topics.
// The handler will be called when a message is received.
// The handler will be called concurrently
// The application will shutdown if the consumer fails to start or reconnect.
var InvokeConsumer = fx.Invoke(
	registerConsumer,
)

var InvokeProvidedConsumer = fx.Invoke(
	registerProvidedConsumer,
)

var InvokeNamedConsumers = fx.Invoke(
	registerNamedConsumers,
)

var InvokeNamedConsumer = fx.Invoke(
	registerNamedConsumer,
)
