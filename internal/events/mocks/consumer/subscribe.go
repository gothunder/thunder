package consumer

import (
	"context"

	"github.com/gothunder/thunder/pkg/events"
)

func (m *mockedConsumer) Subscribe(
	ctx context.Context,
	topics []string,
	handler events.Handler,
) error {
	for msg := range m.mockedChan {
		go handler.Handle(ctx, msg.Topic, msg.Payload)
	}

	return nil
}
