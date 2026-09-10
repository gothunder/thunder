package publisher

import (
	"context"
	"time"

	"github.com/gothunder/thunder/internal/events/rabbitmq/manager"
	"github.com/rotisserie/eris"
)

func (r *rabbitmqPublisher) StartPublisher(ctx context.Context) error {
	go r.proccessingLoop()
	go r.healthCheckLoop()

	for {
		if r.chManager == nil {
			return eris.New("r.chManager is nil! Invalid publisher")
		}
		err := r.confirmWithRetry(ctx)
		if err != nil {
			return eris.Wrap(err, "failed to enable publisher confirms")
		}
		r.listenForNotifications()

		r.resume()

		// Wait for reconnection
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err = <-r.chManager.NotifyReconnection:
			if err != nil {
				return eris.Wrap(err, "failed to reconnect to the amqp channel")
			}
		}

		r.logger.Info().Msg("restarting publisher after reconnection")
	}
}

// confirmWithRetry puts the current channel in confirm mode, retrying with the
// same backoff used for reconnections.
//
// Like the consumer registration, this is the hop right after a reconnection:
// the broker can still refuse it, and failing on the first try used to stop the
// publisher for good.
func (r *rabbitmqPublisher) confirmWithRetry(ctx context.Context) error {
	exponentialBackOff := manager.NewExponentialBackOff(manager.ReconnectionBudget)

	for {
		err := r.confirm()
		if err == nil {
			return nil
		}

		interval := exponentialBackOff.NextBackOff()
		if interval == exponentialBackOff.Stop {
			return err
		}

		r.logger.Error().Err(err).Msgf("failed to enable publisher confirms, retrying in %s", interval)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case reconnectionErr := <-r.chManager.NotifyReconnection:
			// The manager reconnected while we were waiting. We consume the
			// notification here so the loop above doesn't read a stale one, and
			// retry straight away on the fresh channel. The budget is not reset,
			// so a permanent failure still gives up instead of idling forever.
			if reconnectionErr != nil {
				return eris.Wrap(reconnectionErr, "failed to reconnect to the amqp channel")
			}

			r.logger.Info().Msg("restarting publisher after reconnection")
		case <-time.After(interval):
		}
	}
}

// confirm reads the current channel under the manager lock, since a concurrent
// reconnection may have swapped it.
func (r *rabbitmqPublisher) confirm() error {
	r.chManager.ChannelMux.RLock()
	channel := r.chManager.Channel
	r.chManager.ChannelMux.RUnlock()

	return channel.Confirm(false)
}

func (r *rabbitmqPublisher) proccessingLoop() {
	for {
		select {
		// If we are reconnecting, we want to pause the publishing
		case <-r.pauseSignalChan:
			r.pausePublishMux.RLock()
			// If we are still reconnecting, we want to pause the publishing
			if r.pausePublish && len(r.pauseSignalChan) == 0 {
				r.pauseSignalChan <- true
			}
			r.pausePublishMux.RUnlock()

		// If we are not reconnecting, we want to publish the messages
		default:
			go r.publisherFunc(<-r.unpublishedMessages)
		}
	}
}
