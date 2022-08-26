package consumer

import (
	"sync"

	"github.com/gothunder/thunder/internal/events/rabbitmq"
	"github.com/gothunder/thunder/internal/events/rabbitmq/manager"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type rabbitmqConsumer struct {
	// Customizable fields
	config rabbitmq.Config
	logger *zerolog.Logger

	// Connection manager
	chManager *manager.ChannelManager

	// Wait group used to wait for all the consumer handlers to finish
	wg *sync.WaitGroup
}

func NewConsumer(url string, config amqp.Config, log *zerolog.Logger) (rabbitmqConsumer, error) {
	chManager, err := manager.NewChannelManager(url, config, log)
	if err != nil {
		return rabbitmqConsumer{}, err
	}

	consumer := rabbitmqConsumer{
		config: rabbitmq.LoadConfig(log),
		logger: log,

		chManager: chManager,

		wg: &sync.WaitGroup{},
	}

	consumer.startConsumer([]string{"test"})
	return consumer, nil
}
