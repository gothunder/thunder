package consumer

import "context"

func (m *mockedConsumer) Close(ctx context.Context) error {
	// TODO close events channel

	return nil
}
