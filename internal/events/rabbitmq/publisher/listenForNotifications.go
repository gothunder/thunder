package publisher

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *rabbitmqPublisher) listenForNotifications() {
	r.pausePublishMux.Lock()
	r.pausePublish = false
	r.pausePublishMux.Unlock()

	go r.handleNotifyFlow()
	go r.handleNotifyBlocked()
	go r.handleNotifyReturn()
	go r.handleNotifyConfirm()
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
			r.logger.Warn().Msg("pausing publishing due to TCP blocking from server")
		} else {
			r.pausePublish = false
			r.logger.Warn().Msg("resuming publishing due to TCP unblocking from server")
		}
		r.pausePublishMux.Unlock()
	}
}

// Check if a message was returned (failed to be published)
func (r *rabbitmqPublisher) handleNotifyReturn() {
	notifyReturnChan := r.chManager.Channel.NotifyReturn(
		make(chan amqp.Return),
	)

	for ret := range notifyReturnChan {
		r.logger.Error().Interface("return", ret).Msg("failed to publish event")
	}
}

// Check if the messages are being delivered
func (r *rabbitmqPublisher) handleNotifyConfirm() {
	err := r.chManager.Channel.Confirm(false)
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to enable publisher confirmations")
		return
	}

	notifyConfirmChan := r.chManager.Channel.NotifyPublish(
		make(chan amqp.Confirmation),
	)

	for conf := range notifyConfirmChan {
		if !conf.Ack {
			r.logger.Error().Interface("confirmation", conf).Msg("failed to publish event")
		}

		r.wg.Done()
	}
}
