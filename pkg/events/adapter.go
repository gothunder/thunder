package events

import "context"

type EventConsumer interface {
	Subscribe(
		ctx context.Context,
		topics []string,
		handler HandlerFunc,
	) error

	Close() error
}

type EventPublisher interface {
	StartPublisher() error

	Publish(
		context.Context,
		Event,
	)

	Close() error
}
