package publisher

import "github.com/rotisserie/eris"

func (r rabbitmqPublisher) StartPublisher() error {
	for {
		r.listenForNotifications()

		// Check if the channel reconnects
		err := <-r.chManager.NotifyReconnection
		if err != nil {
			return eris.Wrap(err, "failed to reconnect to the amqp channel")
		}

		r.logger.Info().Msg("restarting publisher after reconnection")
	}
}
