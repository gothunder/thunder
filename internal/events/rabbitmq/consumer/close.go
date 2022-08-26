package consumer

// Graceful shutdown of the consumer.
func (r rabbitmqConsumer) Close() error {
	r.logger.Info().Msg("closing consumer")

	// First we close the channel manager
	// This also stops the consumer from receiving new messages
	err := r.chManager.Close()
	if err != nil {
		return err
	}

	// Now we'll block till all the active go routines are done
	r.wg.Wait()
	r.logger.Info().Msg("consumer closed gracefully")
	return nil
}
