package publisher

import (
	"context"
	"encoding/json"

	"github.com/gothunder/thunder/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
)

// TODO properly implement this
func (r rabbitmqPublisher) Publish(ctx context.Context, event events.Event) error {
	r.pausePublishMux.RLock()
	defer r.pausePublishMux.RUnlock()

	if r.pausePublish {
		// TODO handle this
		return eris.New("publishing is paused")
	}

	body, err := json.Marshal(event.Payload)
	if err != nil {
		return eris.Wrap(err, "failed to marshal event")
	}

	message := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	}

	// Actual publish.
	err = r.chManager.Channel.Publish(
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

// TODO properly implement this
func (r rabbitmqPublisher) PublishInternally(ctx context.Context, event events.Event) error {
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
		"",
		r.config.QueueName,
		true,
		false,
		message,
	)
	if err != nil {
		return err
	}
	return nil
}
