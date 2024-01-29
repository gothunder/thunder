package publisher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog/log"
)

const confirmTimeout = 10 * time.Second

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
	// We'll timeout the publish after confirmTimeout seconds and consider as failed
	ctx, cancel := context.WithTimeout(context.Background(), confirmTimeout)

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

		// If we failed to publish, it means that the connection is down.
		// So we can pause the publisher and re-publish the event.
		// The publisher will be unpaused when the connection is re-established.
		r.pausePublishMux.Lock()
		r.pausePublish = true
		r.pausePublishMux.Unlock()

		// If the channel is empty, we can send a signal to pause the publisher
		if len(r.pauseSignalChan) == 0 {
			r.pauseSignalChan <- true
		}

		// Re-publish the event
		r.unpublishedMessages <- msg
		cancel()
		return
	}

	// Wait for confirmation. Timeouts after confirmTimeout seconds.
	confirmed, err := deferredConfirmation.WaitContext(ctx)
	cancel()
	if err != nil {
		log.Ctx(msg.Context).Error().Err(err).Msg("error on confirming publish, retrying")

		// If we timed out, we need to re-publish the event. We don't pause publisher in this circunstances
		// because it may be a temporary issue with a leader node and the connection is still up
		r.unpublishedMessages <- msg
		return
	}
	if !confirmed {
		log.Ctx(msg.Context).Error().Msg("failed to confirm publish, retrying")

		// If we didn't get confirmation, we need to re-publish the event.
		r.unpublishedMessages <- msg
		return
	}

	log.Ctx(msg.Context).Info().Str("topic", msg.Topic).Msg("message published")
	r.wg.Done()
}
