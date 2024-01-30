package outboxpublisher

import (
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexdrl/zerowater"
	"github.com/gothunder/thunder/internal/events/rabbitmq"
	"github.com/rs/zerolog"
	amqpclient "github.com/streadway/amqp"
)

// newWatermillConfig addapts thunder constraints on top of watermill
// It loads the rabbitmq config and maps it to the watermill's amqp config
func newWatermillConfig(logger *zerolog.Logger) amqp.Config {
	config := rabbitmq.LoadConfig(logger)
	dlxName := config.QueueName + "_dlx"
	return amqp.Config{
		Connection: amqp.ConnectionConfig{
			AmqpURI: config.URL,
		},

		Marshaler: amqp.DefaultMarshaler{
			PostprocessPublishing: func(publishing amqpclient.Publishing) amqpclient.Publishing {
				publishing.ContentType = "application/json"
				return publishing
			},
		},

		Exchange: amqp.ExchangeConfig{
			GenerateName: func(topic string) string {
				return config.ExchangeName
			},
			Durable: true,
			Type:    "topic",
		},
		Queue: amqp.QueueConfig{
			GenerateName: func(topic string) string {
				return config.QueueName
			},
			Durable: true,
			Arguments: amqpclient.Table{
				"x-queue-type":           "quorum",
				"x-dead-letter-exchange": dlxName,
				"x-queue-leader-locator": "balanced",
			},
		},
		QueueBind: amqp.QueueBindConfig{
			GenerateRoutingKey: func(topic string) string {
				return topic
			},
		},

		Publish: amqp.PublishConfig{
			GenerateRoutingKey: func(topic string) string {
				return topic
			},
			Mandatory: true,
			Immediate: false,
		},
		Consume: amqp.ConsumeConfig{
			Qos: amqp.QosConfig{
				PrefetchCount: config.ConsumerConcurrency,
			},
		},

		TopologyBuilder: &ThunderTopologyBuilder{},
	}
}

// newRabbitMQOutPublisher creates a new rabbitmq publisher that publishes messages to the rabbitmq broker
// It uses the watermill library to publish messages
// It is used by the forwarder to publish messages from the outbox table to the rabbitmq broker
func newRabbitMQOutPublisher(logger *zerolog.Logger) (message.Publisher, error) {
	return amqp.NewPublisher(newWatermillConfig(logger), zerowater.NewZerologLoggerAdapter(logger.With().Logger()))
}
