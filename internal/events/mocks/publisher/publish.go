package publisher

import (
	"context"
	"encoding/json"

	"github.com/gothunder/thunder/internal/events/mocks"
	"github.com/rotisserie/eris"
)

func (m *mockedPublisher) Publish(
	ctx context.Context,
	topic string,
	payload interface{},
) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return eris.Wrap(err, "failed to encode event")
	}

	m.mockedChan <- mocks.MockedEvent{
		Topic:   topic,
		Payload: body,
	}
	return nil
}
