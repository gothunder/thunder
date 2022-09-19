package publisher

import (
	"context"
	"encoding/json"

	"github.com/gothunder/thunder/pkg/events"
	"github.com/rabbitmq/amqp091-go"
)

func (r *rabbitmqPublisher) Publish(ctx context.Context, event events.Event) {
	r.wg.Add(1)
	r.unpublishedEvents <- event
}

func (r *rabbitmqPublisher) publishEvent(event events.Event) {
	r.pausePublishMux.RLock()
	if r.pausePublish {
		r.unpublishedEvents <- event
		r.pausePublishMux.RUnlock()
		return
	}
	r.pausePublishMux.RUnlock()

	body, err := json.Marshal(event.Payload)
	if err != nil {
		r.logger.Error().Interface("event", event).Err(err).Msg("failed to encode event")
		r.wg.Done()
		return
	}

	message := amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
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
		r.unpublishedEvents <- event
	}
}
