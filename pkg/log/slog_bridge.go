package log

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/rs/zerolog"
)

type slogWriter struct {
	logger *slog.Logger
}

func (w *slogWriter) Write(p []byte) (n int, err error) {
	var entry map[string]interface{}
	if err := json.Unmarshal(p, &entry); err != nil {
		w.logger.Info(string(p))
		return len(p), nil
	}

	level := slog.LevelInfo
	if lvl, ok := entry["level"].(string); ok {
		switch lvl {
		case "debug", "trace":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn", "warning":
			level = slog.LevelWarn
		case "error", "fatal", "panic":
			level = slog.LevelError
		}
		delete(entry, "level")
	}

	msg := ""
	if m, ok := entry["message"].(string); ok {
		msg = m
		delete(entry, "message")
	}

	delete(entry, "time")

	args := make([]any, 0, len(entry)*2)
	for k, v := range entry {
		args = append(args, k, v)
	}

	w.logger.Log(context.Background(), level, msg, args...)
	return len(p), nil
}

// NewLoggerFromSlog creates a *zerolog.Logger that outputs to the given slog.Logger.
// Use this when your service uses slog but needs to pass a logger to Thunder.
//
// Example:
//
//	logger := log.NewLoggerFromSlog(slog.Default())
//	consumer, _ := rabbitmq.NewRabbitMQConsumer(logger, ...)
func NewLoggerFromSlog(slogger *slog.Logger) *zerolog.Logger {
	writer := &slogWriter{logger: slogger}
	logger := zerolog.New(writer).With().Timestamp().Logger()
	return &logger
}
