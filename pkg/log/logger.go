package log

import (
	"fmt"
	"os"

	internalLog "github.com/gothunder/thunder/internal/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

func NewLogger(output diode.Writer) *zerolog.Logger {
	return internalLog.NewLogger(output)
}

func NewDiode() diode.Writer {
	return diode.NewWriter(os.Stdout, 1000, 0, func(missed int) {
		fmt.Printf("Dropped %d messages\n", missed)
	})
}

func DiodeShutdown(d diode.Writer) {
	d.Close()
}
