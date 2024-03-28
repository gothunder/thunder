package outboxpublisher

import (
	"context"
	"encoding/json"

	"github.com/TheRafaBonin/roxy"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	thunderContext "github.com/gothunder/thunder/pkg/context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Publish publishes a message to the given topic
// The message is published into the outbox table on the database
// There must be a publisher factory in the context in order to publish the message
// because it is expected that this is running inside a transaction, so the publisher
// is transactional
func (r *rabbitmqOutboxPublisher[T]) Publish(ctx context.Context, topic string, payload interface{}) (err error) {
	// Tracing instrumentation
	tctx, span := otel.Tracer(scope).Start(ctx, "rabbitmqPublisher.Publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystem("rabbitmq"),                   // This indicates the messaging system is rabbitmq
			semconv.MessagingRabbitmqDestinationRoutingKey(topic), // This indicates the routing key
			semconv.MessagingOperationPublish,                     // This indicates the operation is a publish
		),
	)
	defer func() {
		// If there is an error, record it and set the span status to error
		// so it can be seen on the tracing system
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to publish message")
		}
		span.End()
	}()

	// we need to extract the publisher factory from the context and then check if it is zerovalue
	// if it is, then we assume that it is not running inside a transaction and return an error
	publisherFactory := r.outboxPublisherFactoryCtxExtractor(tctx)
	var zeroValue T
	if publisherFactory == zeroValue {
		return roxy.New("Outbox publisher factory not found in context. Make sure it is running inside a transaction")
	}

	// Create the outbox publisher given the publisher factory
	// This publisher should be transactional and store messages in the outbox table
	outboxPublisher, err := publisherFactory.OutboxPublisher()
	if err != nil {
		return roxy.Wrap(err, "failed to create outbox publisher")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return roxy.Wrap(err, "failed to encode event")
	}

	// Ensure the correlation is propagated or generated
	tctx = thunderContext.ContextWithCorrelationID(tctx, thunderContext.CorrelationIDFromContext(tctx))
	metadata := thunderContext.MetadataFromContext(tctx)
	// We must generate a new message ID on every publish, because the current message ID in the
	// context is the one that was received from the client, and it's not unique.
	metadata.Set(thunderContext.ThunderIDMetadataKey, uuid.Must(uuid.NewV7()).String())

	// As we are using the watermill publisher interface, we need to create a watermill message
	msg := message.NewMessage(metadata.Get(thunderContext.ThunderIDMetadataKey), body)
	msg.Metadata = msgMetadataFromContextMetadata(metadata)
	msg.SetContext(tctx)

	// Publish the message with the trace context propagated in the message headers
	err = outboxPublisher.Publish(topic, r.tracePropagator.WithTrace(tctx, msg))
	if err != nil {
		return roxy.Wrap(err, "failed to publish message")
	}

	return nil
}

func msgMetadataFromContextMetadata(metadata *thunderContext.Metadata) message.Metadata {
	stringMap := metadata.MarshalMap()
	msgMetadata := make(message.Metadata, len(stringMap))
	for k, v := range stringMap {
		msgMetadata[k] = v
	}
	return msgMetadata
}
