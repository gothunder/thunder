package consumer

import (
	"sync"

	"github.com/gothunder/thunder/internal/events/rabbitmq"
	"github.com/gothunder/thunder/internal/events/rabbitmq/manager"
	"github.com/gothunder/thunder/internal/events/rabbitmq/tracing"
	"github.com/gothunder/thunder/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const scope = "github.com/gothunder/thunder/internal/events/rabbitmq/consumer"

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
	tracePropagator *tracing.AmqpTracePropagator
}

func NewConsumer(amqpConf amqp.Config, log *zerolog.Logger, opts ...rabbitmq.RabbitmqConfigOption) (events.EventConsumer, error) {
	config := rabbitmq.LoadConfig(log, opts...)

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

		tracePropagator: tracing.NewAmqpTracing(log),
	}

	return &consumer, nil
}
