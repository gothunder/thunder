package publisher

// Graceful shutdown of the publisher.
func (r rabbitmqPublisher) Close() error {
	// TODO check if there's any handler running and if so, wait for them to finish.

	r.logger.Info().Msg("amqp closing publisher...")
	return r.chManager.Close()
}
