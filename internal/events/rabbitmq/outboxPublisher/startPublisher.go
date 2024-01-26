package outboxpublisher

import "context"

func (r *rabbitmqOutboxPublisher[T]) StartPublisher(ctx context.Context) error {
	return r.msgForwarder.Run(ctx)
}
