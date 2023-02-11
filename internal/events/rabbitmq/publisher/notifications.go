package publisher

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *rabbitmqPublisher) listenForNotifications() {
	go r.handleNotifyFlow()
	go r.handleNotifyBlocked()
}

// If the publisher sends more messages than the queue can handle, the rabbitmq server will start throttling the publisher.
func (r *rabbitmqPublisher) handleNotifyFlow() {
	notifyFlowChan := r.chManager.Channel.NotifyFlow(
		make(chan bool, 1),
	)

	for flowMode := range notifyFlowChan {
		if flowMode {
			r.logger.Warn().Msg("publisher is sending too many messages, throttling the channel")
			continue
		}

		r.logger.Warn().Msg("stop throttling the channel")
	}
}

// If the rabbitmq server is blocked, the publisher will be blocked as well.
func (r *rabbitmqPublisher) handleNotifyBlocked() {
	notifyBlockedChan := r.chManager.Connection.NotifyBlocked(
		make(chan amqp.Blocking),
	)

	for blocking := range notifyBlockedChan {
		r.pausePublishMux.Lock()
		if blocking.Active {
			r.pausePublish = true
			if len(r.pauseSignalChan) == 0 {
				r.pauseSignalChan <- true
			}
			r.logger.Warn().Msg("pausing publishing due to TCP blocking from server")
		} else {
			r.pausePublish = false
			r.logger.Warn().Msg("resuming publishing due to TCP unblocking from server")
		}
		r.pausePublishMux.Unlock()
	}
}
