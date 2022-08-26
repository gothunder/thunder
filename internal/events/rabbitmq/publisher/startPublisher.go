package publisher

func (r rabbitmqPublisher) startPublisher() {
	go func() {
		for {
			r.listenForNotifications()

			// Check if the channel reconnects
			err := <-r.chManager.NotifyReconnection
			if err != nil {
				// TODO handle error
				break
			}

			r.logger.Info().Msg("restarting publisher after reconnection")
		}
	}()
}
