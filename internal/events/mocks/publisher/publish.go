package publisher

import (
	"context"
	"encoding/json"

	"github.com/gothunder/thunder/internal/events/mocks"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rotisserie/eris"
)

func (m *mockedPublisher) Publish(
	ctx context.Context,
	event events.Event,
) error {
	body, err := json.Marshal(event.Payload)
	if err != nil {
		return eris.Wrap(err, "failed to encode event")
	}

	m.mockedChan <- mocks.MockedEvent{
		Topic:   event.Topic,
		Payload: body,
	}
	return nil
}
