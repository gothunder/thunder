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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
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
	ctx, span := otel.Tracer(scope).Start(ctx, "rabbitmqPublisher.Publish",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			semconv.MessagingSystem("rabbitmq"),
			semconv.MessagingRabbitmqDestinationRoutingKey(topic),
			semconv.MessagingOperationPublish,
		),
	)
	defer span.End()

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
		Message: amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
			Headers: amqp091.Table{
				metadata.ThunderIDMetadataKey:            uuid.NewString(),
				metadata.ThunderCorrelationIDMetadataKey: thunderContext.CorrelationIDFromContext(ctx),
			},
		},
	}

	return nil
}

func (r *rabbitmqPublisher) publishMessage(msg message) {
	// We'll timeout the publish after confirmTimeout seconds and consider as failed
	ctx, cancel := context.WithTimeout(context.Background(), confirmTimeout)
	_, span := otel.Tracer(scope).Start(msg.Context, "rabbitmqPublisher.publishMessage",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystem("rabbitmq"),
			semconv.MessagingRabbitmqDestinationRoutingKey(msg.Topic),
			semconv.MessagingOperationPublish,
		),
	)
	defer span.End()

	// Actual publish.
	deferredConfirmation, err := r.chManager.Channel.PublishWithDeferredConfirmWithContext(
		ctx,
		r.config.ExchangeName,
		msg.Topic,
		true,
		false,
		*r.tracePropagator.WithTrace(ctx, &msg.Message),
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to publish message")
		log.Ctx(msg.Context).Error().Err(err).Msg("failed to publish event, retrying")

		// If we failed to publish, it means that the connection is down.
		// So we can pause the publisher and re-publish the event.
		// The publisher will be unpaused when the connection is re-established.
		r.pause()

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
		span.RecordError(errors.New("failed to confirm publish"))
		span.SetStatus(codes.Error, "failed to confirm publish")
		log.Ctx(msg.Context).Error().Msg("failed to confirm publish, retrying")

		// If we didn't get confirmation, we need to re-publish the event.
		r.unpublishedMessages <- msg
		return
	}

	log.Ctx(msg.Context).Info().Str("topic", msg.Topic).Msg("message published")
	r.wg.Done()
	r.updatePublishedAt()
}

func (r *rabbitmqPublisher) updatePublishedAt() {
	r.lastPublishedAtMux.Lock()
	r.lastPublishedAt = time.Now()
	r.lastPublishedAtMux.Unlock()
}

func (r *rabbitmqPublisher) getLastPublishedAt() time.Time {
	r.lastPublishedAtMux.RLock()
	defer r.lastPublishedAtMux.RUnlock()
	return r.lastPublishedAt
}
