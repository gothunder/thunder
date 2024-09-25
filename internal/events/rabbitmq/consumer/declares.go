package consumer

import (
	"math"
	"strings"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rotisserie/eris"
)

func (r *rabbitmqConsumer) declare(routingKeys []string) error {
	err := r.declareAttempt(routingKeys, true)
	if err == nil {
		// If there is no error, then the declaration was successful
		r.logger.Info().Msg("successfully declared queues with DLX")
		return nil
	}

	err = r.closeChannelAndWaitForReconnect()
	if err != nil {
		return eris.Wrap(err, "closeChannelAndWaitForReconnect")
	}

	err = r.declareAttempt(routingKeys, false)
	if err != nil {
		return eris.Wrap(err, "declareAttempt - second attempt")
	}

	r.logger.Info().Msg("successfully declared queues without DLX")
	return nil
}

func (r *rabbitmqConsumer) closeChannelAndWaitForReconnect() error {
	// Wait for the channel to close - this is a workaround for the channel not closing immediately
	// Check if the channel is open with backoff

	r.chManager.Channel.Close()
	time.Sleep(2 * time.Second)

	const maxAttempts = 5
	const backoffFactor = 2
	for i := 1; i <= maxAttempts; i++ {
		if r.chManager.Channel.IsClosed() {
			x := math.Pow(backoffFactor, float64(i))
			time.Sleep(time.Duration(x) * time.Second)

			i++
		}
	}
	if r.chManager.Channel.IsClosed() {
		return eris.New("failed to reconnect to the amqp channel in time for subscriber")
	}

	return nil
}

func (r *rabbitmqConsumer) declareAttempt(routingKeys []string, useDLX bool) error {
	r.chManager.ChannelMux.RLock()
	defer r.chManager.ChannelMux.RUnlock()

	if useDLX {
		err := r.declareQueuesWithDLX()
		if err != nil {
			return eris.Wrap(err, "declareQueuesWithDLX")
		}
	} else {
		err := r.declareQueuesWithoutDLX()
		if err != nil {
			return eris.Wrap(err, "declareQueuesWithoutDLX")
		}
	}

	err := r.queueBindDeclare(routingKeys)
	if err != nil {
		return eris.Wrap(err, "queueBindDeclare")
	}

	err = r.chManager.Channel.Qos(
		r.config.PrefetchCount, 0, false,
	)
	if err != nil {
		return eris.Wrap(err, "failed to set QoS")
	}

	return nil
}

func (r *rabbitmqConsumer) declareQueuesWithDLX() error {
	dlxName := r.config.QueueName + "_dlx"
	err := r.deadLetterDeclare(dlxName)
	if err != nil {
		return eris.Wrap(err, "deadLetterDeclare")
	}

	err = r.queueDeclare(&dlxName)
	if err != nil {
		return eris.Wrap(err, "queueDeclare")
	}

	return nil
}

func (r *rabbitmqConsumer) declareQueuesWithoutDLX() error {
	err := r.queueDeclare(nil)
	if err != nil {
		return eris.Wrap(err, "queueDeclare")
	}

	return nil
}

func (r *rabbitmqConsumer) queueDeclare(dlxName *string) error {
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

	args := amqp091.Table{
		amqp091.QueueTypeArg: amqp091.QueueTypeQuorum,
	}
	if dlxName != nil && *dlxName != "" {
		args["x-dead-letter-exchange"] = dlxName
	}

	_, err = r.chManager.Channel.QueueDeclare(
		r.config.QueueName,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		if strings.Contains(err.Error(), "PRECONDITION_FAILED") {
			// Queue already exists - no need to redeclare
			// RabbitMQ will force close, we want to wait for the channel to open again
			err = r.closeChannelAndWaitForReconnect()
			if err != nil {
				return eris.Wrap(err, "closeChannelAndWaitForReconnect")
			}

			return nil
		}
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
		return eris.Wrap(err, "failed to declare DLX exchange")
	}

	_, err = r.chManager.Channel.QueueDeclare(
		dlxName,
		true,
		false,
		false,
		false,
		amqp091.Table{
			amqp091.QueueMessageTTLArg: 1000 * 60 * 60 * 24 * 14, // 14 day
			amqp091.QueueMaxLenArg:     10000,                    // 10k messages
		},
	)
	if err != nil {
		return eris.Wrap(err, "failed to declare DLX queue")
	}

	err = r.chManager.Channel.QueueBind(
		dlxName,
		"",
		dlxName,
		false,
		nil,
	)
	if err != nil {
		return eris.Wrap(err, "failed to bind DLX queue")
	}

	return nil
}
