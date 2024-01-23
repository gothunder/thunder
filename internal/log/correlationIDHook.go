package log

import (
	"github.com/rs/zerolog"

	thunderContext "github.com/gothunder/thunder/pkg/context"
)

type CorrelationIDHook struct{}

func (h CorrelationIDHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	e.Str("correlation-id", thunderContext.CorrelationIDFromContext(ctx))
}
