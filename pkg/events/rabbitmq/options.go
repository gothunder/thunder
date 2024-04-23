package rabbitmq

import (
	"fmt"
	"time"

	"github.com/gothunder/thunder/internal/events/rabbitmq"
)

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
