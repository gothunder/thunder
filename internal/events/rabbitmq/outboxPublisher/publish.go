package outboxpublisher

import (
	"context"
	"encoding/json"

	"github.com/TheRafaBonin/roxy"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	thunderContext "github.com/gothunder/thunder/pkg/context"
	"github.com/gothunder/thunder/pkg/events/metadata"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func (r *rabbitmqOutboxPublisher[T]) Publish(ctx context.Context, topic string, payload interface{}) (err error) {
	tctx, span := otel.Tracer(scope).Start(ctx, "rabbitmqPublisher.Publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystem("rabbitmq"),
			semconv.MessagingRabbitmqDestinationRoutingKey(topic),
			semconv.MessagingOperationPublish,
		),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to publish message")
		}
		span.End()
	}()

	publisherFactory := r.outboxPublisherFactoryCtxExtractor(tctx)
	var zeroValue T
	if publisherFactory == zeroValue {
		return roxy.New("Outbox publisher factory not found in context. Make sure it is running inside a transaction")
	}

	outboxPublisher, err := publisherFactory.OutboxPublisher()
	if err != nil {
		return roxy.Wrap(err, "failed to create outbox publisher")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return roxy.Wrap(err, "failed to encode event")
	}
	msg := message.NewMessage(uuid.NewString(), body)
	msg.Metadata.Set(metadata.ThunderIDMetadataKey, msg.UUID)
	msg.Metadata.Set(metadata.ThunderCorrelationIDMetadataKey, thunderContext.CorrelationIDFromContext(tctx))
	msg.SetContext(tctx)

	err = outboxPublisher.Publish(topic, r.tracePropagator.WithTrace(tctx, msg))
	if err != nil {
		return roxy.Wrap(err, "failed to publish message")
	}

	return nil
}
