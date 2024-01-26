package rabbitmq

import (
	"time"

	watermillamqp "github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/cenkalti/backoff/v4"
	stdAmqp "github.com/streadway/amqp"
)

func WhatermillConfig(conf Config) (watermillConfig watermillamqp.Config) {
	return watermillamqp.Config{
		Connection: watermillamqp.ConnectionConfig{
			AmqpURI: conf.URL,
			Reconnect: &watermillamqp.ReconnectConfig{
				BackoffInitialInterval:     backoff.DefaultInitialInterval,
				BackoffRandomizationFactor: backoff.DefaultRandomizationFactor,
				BackoffMultiplier:          backoff.DefaultMultiplier,
				BackoffMaxInterval:         15 * time.Minute,
			},
		},

		Marshaler: watermillamqp.DefaultMarshaler{
			PostprocessPublishing: func(publishing stdAmqp.Publishing) stdAmqp.Publishing {
				publishing.ContentType = "application/json"
				publishing.DeliveryMode = stdAmqp.Persistent

				return publishing
			},
			MessageUUIDHeaderKey: "x-outbox-message-id",
		},

		Exchange: watermillamqp.ExchangeConfig{
			GenerateName: func(_ string) string {
				return conf.ExchangeName
			},
			Type:    "topic",
			Durable: true,
		},
		Queue: watermillamqp.QueueConfig{
			GenerateName: func(_ string) string {
				return conf.QueueName
			},
			Durable: true,
			Arguments: map[string]interface{}{
				"x-queue-type":           "quorum",
				"x-dead-letter-exchange": "dlxName", // TODO
			},
		},
		QueueBind: watermillamqp.QueueBindConfig{
			GenerateRoutingKey: func(topic string) string {
				return topic
			},
		},
		Publish: watermillamqp.PublishConfig{
			GenerateRoutingKey: func(topic string) string {
				return topic
			},
			Mandatory: true,
		},
		Consume: watermillamqp.ConsumeConfig{
			Qos: watermillamqp.QosConfig{
				PrefetchCount: conf.ConsumerConcurrency,
			},
			// If more than one replica instance of service is is running, only one should consume
			Exclusive: true,
		},
		TopologyBuilder: &watermillamqp.DefaultTopologyBuilder{},
	}
}
