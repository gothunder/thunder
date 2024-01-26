package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	thunderContext "github.com/gothunder/thunder/pkg/context"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/gothunder/thunder/pkg/events/metadata"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

func (r *rabbitmqConsumer) handler(msgs <-chan amqp.Delivery, handler events.Handler) {
	for msg := range msgs {
		ctx := r.tracePropagator.ExtractTrace(context.Background(), &msg)
		ctx, span := otel.Tracer(scope).Start(ctx, "rabbitmqConsumer.handler",
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				semconv.MessagingSystem("rabbitmq"),
				semconv.MessagingRabbitmqDestinationRoutingKey(msg.RoutingKey),
				semconv.MessagingOperationProcess,
			),
		)

		logger := r.logger.With().
			Str("topic", msg.RoutingKey).Logger()
		ctx = logger.WithContext(ctx)
		ctx = ctxWithMsgID(ctx, msg)
		ctx = ctxWithCorrID(ctx, msg)

		var decoder events.EventDecoder
		if msg.ContentType == "application/msgpack" {
			decoder = msgpack.NewDecoder(bytes.NewReader(msg.Body))
		} else {
			decoder = json.NewDecoder(bytes.NewReader(msg.Body))
		}

		res := r.handleWithRecoverer(ctx, handler, msg.RoutingKey, decoder)

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
			msgcopy := msg
			go r.retryBackoff(ctx, msgcopy)
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

func idFromMessage(msg amqp.Delivery) string {
	if msgID, ok := msg.Headers[metadata.ThunderIDMetadataKey]; ok && msgID != nil {
		if msgIDStr, ok := msgID.(string); ok {
			return msgIDStr
		}
	}

	return ""
}

func ctxWithMsgID(ctx context.Context, msg amqp.Delivery) context.Context {
	msgID := idFromMessage(msg)
	return thunderContext.ContextWithMessageID(ctx, msgID)
}

func corrIdFromMessage(msg amqp.Delivery) string {
	if corrID, ok := msg.Headers[metadata.ThunderCorrelationIDMetadataKey]; ok && corrID != nil {
		if corrIDStr, ok := corrID.(string); ok {
			return corrIDStr
		}
	}

	return ""
}

func ctxWithCorrID(ctx context.Context, msg amqp.Delivery) context.Context {
	corrID := corrIdFromMessage(msg)
	return thunderContext.ContextWithCorrelationID(ctx, corrID)
}
