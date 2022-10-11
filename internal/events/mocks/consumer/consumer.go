package consumer

import (
	"github.com/gothunder/thunder/internal/events/mocks"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rs/zerolog"
)

type mockedConsumer struct {
	mockedChan chan mocks.MockedEvent
}

func NewConsumer(mockedChan chan mocks.MockedEvent, log *zerolog.Logger) (events.EventConsumer, error) {
	consumer := mockedConsumer{
		mockedChan: mockedChan,
	}

	return &consumer, nil
}
