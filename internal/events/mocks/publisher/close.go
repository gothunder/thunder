package publisher

import "context"

func (m *mockedPublisher) Close(ctx context.Context) error {
	// TODO close events channel

	return nil
}
