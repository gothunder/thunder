package manager

import (
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rotisserie/eris"
)

// Try to reconnect to amqp, if we fail to reconnect, then we will wait and try again
func (chManager *ChannelManager) reconnect() error {
	exponentialBackOff := backoff.NewExponentialBackOff()

	// We'll keep retrying until we get a successful connection
	for {
		// We want to increase the wait time each time we fail to reconnect
		interval := exponentialBackOff.NextBackOff()

		// After some time we want to give up and return an error
		if interval == exponentialBackOff.Stop {
			return eris.New("failed to reconnect to amqp")
		}

		chManager.logger.Info().Msgf("amqp channel reconnecting in %s", interval)
		time.Sleep(interval)

		err := chManager.reconnectAttempt()
		if err != nil {
			chManager.logger.Error().Err(err).Msg("amqp channel reconnection failed")
			continue
		}

		chManager.logger.Info().Msg("amqp channel reconnected")

		// Listen to the new channel for notifications
		go chManager.listenForNotifications()
		return nil
	}
}

// Close current channel and connection and reconnect to the server.
func (chManager *ChannelManager) reconnectAttempt() error {
	// Lock the channel manager mutex
	chManager.ChannelMux.Lock()
	defer chManager.ChannelMux.Unlock()

	// Get a new channel and connection
	newConn, newChannel, err := connect(chManager.url, chManager.amqpConfig)
	if err != nil {
		return err
	}

	// We got a new channel and connection, close the old ones
	err = chManager.Channel.Close()
	if err != nil {
		chManager.logger.Error().Err(err).Msg("amqp failed to close channel")
	}
	err = chManager.Connection.Close()
	if err != nil {
		chManager.logger.Error().Err(err).Msg("amqp failed to close connection")
	}

	chManager.Connection = newConn
	chManager.Channel = newChannel

	return nil
}
