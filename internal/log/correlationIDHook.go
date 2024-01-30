package log

import (
	"github.com/rs/zerolog"

	thunderContext "github.com/gothunder/thunder/pkg/context"
)

// CorrelationIDHook is a hook that adds correlation ID to the log
// It helps to correlate logs that belong to the same request
type CorrelationIDHook struct{}

func (h CorrelationIDHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	correlationID := thunderContext.CorrelationIDFromContext(ctx)
	if correlationID != "" {
		e.Str("correlation-id", correlationID)
	}
}
