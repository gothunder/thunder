package consumer

import (
	"context"
	"time"

	"github.com/gothunder/thunder/internal/events/rabbitmq/manager"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rotisserie/eris"
)

func (r *rabbitmqConsumer) Subscribe(
	ctx context.Context,
	handler events.Handler,
) error {
	if r.config.DisableConsumer {
		r.logger.Info().Msg("consumer is disabled, skipping subscription")
		// Block until context is cancelled
		<-ctx.Done()
		return nil
	}

	for {
		// Start the go routines that will consume messages
		err := r.startGoRoutinesWithRetry(ctx, handler)
		if err != nil {
			return eris.Wrap(err, "failed to start go routines")
		}

		// Check if the channel reconnects
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err = <-r.chManager.NotifyReconnection:
			if err != nil {
				return eris.Wrap(err, "failed to reconnect to the amqp channel")
			}
		}

		r.logger.Info().Msg("restarting consumer after reconnection")
	}
}

// startGoRoutinesWithRetry registers the consumer, retrying with the same
// backoff used for reconnections.
//
// The broker can refuse the registration right after a successful reconnection
// (a 503 while it is still settling). That used to kill the subscription on the
// first try, leaving the pod alive and ready but consuming nothing at all.
func (r *rabbitmqConsumer) startGoRoutinesWithRetry(ctx context.Context, handler events.Handler) error {
	exponentialBackOff := manager.NewExponentialBackOff(manager.ReconnectionBudget)

	for {
		// Nothing is added to the wait group until the registration succeeds,
		// so a retry cannot double count the consumer handlers.
		err := r.startGoRoutinesFunc(handler)
		if err == nil {
			return nil
		}

		interval := exponentialBackOff.NextBackOff()
		if interval == exponentialBackOff.Stop {
			return err
		}

		r.logger.Error().Err(err).Msgf("failed to register the consumer, retrying in %s", interval)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case reconnectionErr := <-r.chManager.NotifyReconnection:
			// The manager noticed the broken channel before we did and already
			// reconnected. We have to consume that notification here, otherwise
			// Subscribe would read it right after our retry succeeds and
			// register a second consumer on the very same channel.
			if reconnectionErr != nil {
				return eris.Wrap(reconnectionErr, "failed to reconnect to the amqp channel")
			}

			r.logger.Info().Msg("restarting consumer after reconnection")
			exponentialBackOff.Reset()
		case <-time.After(interval):
		}
	}
}

func (r *rabbitmqConsumer) startGoRoutines(handler events.Handler) error {
	// Declare exchange, queues, and bind them together
	err := r.declare(handler.Topics())
	if err != nil {
		return err
	}

	// A concurrent reconnection may have swapped the channel, so we read the
	// current one under the lock.
	r.chManager.ChannelMux.RLock()
	channel := r.chManager.Channel
	r.chManager.ChannelMux.RUnlock()

	msgs, err := channel.Consume(
		r.config.QueueName,
		r.config.ConsumerName,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return eris.Wrap(err, "failed to consume messages")
	}

	// We'll keep track of the go routines that we start
	r.wg.Add(r.config.ConsumerConcurrency)
	for i := 0; i < r.config.ConsumerConcurrency; i++ {
		go func() {
			r.handler(msgs, handler)
			// The handler will return when the channel is closed
			r.wg.Done()
		}()
	}
	r.logger.Info().Msgf("processing messages on %v goroutines", r.config.ConsumerConcurrency)

	return nil
}
