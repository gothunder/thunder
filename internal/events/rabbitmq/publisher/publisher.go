package publisher

import (
	"sync"

	"github.com/gothunder/thunder/internal/events/rabbitmq"
	"github.com/gothunder/thunder/internal/events/rabbitmq/manager"
	"github.com/gothunder/thunder/pkg/events"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type rabbitmqPublisher struct {
	// Customizable fields
	config rabbitmq.Config
	logger *zerolog.Logger

	// Connection manager
	chManager *manager.ChannelManager

	// These flags are used to prevent the publisher from publishing messages to the queue
	pausePublish    bool
	pausePublishMux *sync.RWMutex

	// Wait group used to wait for all the publishes to finish
	wg *sync.WaitGroup

	// These fields are used to keep track of the publisher's state
	notifyReturnChan  chan amqp.Return
	notifyPublishChan chan amqp.Confirmation
}

func NewPublisher(amqpConf amqp.Config, log *zerolog.Logger) (events.EventPublisher, error) {
	config := rabbitmq.LoadConfig(log)

	chManager, err := manager.NewChannelManager(config.URL, amqpConf, log)
	if err != nil {
		return rabbitmqPublisher{}, err
	}

	publisher := rabbitmqPublisher{
		config: config,
		logger: log,

		chManager: chManager,

		pausePublish:    false,
		pausePublishMux: &sync.RWMutex{},

		wg: &sync.WaitGroup{},

		notifyReturnChan:  make(chan amqp.Return),
		notifyPublishChan: make(chan amqp.Confirmation),
	}

	return publisher, nil
}
