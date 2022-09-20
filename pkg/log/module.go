package log

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

func NewLogger() *zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	logger := zerolog.
		New(output).
		With().
		Timestamp().
		Logger()

	return &logger
}

var Module = fx.Provide(NewLogger)
