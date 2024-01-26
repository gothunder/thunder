package outbox

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyTopic   = errors.New("empty topic")
	ErrEmptyPayload = errors.New("empty payload")
)

type Message struct {
	// Fields to be used by the outbox
	ID          uuid.UUID
	CreatedAt   time.Time
	DeliveredAt time.Time

	// Fields to be used during the storage of the message
	Topic   string
	Payload []byte
	Headers map[string]string
}

func NewMessage(topic string, payload []byte, headers map[string]string) Message {
	if headers == nil {
		headers = make(map[string]string)
	}
	return Message{
		ID:      uuid.New(),
		Topic:   topic,
		Payload: payload,
		Headers: headers,
	}
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
