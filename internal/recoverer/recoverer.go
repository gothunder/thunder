package recoverer

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Recoverer(ctx context.Context, p interface{}) {
	// Convert panic to error
	err, ok := p.(error)
	if !ok {
		err = fmt.Errorf("%v", p)
	}

	err = eris.Wrap(err, "recovered from panic")

	logger := log.Ctx(ctx).With().Stack().Logger()
	logger.WithLevel(zerolog.PanicLevel).
		Err(err).
		Msg("panic")
}
