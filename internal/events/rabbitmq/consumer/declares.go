package consumer

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
)

func (r *rabbitmqConsumer) declare(routingKeys []string) error {
	r.chManager.ChannelMux.RLock()
	defer r.chManager.ChannelMux.RUnlock()

	err := r.createDLX()
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to create DLX")
	}

	err = r.queueBindDeclare(routingKeys)
	if err != nil {
		return eris.Wrap(err, "failed to declare queue bind")
	}

	err = r.chManager.Channel.Qos(
		r.config.PrefetchCount, 0, false,
	)
	if err != nil {
		return eris.Wrap(err, "failed to set QoS")
	}

	return nil
}

func (r *rabbitmqConsumer) createDLX() error {
	dlxName := r.config.QueueName + "_dlx"
	err := r.deadLetterDeclare(dlxName)
	if err != nil {
		return eris.Wrap(err, "failed to declare dead letter")
	}

	err = r.queueDeclare(dlxName)
	if err != nil {
		return eris.Wrap(err, "failed to declare queue")
	}

	return nil
}

func (r *rabbitmqConsumer) queueDeclare(dlxName string) error {
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

	_, err = r.chManager.Channel.QueueDeclare(
		r.config.QueueName,
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-queue-type":           "quorum",
			"x-dead-letter-exchange": dlxName,
		},
	)
	if err != nil {
		return eris.Wrap(err, "failed to declare queue")
	}

	return nil
}

func (r *rabbitmqConsumer) queueBindDeclare(routingKeys []string) error {
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

func (r *rabbitmqConsumer) deadLetterDeclare(dlxName string) error {
	err := r.chManager.Channel.ExchangeDeclare(
		dlxName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return eris.Wrap(err, "failed to declare exchange")
	}

	_, err = r.chManager.Channel.QueueDeclare(
		dlxName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return eris.Wrap(err, "failed to declare queue")
	}

	err = r.chManager.Channel.QueueBind(
		dlxName,
		"",
		dlxName,
		false,
		nil,
	)
	if err != nil {
		return eris.Wrap(err, "failed to bind queue")
	}

	return nil
}
