package events

import "context"

type HandlerResponse int

const (
	// Default, we remove the message from the queue.
	Success HandlerResponse = iota

	// The message will be delivered to a server configured dead-letter queue.
	DeadLetter

	// Deliver this message to a different worker.
	Retry

	RetryBackoff
)

type EventDecoder interface {
	// Decode decodes the payload into the given interface.
	// Returns an error if the payload cannot be decoded.
	Decode(v interface{}) error
}

type Handler interface {
	// The topics that the consumer will be subscribed to.
	Topics() []string

	// The function that will be called when a subscribed message is received.
	Handle(ctx context.Context, topic string, decoder EventDecoder) HandlerResponse
}
