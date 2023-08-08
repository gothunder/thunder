package log

import (
	"context"

	"github.com/TheRafaBonin/roxy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogError logs error using zerolog based on roxy's log level
func LogError(ctx context.Context, err error) {
	var loggerEvent *zerolog.Event

	logger := log.Ctx(ctx).With().Stack().Logger()
	logLevel := roxy.GetErrorLogLevel(err)

	if err == nil {
		return
	}

	switch logLevel {
	case roxy.Disabled:
		return
	case roxy.TraceLevel:
		loggerEvent = logger.Trace()
	case roxy.DebugLevel:
		loggerEvent = logger.Debug()
	case roxy.InfoLevel:
		loggerEvent = logger.Info()
	case roxy.WarnLevel:
		loggerEvent = logger.Warn()
	case roxy.ErrorLevel:
		loggerEvent = logger.Error()
	case roxy.PanicLevel:
		loggerEvent = logger.Panic()
	case roxy.FatalLevel:
		loggerEvent = logger.Fatal()
	default:
		loggerEvent = logger.Error()
	}

	loggerEvent.Err(err).Msg(roxy.Cause(err).Error())
}
