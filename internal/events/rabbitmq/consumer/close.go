package consumer

// Graceful shutdown of the consumer.
func (r *rabbitmqConsumer) Close() error {
	r.logger.Info().Msg("closing consumer")

	// We send a stop signal to all the handlers
	for i := 0; i < r.config.ConsumerConcurrency; i++ {
		r.stop <- true
	}

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
