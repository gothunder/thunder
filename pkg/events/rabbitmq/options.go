package rabbitmq

import (
	"fmt"
	"time"

	"github.com/gothunder/thunder/internal/events/rabbitmq"
)

func WithURL(url string) rabbitmq.RabbitmqConfigOption {
	return func(c *rabbitmq.Config) {
		c.URL = url
	}
}

func WithExchangeName(name string) rabbitmq.RabbitmqConfigOption {
	return func(c *rabbitmq.Config) {
		c.ExchangeName = name
	}
}

func WithQueueName(name string) rabbitmq.RabbitmqConfigOption {
	return func(c *rabbitmq.Config) {
		c.QueueName = name
	}
}

func WithConsumerName(name string) rabbitmq.RabbitmqConfigOption {
	return func(c *rabbitmq.Config) {
		c.ConsumerName = name
	}
}

// WithConsumerConcurrency sets the number of concurrent message handlers
func WithConsumerConcurrency(concurrency int) rabbitmq.RabbitmqConfigOption {
	return func(c *rabbitmq.Config) {
		c.ConsumerConcurrency = concurrency
	}
}

func WithPrefetchCount(prefetch int) rabbitmq.RabbitmqConfigOption {
	return func(c *rabbitmq.Config) {
		c.PrefetchCount = prefetch
	}
}

func WithMaxRetries(maxRetries int) rabbitmq.RabbitmqConfigOption {
	return func(c *rabbitmq.Config) {
		c.MaxRetries = maxRetries
	}
}

func WithDeleteDLX(deleteDLX bool) rabbitmq.RabbitmqConfigOption {
	return func(c *rabbitmq.Config) {
		c.DeleteDLX = deleteDLX
	}
}

func WithQueueNamePosfix(posfix string) rabbitmq.RabbitmqConfigOption {
	return func(c *rabbitmq.Config) {
		c.QueueName = fmt.Sprintf("%s_%s", c.QueueName, posfix)
	}
}

type ExponentialBackoff struct {
	InitialInterval     time.Duration
	RandomizationFactor float64
	Multiplier          float64
	MaxInterval         time.Duration
	MaxRetries          int
}

func WithExponentialBackoff(exponentialBackoff ExponentialBackoff) rabbitmq.RabbitmqConfigOption {
	return func(c *rabbitmq.Config) {
		c.InitialInterval = exponentialBackoff.InitialInterval
		c.RandomizationFactor = exponentialBackoff.RandomizationFactor
		c.Multiplier = exponentialBackoff.Multiplier
		c.MaxInterval = exponentialBackoff.MaxInterval
		c.MaxRetries = exponentialBackoff.MaxRetries
	}
}
