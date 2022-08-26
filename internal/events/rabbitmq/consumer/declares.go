package consumer

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
)

func (r rabbitmqConsumer) declare(routingKeys []string) error {
	r.chManager.ChannelMux.RLock()
	defer r.chManager.ChannelMux.RUnlock()

	err := r.queueDeclare()
	if err != nil {
		return err
	}

	err = r.exchangeDeclare()
	if err != nil {
		return err
	}

	err = r.queueBindDeclare(routingKeys)
	if err != nil {
		return err
	}

	err = r.chManager.Channel.Qos(
		0, 0, false,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r rabbitmqConsumer) queueDeclare() error {
	_, err := r.chManager.Channel.QueueDeclare(
		r.config.QueueName,
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-queue-type": "quorum",
		},
	)
	if err != nil {
		return eris.Wrap(err, "failed to declare queue")
	}

	return nil
}

func (r rabbitmqConsumer) exchangeDeclare() error {
	err := r.chManager.Channel.ExchangeDeclare(
		r.config.ExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return eris.Wrap(err, "failed to declare exchange")
	}

	return nil
}

func (r rabbitmqConsumer) queueBindDeclare(routingKeys []string) error {
	for _, routingKey := range routingKeys {
		err := r.chManager.Channel.QueueBind(
			r.config.QueueName,
			routingKey,
			r.config.ExchangeName,
			false,
			nil,
		)
		if err != nil {
			return eris.Wrapf(err, "failed to bind queue, topic: %s", routingKey)
		}
	}

	return nil
}
