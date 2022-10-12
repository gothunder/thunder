package events

import "context"

type EventConsumer interface {
	// Subscribe subscribes to the given topics
	// The handler will be called when a message is received
	// The handler will be called concurrently
	// Returns an error if the subscription fails to start or reconnect
	Subscribe(
		ctx context.Context,
		topics []string,
		handler Handler,
	) error

	// Close gracefully closes the consumer, making sure all messages are processed
	Close(context.Context) error
}

type EventPublisher interface {
	// StartPublisher starts the background go routine that will publish messages
	// Returns an error if the publisher fails to start or reconnect
	StartPublisher(context.Context) error

	// Publish publishes a message to the given topic
	// The message is published asynchronously
	// The message will be republished if the connection is lost
	Publish(
		ctx context.Context,
		// The name of the event.
		topic string,
		// The payload of the event.
		payload interface{},
	) error

	// Close gracefully closes the publisher, making sure all messages are published
	Close(context.Context) error
}
