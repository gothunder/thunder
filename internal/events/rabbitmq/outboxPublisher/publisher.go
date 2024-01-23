package outboxpublisher

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gothunder/thunder/internal/events/rabbitmq/tracing"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type rabbitmqOutboxPublisher struct {
	outPublisher message.Publisher
	msgForwarder *forwarder.Forwarder

	outboxPublisherFactoryCtxExtractor OutboxPublisherFactoryCtxExtractor

	// tracing
	tracer          trace.Tracer
	tracePropagator *tracing.WatermillTracePropagator
}

type ForwarderFactory interface {
	Forwarder(consumerGroup string, outPublisher message.Publisher) *forwarder.Forwarder
}

type OutboxPublisherFactory interface {
	OutboxPublisher() (message.Publisher, error)
}

type OutboxPublisherFactoryCtxExtractor func(ctx context.Context) OutboxPublisherFactory

func NewRabbitMQOutboxPublisher(
	logger *zerolog.Logger,
	forwarderFactory ForwarderFactory,
	outboxPublisherFactoryCtxExtractor OutboxPublisherFactoryCtxExtractor,
) (events.EventPublisher, error) {
	outPublisher, err := newRabbitMQOutPublisher(logger)
	if err != nil {
		return nil, eris.Wrap(err, "failed to create out publisher")
	}

	msgForwarder := forwarderFactory.Forwarder("default", outPublisher)

	return &rabbitmqOutboxPublisher{
		outPublisher: outPublisher,
		msgForwarder: msgForwarder,

		outboxPublisherFactoryCtxExtractor: outboxPublisherFactoryCtxExtractor,

		tracer:          otel.Tracer("thunder-message-publisher-tracer"),
		tracePropagator: tracing.NewWatermillTracePropagator(),
	}, nil
}
