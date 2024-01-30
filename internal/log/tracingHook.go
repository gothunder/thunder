package log

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// TracingHook is a hook that adds trace and span IDs to the log
// It helps to correlate logs with traces, then improving observability and debugging
type TracingHook struct{}

func (h TracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return
	}
	spanID := spanContext.SpanID().String()
	traceID := spanContext.TraceID().String()

	e.Str("trace-id", traceID).Str("span-id", spanID)
}
