package publisher

import (
	"context"

	"github.com/rotisserie/eris"
)

func (r *rabbitmqPublisher) StartPublisher(ctx context.Context) error {
	for {
		err := r.chManager.Channel.Confirm(false)
		if err != nil {
			return eris.Wrap(err, "failed to enable publisher confirms")
		}
		r.listenForNotifications()

		r.pausePublishMux.Lock()
		r.pausePublish = false
		r.pausePublishMux.Unlock()

		err = r.proccessingLoop()
		if err != nil {
			return err
		}

		r.logger.Info().Msg("restarting publisher after reconnection")
	}
}

func (r *rabbitmqPublisher) proccessingLoop() error {
	for {
		select {
		case err := <-r.chManager.NotifyReconnection:
			if err != nil {
				return eris.Wrap(err, "failed to reconnect to the amqp channel")
			}
			return nil
		case msg := <-r.unpublishedMessages:
			r.publisherFunc(msg)
		}
	}
}
