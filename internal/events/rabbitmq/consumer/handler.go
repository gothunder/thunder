package consumer

import (
	"context"

	"github.com/gothunder/thunder/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *rabbitmqConsumer) handler(msgs <-chan amqp.Delivery, handler events.Handler) {
	for msg := range msgs {
		logger := r.logger.With().
			Str("topic", msg.RoutingKey).Logger()
		ctx := logger.WithContext(context.Background())

		logger.Info().Msg("consuming message")
		res := handler.Handle(ctx, msg.RoutingKey, msg.Body)

		switch res {
		case events.Success:
			// Message was successfully processed
			err := msg.Ack(false)
			if err != nil {
				logger.Error().Err(err).Msg("failed to ack message")
			}
		case events.DeadLetter:
			// We should retry to process the message on a different worker
			err := msg.Nack(false, false)
			if err != nil {
				logger.Error().Err(err).Msg("failed to requeue message")
			}
		default:
			// We should stop processing the message
			err := msg.Nack(false, true)
			if err != nil {
				logger.Error().Err(err).Msg("failed to discard message")
			}
		}
	}
}
