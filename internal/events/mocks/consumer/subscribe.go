package consumer

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/gothunder/thunder/pkg/events"
)

func (m *mockedConsumer) Subscribe(
	ctx context.Context,
	handler events.Handler,
) error {
	for msg := range m.mockedChan {
		decoder := json.NewDecoder(bytes.NewReader(msg.Payload))

		go handler.Handle(ctx, msg.Topic, decoder)
	}

	return nil
}
