package consumer

// Graceful shutdown of the consumer.
func (r *rabbitmqConsumer) Close() error {
	r.logger.Info().Msg("closing consumer")

	// First we stop sending new messages to the consumer
	r.chManager.Channel.Cancel(r.config.ConsumerName, true)

	// We'll block till all the active go routines are done
	r.wg.Wait()

	// Now we'll close the channel
	err := r.chManager.Close()
	if err != nil {
		return err
	}

	r.logger.Info().Msg("consumer closed gracefully")
	return nil
}
