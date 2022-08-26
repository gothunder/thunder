package events

import "context"

type EventConsumer interface {
	Subscribe(
		context.Context,
		[]EventHandler,
	) error

	Close() error
}

type EventPublisher interface {
	StartPublisher() error

	Publish(
		context.Context,
		Event,
	) error

	PublishInternally(
		context.Context,
		Event,
	) error

	Close() error
}
