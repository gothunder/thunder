package publisher

import (
	"context"

	"github.com/rotisserie/eris"
)

func (r *rabbitmqPublisher) StartPublisher(ctx context.Context) error {
	go r.proccessingLoop()

	for {
		err := r.chManager.Channel.Confirm(false)
		if err != nil {
			return eris.Wrap(err, "failed to enable publisher confirms")
		}
		r.listenForNotifications()

		r.pausePublishMux.Lock()
		r.pausePublish = false
		r.pausePublishMux.Unlock()

		// Wait for reconnection
		err = <-r.chManager.NotifyReconnection
		if err != nil {
			return eris.Wrap(err, "failed to reconnect to the amqp channel")
		}

		r.logger.Info().Msg("restarting publisher after reconnection")
	}
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
