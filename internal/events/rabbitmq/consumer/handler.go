package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gothunder/thunder/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack/v5"
)

func (r *rabbitmqConsumer) handler(msgs <-chan amqp.Delivery, handler events.Handler) {
	for msg := range msgs {
		logger := r.logger.With().
			Str("topic", msg.RoutingKey).Logger()
		ctx := logger.WithContext(context.Background())

		var decoder events.EventDecoder
		if msg.ContentType == "application/json" {
			decoder = json.NewDecoder(bytes.NewReader(msg.Body))
		} else if msg.ContentType == "application/msgpack" {
			decoder = msgpack.NewDecoder(bytes.NewReader(msg.Body))
		}

		res := r.handleWithRecoverer(ctx, handler, msg.RoutingKey, decoder)

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

func (r *rabbitmqConsumer) handleWithRecoverer(ctx context.Context, handler events.Handler, topic string, decoder events.EventDecoder) (res events.HandlerResponse) {
	logger := log.Ctx(ctx).With().Stack().Logger()
	logger.Info().Msg("consuming message")

	// Catch panics
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = eris.New(fmt.Sprintf("%v", r))
			}

			err = eris.Wrap(err, "panic")
			logger.Error().Err(err).Msg("panic while consuming message")

			// If there's a panic, we should stop processing the message
			res = events.DeadLetter
		}
	}()

	return handler.Handle(ctx, topic, decoder)
}
