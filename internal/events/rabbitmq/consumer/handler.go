package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	thunderContext "github.com/gothunder/thunder/pkg/context"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	topicHeaderKey = "x-thunder-topic"
)

func (r *rabbitmqConsumer) handler(msgs <-chan amqp.Delivery, handler events.Handler) {
	for msg := range msgs {
		// in case of requeue backoff, we want to make sure we have the correct topic
		topic := extractTopic(msg)
		// we always inject the correct topic so the requeue backoff can work
		injectTopic(&msg, topic)

		ctx := r.tracePropagator.ExtractTrace(context.Background(), &msg)
		ctx, span := otel.Tracer(scope).Start(ctx, "rabbitmqConsumer.handler",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				semconv.MessagingSystem("rabbitmq"),
				semconv.MessagingRabbitmqDestinationRoutingKey(topic),
				semconv.MessagingOperationProcess,
			),
		)

		logger := r.logger.With().Str("topic", topic).Ctx(ctx).Logger()
		ctx = logger.WithContext(ctx)
		ctx = thunderContext.ContextWithMetadata(ctx, metadataFromAmqpTable(msg.Headers))
		// ensures that the correlation ID is propagated or generated
		ctx = thunderContext.ContextWithCorrelationID(ctx, thunderContext.CorrelationIDFromContext(ctx))

		decoder := newDecoder(msg)
		res := r.handleWithRecoverer(ctx, handler, topic, decoder)

		switch res {
		case events.Success:
			// Message was successfully processed
			err := msg.Ack(false)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to ack message")
				logger.Error().Err(err).Ctx(ctx).Msg("failed to ack message")
			}
		case events.DeadLetter:
			// We should retry to process the message on a different worker
			err := msg.Nack(false, false)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to nack message")
				logger.Error().Err(err).Ctx(ctx).Msg("failed to requeue message")
			}
		case events.RetryBackoff:
			// We should send to a go routine that will requeue the message after a backoff time
			go r.retryBackoff(ctx, msg)
		default:
			// We should stop processing the message
			err := msg.Nack(false, true)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to nack message")
				logger.Error().Err(err).Ctx(ctx).Msg("failed to discard message")
			}
		}
		span.End()
	}
}

func (r *rabbitmqConsumer) handleWithRecoverer(ctx context.Context, handler events.Handler, topic string, decoder events.EventDecoder) (res events.HandlerResponse) {
	logger := log.Ctx(ctx).With().Stack().Logger()
	logger.Info().Msg("consuming message")

	// Catch panics
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = eris.New(fmt.Sprintf("%v", r))
			}

			err = eris.Wrap(err, "panic")
			logger.Error().Err(err).Msg("panic while consuming message")

			// If there's a panic, we should stop processing the message
			res = events.DeadLetter
		}
	}()

	return handler.Handle(ctx, topic, decoder)
}

// extractTopic extracts the topic from the message.
// It looks at the headers first, then the routing key.
func extractTopic(msg amqp091.Delivery) string {
	if headerTopic, ok := msg.Headers[topicHeaderKey]; ok {
		return headerTopic.(string)
	}

	return msg.RoutingKey
}

// injectTopic injects the topic into the message headers.
func injectTopic(msg *amqp091.Delivery, topic string) {
	if msg.Headers == nil {
		msg.Headers = make(amqp.Table)
	}

	msg.Headers[topicHeaderKey] = topic
}

// newDecoder creates a new decoder given the message.
// It looks at the content type to determine the decoder with fallback to json.
func newDecoder(msg amqp091.Delivery) events.EventDecoder {
	if msg.ContentType == "application/msgpack" {
		return msgpack.NewDecoder(bytes.NewReader(msg.Body))
	}

	return json.NewDecoder(bytes.NewReader(msg.Body))
}

func metadataFromAmqpTable(headers amqp.Table) *thunderContext.Metadata {
	metadata := thunderContext.NewMetadata()

	// add all string headers to the metadata
	stringMap := make(map[string]string, len(headers))
	for k, v := range headers {
		if s, ok := v.(string); ok {
			stringMap[k] = s
		}
	}

	// unmarshal the headers into the metadata
	// it will get only the ones with the "x-thunder-metadata-" prefix
	// meaning it belongs to the metadata
	metadata.UnmarshalMap(stringMap)

	return metadata
}
