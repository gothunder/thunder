package publisher

import (
	"github.com/rotisserie/eris"
)

func (r *rabbitmqPublisher) proccess() error {
	for {
		select {
		case err := <-r.chManager.NotifyReconnection:
			if err != nil {
				return eris.Wrap(err, "failed to reconnect to the amqp channel")
			}
			return nil
		case event := <-r.unpublishedEvents:
			go r.publishEvent(event)
		}
	}
}
