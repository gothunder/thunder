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
)

// The function that will be called when a message is received.
type Handler interface {
	Handle(ctx context.Context, topic string, payload []byte) HandlerResponse
}
