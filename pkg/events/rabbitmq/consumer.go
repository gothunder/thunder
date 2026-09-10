package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

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

func registerNamedConsumers(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, params namedHandlerParams) error {
	for _, namedHandler := range params.NamedHandlers {
		if err := registerNamedConsumer(lc, s, logger, namedHandler); err != nil {
			return err
		}
	}

	return nil
}

func registerNamedConsumer(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, namedHandler events.NamedHandler) error {
	consumer, err := NewRabbitMQConsumer(logger, WithQueueNamePosfix(namedHandler.QueuePosfix()))
	if err != nil {
		logger.Error().Err(err).Msg("failed to create consumer")
		return fmt.Errorf("create rabbitmq consumer: %w", err)
	}

	return registerProvidedConsumer(lc, s, logger, namedHandler, consumer)
}

func registerConsumer(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, handler events.Handler) error {
	consumer, err := NewRabbitMQConsumer(logger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create consumer")
		return fmt.Errorf("create rabbitmq consumer: %w", err)
	}

	return registerProvidedConsumer(lc, s, logger, handler, consumer)
}

func registerProvidedConsumer(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, handler events.Handler, consumer events.EventConsumer) error {
	var (
		cancel   context.CancelFunc
		stopping atomic.Bool
	)

	lc.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				// The subscription outlives the start hook, so it cannot use
				// the OnStart context: fx cancels that one as soon as startup
				// finishes.
				var subscribeCtx context.Context
				subscribeCtx, cancel = context.WithCancel(context.Background())

				go func() {
					err := consumer.Subscribe(subscribeCtx, handler)
					if err == nil || errors.Is(err, context.Canceled) || stopping.Load() {
						return
					}

					logger.Error().Err(err).Msg("failed to subscribe to topics")
					shutdown(logger, s)
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info().Msg("stopping consumer")
				stopping.Store(true)

				if cancel != nil {
					cancel()
				}

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

	return nil
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
