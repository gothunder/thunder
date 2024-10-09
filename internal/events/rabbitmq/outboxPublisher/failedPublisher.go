package outboxpublisher

import (
	"github.com/TheRafaBonin/roxy"
	"github.com/ThreeDotsLabs/watermill/message"
)

// Dummy publisher that always returns error
type FailedPublisher struct {}

func (f FailedPublisher) Publish(topic string, messages ...*message.Message) error {
	return roxy.New("failed to connect to broker on startup, can't publish messages")
}


func (f FailedPublisher) Close() error {
	return roxy.New("failed to connect to broker on startup, can't close publisher")
}
