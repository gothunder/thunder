package rabbitmq

import (
	"os"
	"time"

	"github.com/gothunder/thunder/internal/events/rabbitmq/publisher"
	"github.com/gothunder/thunder/pkg/events"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

func NewRabbitMQPublisher() (events.EventPublisher, error) {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	logger := zerolog.
		New(output).
		With().
		Timestamp().
		Logger()

	return publisher.NewPublisher(amqp091.Config{}, &logger)
}
