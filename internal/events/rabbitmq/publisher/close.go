package publisher

import "context"

// Graceful shutdown of the publisher.
func (r *rabbitmqPublisher) Close(ctx context.Context) error {
	r.logger.Info().Msg("closing publisher")

	// Wait till all events are published.
	r.wg.Wait()

	// Now we'll close the channel
	err := r.chManager.Close()
	if err != nil {
		return err
	}

	r.logger.Info().Msg("publisher closed gracefully")
	return nil
}
