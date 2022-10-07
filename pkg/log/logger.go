package log

import (
	"context"
	"fmt"
	"os"

	internalLog "github.com/gothunder/thunder/internal/log"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"go.uber.org/fx"
)

func NewLogger(output diode.Writer) *zerolog.Logger {
	return internalLog.NewLogger(output)
}

func NewDiode() diode.Writer {
	return diode.NewWriter(os.Stdout, 1000, 0, func(missed int) {
		fmt.Printf("Dropped %d messages\n", missed)
	})
}

func diodeShutdown(d diode.Writer, lc fx.Lifecycle) {
	lc.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				// Wait till the rest of the app has stopped or timed out
				<-ctx.Done()

				// Flush the diode
				err := d.Close()
				if err != nil {
					return eris.Wrap(err, "failed to close diode")
				}

				return nil
			},
		},
	)
}
