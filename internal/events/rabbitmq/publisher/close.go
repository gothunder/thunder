package publisher

import (
	"context"

	"github.com/rotisserie/eris"
)

// Graceful shutdown of the publisher.
func (r *rabbitmqPublisher) Close(ctx context.Context) error {
	r.logger.Info().Msg("closing publisher")

	// Wait till all events are published.
	err := r.waitOrTimeout(ctx)
	if err != nil {
		return err
	}

	// Now we'll close the channel
	err = r.chManager.Close()
	if err != nil {
		return err
	}

	r.logger.Info().Msg("publisher closed gracefully")
	return nil
}

func (r *rabbitmqPublisher) waitOrTimeout(ctx context.Context) error {
	waitChannel := make(chan struct{})
	go func() {
		defer close(waitChannel)
		r.wg.Wait()
	}()

	select {
	case <-waitChannel:
		return nil
	case <-ctx.Done():
		r.logger.Info().Msg("timeout reached, dropping messages")
		r.publisherFunc = r.dropMessage
		return eris.New("timeout while waiting for events to be published")
	}
}
