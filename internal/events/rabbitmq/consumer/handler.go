package consumer

import (
	"context"

	"github.com/gothunder/thunder/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *rabbitmqConsumer) handler(msgs <-chan amqp.Delivery, handlers routingKeyHandlerMap) {
	r.wg.Add(1)

	for msg := range msgs {
		ctx := r.logger.WithContext(context.Background())

		handler := handlers[msg.RoutingKey]
		if handler == nil {
			r.logger.Error().Msgf("no handler for routing key %v", msg.RoutingKey)
			msg.Nack(false, false)
			continue
		}

		res := handler(ctx, events.Event{
			Topic:   msg.RoutingKey,
			Payload: msg.Body, // TODO unmarshal the message
		})

		switch res {
		case events.Success:
			// Message was successfully processed
			err := msg.Ack(false)
			if err != nil {
				r.logger.Error().Err(err).Msg("failed to ack message")
			}
		case events.Requeue:
			// We should retry to process the message on a different worker
			err := msg.Nack(false, true)
			if err != nil {
				r.logger.Error().Err(err).Msg("failed to requeue message")
			}
		default:
			// We should stop processing the message
			err := msg.Nack(false, false)
			if err != nil {
				r.logger.Error().Err(err).Msg("failed to discard message")
			}
		}
	}

	r.wg.Done()
}

type routingKeyHandlerMap map[string]events.HandlerFunc

func mapRoutingKeyToHandler(eventHandlers []events.EventHandler) (routingKeyHandlerMap, []string) {
	r := make(routingKeyHandlerMap)
	var routingKeys []string
	for _, eventHandler := range eventHandlers {
		r[eventHandler.Topic] = eventHandler.Handler
		routingKeys = append(routingKeys, eventHandler.Topic)
	}

	return r, routingKeys
}
