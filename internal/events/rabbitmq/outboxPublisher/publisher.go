package outboxpublisher

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gothunder/thunder/internal/events/rabbitmq/tracing"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

type rabbitmqOutboxPublisher[T OutboxPublisherFactory] struct {
	outPublisher message.Publisher
	msgForwarder *forwarder.Forwarder

	outboxPublisherFactoryCtxExtractor OutboxPublisherFactoryCtxExtractor[T]

	// tracing
	tracePropagator *tracing.WatermillTracePropagator
}

const scope = "github.com/gothunder/thunder/internal/events/rabbitmq/outboxPublisher"

type ForwarderFactory interface {
	Forwarder(consumerGroup string, outPublisher message.Publisher) (*forwarder.Forwarder, error)
}

type OutboxPublisherFactory interface {
	comparable
	OutboxPublisher() (message.Publisher, error)
}

type OutboxPublisherFactoryCtxExtractor[T OutboxPublisherFactory] func(ctx context.Context) T

func NewRabbitMQOutboxPublisher[T OutboxPublisherFactory](
	logger *zerolog.Logger,
	forwarderFactory ForwarderFactory,
	outboxPublisherFactoryCtxExtractor OutboxPublisherFactoryCtxExtractor[T],
) (events.EventPublisher, error) {
	outPublisher, err := newRabbitMQOutPublisher(logger)
	if err != nil {
		return nil, eris.Wrap(err, "failed to create out publisher")
	}

	msgForwarder, err := forwarderFactory.Forwarder("default", outPublisher)
	if err != nil {
		return nil, eris.Wrap(err, "failed to create forwarder")
	}

	return &rabbitmqOutboxPublisher[T]{
		outPublisher: outPublisher,
		msgForwarder: msgForwarder,

		outboxPublisherFactoryCtxExtractor: outboxPublisherFactoryCtxExtractor,

		tracePropagator: tracing.NewWatermillTracePropagator(),
	}, nil
}
