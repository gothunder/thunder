package consumer

import (
	"context"

	"github.com/gothunder/thunder/pkg/events"
)

func (r *rabbitmqConsumer) Subscribe(
	ctx context.Context,
	handlers []events.EventHandler,
) error {
	// TODO stop when there's a graceful shutdown
	for {
		err := r.startGoRoutines(handlers)
		if err != nil {
			// TODO handle error
			break
		}

		// Check if the channel reconnects
		err = <-r.chManager.NotifyReconnection
		if err != nil {
			// TODO handle error
			break
		}

		r.logger.Info().Msg("restarting consumer after reconnection")
	}

	return nil
}

func (r *rabbitmqConsumer) startGoRoutines(handlers []events.EventHandler) error {
	var routingKeys []string
	for _, handler := range handlers {
		routingKeys = append(routingKeys, handler.Topic)
	}

	err := r.declare(routingKeys)
	if err != nil {
		return err
	}

	msgs, err := r.chManager.Channel.Consume(
		r.config.QueueName,
		"",
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
