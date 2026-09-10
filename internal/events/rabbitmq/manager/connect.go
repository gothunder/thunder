package manager

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

// A dialer returns a new channel and connection.
type dialer func() (*amqp.Connection, *amqp.Channel, error)

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

// connectWithRetry dials until the budget is spent, then returns the last error.
//
// The first dial is the only one with no reconnection loop behind it, so a
// broker that is merely restarting used to take the whole service down. The
// budget is deliberately short: this runs inside an fx constructor, before the
// process listens on any port, so a long one would hide the outage as a hang.
func connectWithRetry(dial dialer, log *zerolog.Logger, budget time.Duration) (*amqp.Connection, *amqp.Channel, error) {
	exponentialBackOff := NewExponentialBackOff(budget)

	for {
		conn, ch, err := dial()
		if err == nil {
			return conn, ch, nil
		}

		interval := exponentialBackOff.NextBackOff()
		if interval == exponentialBackOff.Stop {
			return nil, nil, err
		}

		log.Error().Err(err).Msgf("amqp first connection failed, retrying in %s", interval)
		time.Sleep(interval)
	}
}
