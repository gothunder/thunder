package publisher

import (
	"sync"

	"github.com/gothunder/thunder/internal/events/rabbitmq"
	"github.com/gothunder/thunder/internal/events/rabbitmq/manager"
	"github.com/gothunder/thunder/internal/events/rabbitmq/tracing"
	"github.com/gothunder/thunder/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const scope = "github.com/gothunder/thunder/internal/events/rabbitmq/publisher"

type rabbitmqPublisher struct {
	// Customizable fields
	config rabbitmq.Config
	logger *zerolog.Logger

	// Connection manager
	chManager *manager.ChannelManager

	// Channel for publishing events
	unpublishedMessages chan message

	// Function that publishes the message
	publisherFunc func(message)

	// Wait group used to wait for all the publishes to finish
	wg *sync.WaitGroup

	// These flags are used to prevent the publisher from publishing messages to the queue
	pausePublish    bool
	pausePublishMux *sync.RWMutex
	pauseSignalChan chan bool

	// These fields are used to keep track of the publisher's state
	notifyReturnChan  chan amqp.Return
	notifyPublishChan chan amqp.Confirmation

	// tracing
	tracePropagator *tracing.AmqpTracePropagator
}

func NewPublisher(amqpConf amqp.Config, log *zerolog.Logger) (events.EventPublisher, error) {
	config := rabbitmq.LoadConfig(log)

	chManager, err := manager.NewChannelManager(config.URL, amqpConf, log)
	if err != nil {
		return &rabbitmqPublisher{}, err
	}

	publisher := rabbitmqPublisher{
		config: config,
		logger: log,

		chManager: chManager,

		unpublishedMessages: make(chan message),
		wg:                  &sync.WaitGroup{},

		pausePublish:    true,
		pausePublishMux: &sync.RWMutex{},

		// The buffer size of 100 is arbitrary
		// It is used to prevent the pause signal from hanging
		// Realistically, the pause signal should be processed immediately
		// But sometimes, a race condition can occur specially when having high throughput
		pauseSignalChan: make(chan bool, 100),

		notifyReturnChan:  make(chan amqp.Return),
		notifyPublishChan: make(chan amqp.Confirmation),

		tracePropagator: tracing.NewAmqpTracing(log),
	}
	publisher.publisherFunc = publisher.publishMessage

	return &publisher, nil
}
