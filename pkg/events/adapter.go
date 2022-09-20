package events

import "context"

type EventConsumer interface {
	Subscribe(
		ctx context.Context,
		topics []string,
		handler HandlerFunc,
	) error

	Close(context.Context) error
}

type EventPublisher interface {
	StartPublisher(context.Context) error

	Publish(
		context.Context,
		Event,
	)

	Close(context.Context) error
}
