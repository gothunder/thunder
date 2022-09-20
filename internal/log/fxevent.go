// Taken from https://github.com/efectn/fx-zerolog/blob/main/zerolog.go
package log

import (
	"strings"

	"github.com/rs/zerolog"
	"go.uber.org/fx/fxevent"
)

// ZapLogger is an Fx event logger that logs events to Zerolog.
type ZeroLogger struct {
	Logger *zerolog.Logger
}

// LogEvent logs the given event to the provided Zerolog.
func (l *ZeroLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.Logger.Debug().
			Str("callee", e.FunctionName).
			Str("caller", e.CallerName).
			Msg("OnStart hook executing")

	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.Logger.Err(e.Err).
				Str("callee", e.FunctionName).
				Str("caller", e.CallerName).
				Msg("OnStart hook failed")
		} else {
			l.Logger.Debug().
				Str("callee", e.FunctionName).
				Str("caller", e.CallerName).
				Str("runtime", e.Runtime.String()).
				Msg("OnStart hook executed")
		}
	case *fxevent.OnStopExecuting:
		l.Logger.Debug().
			Str("callee", e.FunctionName).
			Str("caller", e.CallerName).
			Msg("OnStop hook executing")
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.Logger.Err(e.Err).
				Str("callee", e.FunctionName).
				Str("caller", e.CallerName).
				Msg("OnStop hook failed")
		} else {
			l.Logger.Debug().
				Str("callee", e.FunctionName).
				Str("caller", e.CallerName).
				Str("runtime", e.Runtime.String()).
				Msg("OnStop hook executed")
		}
	case *fxevent.Supplied:
		l.Logger.Err(e.Err).
			Str("type", e.TypeName).
			Str("module", e.ModuleName).
			Msg("supplied")
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.Logger.Debug().
				Str("constructor", e.ConstructorName).
				Str("module", e.ModuleName).
				Str("type", rtype).
				Msg("provided")
		}
		if e.Err != nil {
			l.Logger.Err(e.Err).
				Str("module", e.ModuleName).
				Msg("error encountered while applying options")
		}
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			l.Logger.Debug().
				Str("decorator", e.DecoratorName).
				Str("module", e.ModuleName).
				Str("type", rtype).
				Msg("decorated")
		}
		if e.Err != nil {
			l.Logger.Err(e.Err).
				Str("module", e.ModuleName).
				Msg("error encountered while applying options")
		}
	case *fxevent.Invoking:
		// Do not log stack as it will make logs hard to read.
		l.Logger.Debug().
			Str("function", e.FunctionName).
			Str("module", e.ModuleName).
			Msg("invoking")
	case *fxevent.Invoked:
		if e.Err != nil {
			l.Logger.Err(e.Err).
				Str("stack", e.Trace).
				Str("function", e.FunctionName).
				Msg("invoke failed")
		}
	case *fxevent.Stopping:
		l.Logger.Debug().
			Str("signal", strings.ToUpper(e.Signal.String())).
			Msg("received signal")
	case *fxevent.Stopped:
		if e.Err != nil {
			l.Logger.Err(e.Err).
				Msg("stop failed")
		}
	case *fxevent.RollingBack:
		l.Logger.Err(e.StartErr).
			Msg("start failed, rolling back")
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.Logger.Err(e.Err).
				Msg("rollback failed")
		}
	case *fxevent.Started:
		if e.Err != nil {
			l.Logger.Err(e.Err).
				Msg("start failed")
		} else {
			l.Logger.Debug().
				Msg("started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.Logger.Err(e.Err).
				Msg("custom logger initialization failed")
		} else {
			l.Logger.Debug().
				Str("function", e.ConstructorName).
				Msg("initialized custom fxevent.Logger")
		}
	}
}
