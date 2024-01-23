package outboxpublisher

import "context"

func (r *rabbitmqOutboxPublisher) Close(ctx context.Context) error {
	return r.msgForwarder.Close()
}
