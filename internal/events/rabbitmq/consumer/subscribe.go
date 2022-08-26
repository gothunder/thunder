package consumer

import (
	"context"

	"github.com/gothunder/thunder/pkg/events"
	"github.com/rotisserie/eris"
)

func (r *rabbitmqConsumer) Subscribe(
	ctx context.Context,
	eventHandlers []events.EventHandler,
) error {
	handlers, routingKeys := mapRoutingKeyToHandler(eventHandlers)

	for {
		err := r.startGoRoutines(handlers, routingKeys)
		if err != nil {
			return eris.Wrap(err, "failed to start go routines")
		}

		// Check if the channel reconnects
		err = <-r.chManager.NotifyReconnection
		if err != nil {
			return eris.Wrap(err, "failed to reconnect to the amqp channel")
		}

		r.logger.Info().Msg("restarting consumer after reconnection")
	}
}

func (r *rabbitmqConsumer) startGoRoutines(handlers routingKeyHandlerMap, routingKeys []string) error {
	err := r.declare(routingKeys)
	if err != nil {
		return err
	}

	msgs, err := r.chManager.Channel.Consume(
		r.config.QueueName,
		r.config.ConsumerName,
		false,
		true,
		false,
		false,
		nil,
	)

	for i := 0; i < r.config.ConsumerConcurrency; i++ {
		// The msg channel will be closed when the amqp channel is closed
		go r.handler(msgs, handlers)
	}
	r.logger.Info().Msgf("processing messages on %v goroutines", r.config.ConsumerConcurrency)

	return nil
}
