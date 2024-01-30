package outboxpublisher

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/rotisserie/eris"
	amqpclient "github.com/streadway/amqp"
)

// ThunderTopologyBuilder is a rabbitmq topology builder addapt thunder constraints
// on top of watermill topology builder. It is used to build the topology of the
// rabbitmq publisher and consumer
type ThunderTopologyBuilder struct{}

// ExchangeDeclare declares an exchange on the rabbitmq broker
func (builder ThunderTopologyBuilder) ExchangeDeclare(channel *amqpclient.Channel, exchangeName string, config amqp.Config) error {
	return channel.ExchangeDeclare(
		exchangeName,
		config.Exchange.Type,
		config.Exchange.Durable,
		config.Exchange.AutoDeleted,
		config.Exchange.Internal,
		config.Exchange.NoWait,
		config.Exchange.Arguments,
	)
}

// BuildTopology builds the topology of the rabbitmq publisher and consumer
// It declares a dead letter queue
func (builder *ThunderTopologyBuilder) BuildTopology(channel *amqpclient.Channel, queueName string, exchangeName string, config amqp.Config, logger watermill.LoggerAdapter) error {
	if err := builder.deadLetterDeclare(channel, queueName+"_dlx"); err != nil {
		return eris.Wrap(err, "failed to declare dead letter")
	}

	if _, err := channel.QueueDeclare(
		queueName,
		config.Queue.Durable,
		config.Queue.AutoDelete,
		config.Queue.Exclusive,
		config.Queue.NoWait,
		config.Queue.Arguments,
	); err != nil {
		return eris.Wrap(err, "cannot declare queue")
	}

	logger.Debug("Queue declared", nil)

	if exchangeName == "" {
		logger.Debug("No exchange to declare", nil)
		return nil
	}
	if err := builder.ExchangeDeclare(channel, exchangeName, config); err != nil {
		return eris.Wrap(err, "cannot declare exchange")
	}

	logger.Debug("Exchange declared", nil)

	if err := channel.QueueBind(
		queueName,
		config.QueueBind.GenerateRoutingKey(queueName),
		exchangeName,
		config.QueueBind.NoWait,
		config.QueueBind.Arguments,
	); err != nil {
		return eris.Wrap(err, "cannot bind queue")
	}
	return nil
}

// deadLetterDeclare declares a dead letter exchange and queue binding them together
func (builder *ThunderTopologyBuilder) deadLetterDeclare(channel *amqpclient.Channel, dlxName string) error {
	err := channel.ExchangeDeclare(
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

	_, err = channel.QueueDeclare(
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

	err = channel.QueueBind(
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
