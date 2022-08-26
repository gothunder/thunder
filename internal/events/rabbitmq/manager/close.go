package manager

import "github.com/rotisserie/eris"

// Gracefully close the channel and connection
func (chManager *ChannelManager) Close() error {
	// Lock the channel manager mutex
	chManager.ChannelMux.Lock()
	defer chManager.ChannelMux.Unlock()

	err := chManager.Channel.Close()
	if err != nil {
		return eris.Wrap(err, "amqp closing channel")
	}

	err = chManager.Connection.Close()
	if err != nil {
		return eris.Wrap(err, "amqp closing connection")
	}

	return nil
}
