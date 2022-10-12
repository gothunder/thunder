package publisher

import (
	"context"
	"encoding/json"
	"time"

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
func (r *rabbitmqPublisher) Publish(ctx context.Context, topic string, payload interface{}) error {
	// We want to keep track of the messages being published
	r.wg.Add(1)

	body, err := json.Marshal(payload)
	if err != nil {
		r.wg.Done()
		return eris.Wrap(err, "failed to encode event")
	}

	// Queue the message to be published
	r.unpublishedMessages <- message{
		Context: ctx,
		Topic:   topic,
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

	// We'll timeout the publish after 5 seconds and consider as failed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Actual publish.
	deferredConfirmation, err := r.chManager.Channel.PublishWithDeferredConfirmWithContext(
		ctx,
		r.config.ExchangeName,
		msg.Topic,
		true,
		false,
		msg.Message,
	)
	if err != nil {
		log.Ctx(msg.Context).Error().Err(err).Msg("failed to publish event, retrying")

		// If the channel is closed, we need to reconnect and re-publish the event.
		r.pausePublishMux.Lock()
		r.pausePublish = true
		r.pausePublishMux.Unlock()

		r.unpublishedMessages <- msg
		cancel()
		return
	}

	// Wait for confirmation.
	confirmed := deferredConfirmation.Wait()
	cancel()
	if !confirmed {
		log.Ctx(msg.Context).Error().Msg("failed to confirm publish, retrying")

		// If we didn't get confirmation, we need to re-publish the event.
		r.unpublishedMessages <- msg
		return
	}

	log.Ctx(msg.Context).Info().Str("topic", msg.Topic).Msg("message published")
	r.wg.Done()
}
