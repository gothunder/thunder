package publisher

import (
	"sync"

	"github.com/gothunder/thunder/internal/events/rabbitmq"
	"github.com/gothunder/thunder/internal/events/rabbitmq/manager"
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

func NewPublisher(url string, config amqp.Config, log *zerolog.Logger) (rabbitmqPublisher, error) {
	chManager, err := manager.NewChannelManager(url, config, log)
	if err != nil {
		return rabbitmqPublisher{}, err
	}

	publisher := rabbitmqPublisher{
		config: rabbitmq.LoadConfig(log),
		logger: log,

		chManager: chManager,

		pausePublish:    false,
		pausePublishMux: &sync.RWMutex{},

		wg: &sync.WaitGroup{},

		notifyReturnChan:  make(chan amqp.Return),
		notifyPublishChan: make(chan amqp.Confirmation),
	}

	publisher.startPublisher()
	return publisher, nil
}
