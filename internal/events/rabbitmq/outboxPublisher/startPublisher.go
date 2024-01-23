package outboxpublisher

import "context"

func (r *rabbitmqOutboxPublisher) StartPublisher(ctx context.Context) error {
	return r.msgForwarder.Run(ctx)
}
