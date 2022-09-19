package manager

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
)

// This is an utility function that returns a new channel and connection.
func connect(url string, conf amqp.Config) (*amqp.Connection, *amqp.Channel, error) {
	// Create a new connection
	amqpConn, err := amqp.DialConfig(url, amqp.Config(conf))
	if err != nil {
		return nil, nil, eris.Wrap(err, "failed to connect to amqp server")
	}

	// Create a new channel
	ch, err := amqpConn.Channel()
	if err != nil {
		return nil, nil, eris.Wrap(err, "failed to create channel")
	}

	return amqpConn, ch, nil
}
