package log

import (
	"github.com/gothunder/thunder/internal/log"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var Module = fx.Options(
	fx.Provide(NewDiode),
	fx.Provide(NewLogger),
	fx.WithLogger(func(logger *zerolog.Logger) fxevent.Logger {
		return &log.ZeroLogger{
			Logger: logger,
		}
	}),
)
