package consumer

import (
	"sync"

	"github.com/gothunder/thunder/internal/events/rabbitmq"
	"github.com/gothunder/thunder/internal/events/rabbitmq/manager"
	"github.com/gothunder/thunder/internal/events/rabbitmq/tracing"
	"github.com/gothunder/thunder/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type rabbitmqConsumer struct {
	// Customizable fields
	config rabbitmq.Config
	logger *zerolog.Logger

	// Connection manager
	chManager *manager.ChannelManager

	// Wait group used to wait for all the consumer handlers to finish
	wg *sync.WaitGroup

	// Wait group used to wait for all the backoff handlers to finish
	backoffWg *sync.WaitGroup

	// tracing
	tracer          trace.Tracer
	tracePropagator *tracing.AmqpTracePropagator
}

func NewConsumer(amqpConf amqp.Config, log *zerolog.Logger) (events.EventConsumer, error) {
	config := rabbitmq.LoadConfig(log)

	chManager, err := manager.NewChannelManager(config.URL, amqpConf, log)
	if err != nil {
		return &rabbitmqConsumer{}, err
	}

	consumer := rabbitmqConsumer{
		config: config,
		logger: log,

		chManager: chManager,

		wg:        &sync.WaitGroup{},
		backoffWg: &sync.WaitGroup{},

		tracer:          otel.Tracer("thunder-message-consumer-tracer"),
		tracePropagator: tracing.NewAmqpTracing(),
	}

	return &consumer, nil
}
