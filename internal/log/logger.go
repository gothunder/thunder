package log

import (
	"io"
	"os"
	"time"

	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

func NewLogger(output io.Writer) *zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.ErrorStackMarshaler = func(err error) interface{} {
		return eris.ToJSON(err, true)
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "none" {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}
	if logLevel == "trace" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
	if logLevel == "debug" || logLevel == "" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "test" || environment == "local" || environment == "" {
		output = zerolog.ConsoleWriter{Out: output, TimeFormat: time.RFC3339}
	}

	logger := zerolog.
		New(output).
		Hook(TracingHook{}).
		Hook(CorrelationIDHook{}).
		With().
		Timestamp().
		Logger()
	return &logger
}
