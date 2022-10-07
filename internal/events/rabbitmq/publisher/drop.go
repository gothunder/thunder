package publisher

func (r *rabbitmqPublisher) dropMessage(msg message) {
	r.logger.Error().
		Str("topic", msg.Topic).
		Bytes("payload", msg.Message.Body).
		Msg("event dropped")
	r.wg.Done()
}
