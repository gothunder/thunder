package consumer

import (
	"context"

	"github.com/gothunder/thunder/pkg/events"
)

func (m *mockedConsumer) Subscribe(
	ctx context.Context,
	topics []string,
	handler events.HandlerFunc,
) error {
	for msg := range m.mockedChan {
		handler(ctx, msg.Topic, msg.Payload)
	}

	return nil
}
