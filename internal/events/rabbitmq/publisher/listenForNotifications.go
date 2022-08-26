package publisher

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r rabbitmqPublisher) listenForNotifications() {
	r.pausePublishMux.Lock()
	r.pausePublish = false
	r.pausePublishMux.Unlock()

	go r.handleNotifyFlow()
	go r.handleNotifyBlocked()

	// TODO handle pulish and return notifications
	// returnAMQPCh := publisher.chManager.channel.NotifyReturn(make(chan amqp.Return, 1))
	// for ret := range returnAMQPCh {
	// 	publisher.notifyReturnChan <- Return{ret}
	// }
	// publisher.chManager.channel.Confirm(false)
	// go func() {
	// 	publishAMQPCh := publisher.chManager.channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	// 	for conf := range publishAMQPCh {
	// 		publisher.notifyPublishChan <- Confirmation{
	// 			Confirmation:      conf,
	// 			ReconnectionCount: int(publisher.chManager.reconnectionCount),
	// 		}
	// 	}
	// }()
}

func (r rabbitmqPublisher) handleNotifyFlow() {
	notifyFlowChan := r.chManager.Channel.NotifyFlow(
		make(chan bool, 1),
	)

	for pausePublish := range notifyFlowChan {
		r.pausePublishMux.Lock()
		if pausePublish {
			r.pausePublish = true
			r.logger.Warn().Msg("pausing publishing due to TCP flow from server")
		} else {
			r.pausePublish = false
			r.logger.Warn().Msg("resuming publishing due to TCP flow from server")
		}
		r.pausePublishMux.Unlock()
	}
}

func (r rabbitmqPublisher) handleNotifyBlocked() {
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
			r.logger.Warn().Msg("resuming publishing due to TCP blocking from server")
		}
		r.pausePublishMux.Unlock()
	}
}
