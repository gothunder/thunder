package publisher

import "context"

func (r *rabbitmqPublisher) StartPublisher(ctx context.Context) error {
	for {
		r.listenForNotifications()
		r.proccess()

		r.logger.Info().Msg("restarting publisher after reconnection")
	}
}
