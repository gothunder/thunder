package outboxpublisher

import "context"

// Close closes the message forwarder, which will stop forwarding messages from the outbox to the rabbitmq
func (r *rabbitmqOutboxPublisher[T]) Close(ctx context.Context) error {
	return r.msgForwarder.Close()
}
