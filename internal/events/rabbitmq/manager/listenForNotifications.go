package manager

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
)

// Listen for any errors in the channel and attempt to reconnect
func (chManager *ChannelManager) listenForNotifications() {
	// Listen when a channel is closed or cancelled
	notifyCloseChan := chManager.Channel.NotifyClose(
		make(chan *amqp.Error, 1),
	)
	notifyCancelChan := chManager.Channel.NotifyCancel(
		make(chan string, 1),
	)

	select {
	case err := <-notifyCloseChan:
		if err != nil {
			// If we get an error, then it means that the channel closed due to an error
			// in this case, we will attempt to reconnect

			chManager.logger.Error().Err(err).Msg("amqp channel closed, reconnecting")
			chManager.NotifyReconnection <- chManager.reconnect()
		}
		if err == nil {
			// If there is no error, then the channel was closed programatically
			// most likely it's a graceful shutdown, and we will not attempt to reconnect

			chManager.logger.Info().Msg("amqp channel closed gracefully")
		}
	case cancelMsg := <-notifyCancelChan:
		// These occurs when the queue that a consumer is listening on was deleted or
		// got moved to another host

		chManager.logger.Error().Err(eris.New(cancelMsg)).Msg("amqp channel cancelled, reconnecting, reason")
		chManager.NotifyReconnection <- chManager.reconnect()
	}
}
