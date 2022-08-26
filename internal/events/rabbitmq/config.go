package rabbitmq

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

type Config struct {
	ExchangeName        string
	QueueName           string
	URL                 string
	ConsumerConcurrency int
}

func LoadConfig(log *zerolog.Logger) Config {
	c := Config{
		ExchangeName: os.Getenv("RABBITMQ_EXCHANGE_NAME"),
		QueueName:    os.Getenv("RABBITMQ_QUEUE_NAME"),
		URL:          os.Getenv("RABBITMQ_URL"),
	}

	if c.ExchangeName == "" {
		c.ExchangeName = "events"
		log.Warn().Msgf("RABBITMQ_EXCHANGE_NAME is not set, defaulting to %s", c.ExchangeName)
	}

	if c.QueueName == "" {
		c.QueueName = "example_queue"
		log.Warn().Msgf("RABBITMQ_QUEUE_NAME is not set, defaulting to %s", c.QueueName)
	}

	if c.URL == "" {
		c.URL = "amqp://guest:guest@localhost:5672"
		log.Warn().Msgf("RABBITMQ_URL is not set, defaulting to %s", c.URL)
	}

	concurrency := os.Getenv("RABBITMQ_CONSUMER_CONCURRENCY")
	if concurrency != "" {
		parsedConcurrency, err := strconv.Atoi(concurrency)
		if err != nil {
			parsedConcurrency = 2
			log.Warn().Err(err).Msgf("failed to parse RABBITMQ_CONSUMER_CONCURRENCY, defaulting to %s", parsedConcurrency)
		}

		c.ConsumerConcurrency = parsedConcurrency
	}

	return c
}
