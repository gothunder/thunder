package consumer

import "fmt"

// Graceful shutdown of the consumer.
func (r *rabbitmqConsumer) Close() error {
	r.logger.Info().Msg("closing consumer")

	// TODO stop consuming messages first
	// We'll block till all the active go routines are done
	fmt.Printf("waiting on waitgroup\n")
	r.wg.Wait()

	// Now we'll close the channel
	err := r.chManager.Close()
	if err != nil {
		return err
	}

	r.logger.Info().Msg("consumer closed gracefully")
	return nil
}
