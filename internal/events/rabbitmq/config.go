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
	ConsumerName        string
	ConsumerConcurrency int
}

func LoadConfig(log *zerolog.Logger) Config {
	c := Config{
		ExchangeName: os.Getenv("RABBIT_EXCHANGE"),
		QueueName:    os.Getenv("RABBIT_QUEUE"),
		URL:          os.Getenv("RABBIT_URL"),
		ConsumerName: os.Getenv("RABBIT_CONSUMER"),
	}

	if c.ExchangeName == "" {
		c.ExchangeName = "events"
		log.Info().Msgf("RABBIT_EXCHANGE is not set, defaulting to %s", c.ExchangeName)
	}

	if c.QueueName == "" {
		c.QueueName = "example_queue"
		log.Info().Msgf("RABBIT_QUEUE is not set, defaulting to %s", c.QueueName)
	}

	if c.URL == "" {
		c.URL = "amqp://guest:guest@localhost:5672"
		log.Info().Msgf("RABBIT_URL is not set, defaulting to %s", c.URL)
	}

	if c.ConsumerName == "" {
		c.ConsumerName = "example_consumer"
		log.Info().Msgf("RABBIT_CONSUMER is not set, defaulting to %s", c.ConsumerName)
	}

	concurrency := os.Getenv("RABBIT_CONSUMER_CONCURRENCY")
	if concurrency != "" {
		parsedConcurrency, err := strconv.Atoi(concurrency)
		if err == nil {
			c.ConsumerConcurrency = parsedConcurrency
		}
	}
	if c.ConsumerConcurrency == 0 {
		c.ConsumerConcurrency = 10
		log.Info().Msgf("RABBIT_CONSUMER_CONCURRENCY is not set, defaulting to %d", c.ConsumerConcurrency)
	}

	return c
}
