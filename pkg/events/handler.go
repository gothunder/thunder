package events

import "context"

type HandlerResponse int

const (
	// Default, we remove the message from the queue.
	Success HandlerResponse = iota

	// The message will be delivered to a server configured dead-letter queue.
	Requeue

	// Deliver this message to a different worker.
	Retry
)

// The function that will be called when a message is received.
type HandlerFunc func(context.Context, Event) HandlerResponse

type EventHandler struct {
	// The event that will be handled.
	// You can use the Topic field to filter the events.
	// If the Topic field is empty, this will be a catch-all handler.
	Event Event

	Handler HandlerFunc
}
