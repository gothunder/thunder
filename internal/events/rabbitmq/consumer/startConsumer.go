package consumer

func (r rabbitmqConsumer) startConsumer(routingKeys []string) {
	go func() {
		for {
			err := r.startGoRoutines(routingKeys)
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
	}()
}

func (r rabbitmqConsumer) startGoRoutines(routingKeys []string) error {
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
		go r.handler(msgs)
	}
	r.logger.Info().Msgf("processing messages on %v goroutines", r.config.ConsumerConcurrency)

	return nil
}
