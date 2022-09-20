package publisher

import (
	"context"
	"encoding/json"

	"github.com/gothunder/thunder/pkg/events"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog/log"
)

type message struct {
	Context context.Context
	Topic   string
	Message amqp091.Publishing
}

// Publish publishes a message to the given topic
// The message is published asynchronously
// The message will be republished if the connection is lost
func (r *rabbitmqPublisher) Publish(ctx context.Context, event events.Event) error {
	// We want to keep track of the messages being published
	r.wg.Add(1)

	body, err := json.Marshal(event.Payload)
	if err != nil {
		r.wg.Done()
		return eris.Wrap(err, "failed to encode event")
	}

	// Queue the message to be published
	r.unpublishedMessages <- message{
		Context: ctx,
		Topic:   event.Topic,
		Message: amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	}

	return nil
}

func (r *rabbitmqPublisher) publishMessage(msg message) {
	// Check if the publisher is paused
	r.pausePublishMux.RLock()
	if r.pausePublish {
		r.pausePublishMux.RUnlock()
		r.unpublishedMessages <- msg
		return
	}
	r.pausePublishMux.RUnlock()

	// Actual publish.
	deferredConfirmation, err := r.chManager.Channel.PublishWithDeferredConfirmWithContext(
		msg.Context,
		r.config.ExchangeName,
		msg.Topic,
		true,
		false,
		msg.Message,
	)
	if err != nil {
		// If the channel is closed, we need to reconnect and re-publish the event.
		r.pausePublishMux.Lock()
		r.pausePublish = true
		r.pausePublishMux.Unlock()

		r.unpublishedMessages <- msg
		return
	}

	// Wait for confirmation.
	confirmed := deferredConfirmation.Wait()
	if !confirmed {
		// If we didn't get confirmation, we need to re-publish the event.
		r.unpublishedMessages <- msg
		return
	}

	log.Ctx(msg.Context).Info().Str("topic", msg.Topic).Msg("message published")
	r.wg.Done()
}
