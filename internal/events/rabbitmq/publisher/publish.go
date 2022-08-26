package publisher

import (
	"context"

	"github.com/gothunder/thunder/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
)

func (r rabbitmqPublisher) Publish(ctx context.Context, event events.Event) error {
	r.pausePublishMux.RLock()
	defer r.pausePublishMux.RUnlock()

	if r.pausePublish {
		// TODO handle this
		return eris.New("publishing is paused")
	}

	message := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         []byte("{}"),
	}

	// Actual publish.
	err := r.chManager.Channel.Publish(
		r.config.ExchangeName,
		event.Topic,
		true,
		false,
		message,
	)
	if err != nil {
		return err
	}
	return nil
}
