package outboxpublisher

import "context"

func (r *rabbitmqOutboxPublisher[T]) Close(ctx context.Context) error {
	return r.msgForwarder.Close()
}
