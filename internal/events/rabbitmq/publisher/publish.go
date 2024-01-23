package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	thunderContext "github.com/gothunder/thunder/pkg/context"
	"github.com/gothunder/thunder/pkg/events/metadata"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const confirmTimeout = 5 * time.Second

type message struct {
	Context context.Context
	Topic   string
	Message amqp091.Publishing
}

// Publish publishes a message to the given topic
// The message is published asynchronously
// The message will be republished if the connection is lost
func (r *rabbitmqPublisher) Publish(ctx context.Context, topic string, payload interface{}) error {
	ctx, span := r.tracer.Start(ctx, "rabbitmqPublisher.Publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystem("rabbitmq"),
			semconv.MessagingRabbitmqDestinationRoutingKey(topic),
			semconv.MessagingOperationPublish,
		),
	)

	// We want to keep track of the messages being published
	r.wg.Add(1)

	body, err := json.Marshal(payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to encode event")
		span.End()
		r.wg.Done()
		return eris.Wrap(err, "failed to encode event")
	}

	// Queue the message to be published
	r.unpublishedMessages <- message{
		Context: ctx,
		Topic:   topic,
		Message: *r.tracePropagator.WithTrace(ctx, &amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
			Headers: amqp091.Table{
				metadata.ThunderIDMetadataKey:            uuid.NewString(),
				metadata.ThunderCorrelationIDMetadataKey: thunderContext.CorrelationIDFromContext(ctx),
			},
		}),
	}

	return nil
}

func (r *rabbitmqPublisher) publishMessage(msg message) {
	// We'll timeout the publish after confirmTimeout seconds and consider as failed
	ctx, cancel := context.WithTimeout(context.Background(), confirmTimeout)
	span := trace.SpanFromContext(msg.Context)

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
		span.RecordError(err)
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
	confirmed := deferredConfirmation.Wait()
	cancel()
	if !confirmed {
		span.RecordError(errors.New("failed to confirm publish"))
		log.Ctx(msg.Context).Error().Msg("failed to confirm publish, retrying")

		// If we didn't get confirmation, we need to re-publish the event.
		r.unpublishedMessages <- msg
		return
	}
	defer span.End()

	log.Ctx(msg.Context).Info().Str("topic", msg.Topic).Msg("message published")
	r.wg.Done()
}
