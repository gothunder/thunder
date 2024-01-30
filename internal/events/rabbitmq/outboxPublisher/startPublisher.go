package outboxpublisher

import "context"

func (r *rabbitmqOutboxPublisher[T]) StartPublisher(ctx context.Context) error {
	// Starts the message forwarder from the watermill pkg
	// This will forward messages from the outbox to the rabbitmq
	return r.msgForwarder.Run(ctx)
}
