package rabbitmq

import (
	"os"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog"
)

const (
	// DefaultInitialInterval is the default initial interval for the backoff
	DefaultInitialInterval = backoff.DefaultInitialInterval
	// DefaultRandomizationFactor is the default randomization factor for the backoff
	DefaultRandomizationFactor = backoff.DefaultRandomizationFactor
	// DefaultMultiplier is the default multiplier for the backoff
	DefaultMultiplier = 2
	// DefaultMaxInterval is the default max interval for the backoff
	DefaultMaxInterval = backoff.DefaultMaxInterval
	// DefaultMaxRetries is the default max retries for the backoff
	DefaultMaxRetries = 5
	// DefaultDeleteDLX is the default value for deleting the DLX before creating it
	DefaultDeleteDLX = false
	// DefaultDisableConsumer is the default value for disabling the consumer
	DefaultDisableConsumer = false
)

type Config struct {
	ExchangeName        string
	QueueName           string
	URL                 string
	ConsumerName        string
	ConsumerConcurrency int
	PrefetchCount       int

	MaxRetries          int
	InitialInterval     time.Duration
	RandomizationFactor float64
	Multiplier          float64
	MaxInterval         time.Duration

	DeleteDLX       bool
	DisableConsumer bool
}

type RabbitmqConfigOption func(*Config)

func LoadConfig(log *zerolog.Logger, opts ...RabbitmqConfigOption) Config {
	c := Config{
		ExchangeName:        os.Getenv("RABBIT_EXCHANGE"),
		QueueName:           os.Getenv("RABBIT_QUEUE"),
		URL:                 os.Getenv("RABBIT_URL"),
		ConsumerName:        os.Getenv("RABBIT_CONSUMER"),
		MaxRetries:          DefaultMaxRetries,
		InitialInterval:     DefaultInitialInterval,
		RandomizationFactor: DefaultRandomizationFactor,
		Multiplier:          DefaultMultiplier,
		MaxInterval:         DefaultMaxInterval,
		DeleteDLX:           DefaultDeleteDLX,
		DisableConsumer:     DefaultDisableConsumer,
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

	prefetchCount := os.Getenv("RABBIT_PREFETCH_COUNT")
	if prefetchCount != "" {
		parsedPrefetchCount, err := strconv.Atoi(prefetchCount)
		if err == nil {
			c.PrefetchCount = parsedPrefetchCount
		}
	}
	if c.PrefetchCount == 0 {
		c.PrefetchCount = c.ConsumerConcurrency
		log.Info().Msgf("RABBIT_PREFETCH_COUNT is not set, defaulting to %d", c.PrefetchCount)
	}

	maxRetries := os.Getenv("RABBIT_MAX_RETRIES")
	if maxRetries != "" {
		parsedMaxRetries, err := strconv.Atoi(maxRetries)
		if err == nil {
			c.MaxRetries = parsedMaxRetries
		}
	}

	initialInterval := os.Getenv("RABBIT_INITIAL_INTERVAL")
	if initialInterval != "" {
		parsedInitialInterval, err := time.ParseDuration(initialInterval)
		if err == nil {
			c.InitialInterval = parsedInitialInterval
		}
	}

	randomizationFactor := os.Getenv("RABBIT_RANDOMIZATION_FACTOR")
	if randomizationFactor != "" {
		parsedRandomizationFactor, err := strconv.ParseFloat(randomizationFactor, 64)
		if err == nil {
			c.RandomizationFactor = parsedRandomizationFactor
		}
	}

	multiplier := os.Getenv("RABBIT_MULTIPLIER")
	if multiplier != "" {
		parsedMultiplier, err := strconv.ParseFloat(multiplier, 64)
		if err == nil {
			c.Multiplier = parsedMultiplier
		}
	}

	maxInterval := os.Getenv("RABBIT_MAX_INTERVAL")
	if maxInterval != "" {
		parsedMaxInterval, err := time.ParseDuration(maxInterval)
		if err == nil {
			c.MaxInterval = parsedMaxInterval
		}
	}

	deleteDLX := os.Getenv("RABBIT_DELETE_DLX")
	if deleteDLX != "" {
		parsedDeleteDLX, err := strconv.ParseBool(deleteDLX)
		if err == nil {
			c.DeleteDLX = parsedDeleteDLX
		}
	}

	disableConsumer := os.Getenv("RABBIT_DISABLE_CONSUMER")
	if disableConsumer != "" {
		parsedDisableConsumer, err := strconv.ParseBool(disableConsumer)
		if err == nil {
			c.DisableConsumer = parsedDisableConsumer
		}
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}
