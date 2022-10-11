package publisher

import (
	"github.com/gothunder/thunder/internal/events/mocks"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rs/zerolog"
)

type mockedPublisher struct {
	mockedChan chan mocks.MockedEvent
}

func NewPublisher(mockedChan chan mocks.MockedEvent, log *zerolog.Logger) (events.EventPublisher, error) {
	publisher := mockedPublisher{
		mockedChan: mockedChan,
	}

	return &publisher, nil
}
