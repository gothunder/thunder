package outbox

import (
	"errors"
)

var (
	ErrEmptyTopic   = errors.New("empty topic")
	ErrEmptyPayload = errors.New("empty payload")
)

type Message struct {
	Topic   string
	Payload []byte
	Headers map[string]string
}

func (m Message) BuildEntMessage(creator MessageCreator) MessageCreator {
	return creator.
		SetTopic(m.Topic).
		SetPayload(m.Payload).
		SetHeaders(m.Headers)
}

func (m Message) Validate() error {
	if m.Topic == "" {
		return ErrEmptyTopic
	}
	if m.Payload == nil || len(m.Payload) == 0 {
		return ErrEmptyPayload
	}
	return nil
}
