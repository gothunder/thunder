package manager

import (
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

type ChannelManager struct {
	// Customizable configuration
	url        string
	amqpConfig amqp.Config
	logger     *zerolog.Logger

	// Exported fields
	Channel            *amqp.Channel
	Connection         *amqp.Connection
	ChannelMux         *sync.RWMutex
	NotifyReconnection chan error
}

func NewChannelManager(url string, conf amqp.Config, log *zerolog.Logger) (*ChannelManager, error) {
	// First we create a new channel and connection
	conn, ch, err := connect(url, conf)
	if err != nil {
		return nil, eris.Wrap(err, "getting the first channel")
	}

	chManager := ChannelManager{
		// Pass the arguments
		url:        url,
		amqpConfig: conf,
		logger:     log,

		// Pass the channel and connection we just created
		Connection: conn,
		Channel:    ch,
		ChannelMux: &sync.RWMutex{},
	}

	// Launch a gorouting to listen for any errors and reconnect if necessary
	go chManager.listenForNotifications()

	return &chManager, nil
}
