package consumer

import (
	"context"

	"github.com/rotisserie/eris"
)

// Graceful shutdown of the consumer.
func (r *rabbitmqConsumer) Close(ctx context.Context) error {
	r.logger.Info().Msg("closing consumer")

	// First we stop sending new messages to the consumer
	r.chManager.Channel.Cancel(r.config.ConsumerName, true)

	// We'll block till all the active go routines are done
	err := r.waitOrTimeout(ctx)
	if err != nil {
		return err
	}

	// Now we'll close the channel
	err = r.chManager.Close()
	if err != nil {
		return err
	}

	r.logger.Info().Msg("consumer closed gracefully")
	return nil
}

func (r *rabbitmqConsumer) waitOrTimeout(ctx context.Context) error {
	waitChannel := make(chan struct{})
	go func() {
		defer close(waitChannel)
		r.wg.Wait()
	}()

	select {
	case <-waitChannel:
		return nil
	case <-ctx.Done():
		return eris.New("timeout while waiting for events to be consumed")
	}
}
