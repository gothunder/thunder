package graphql

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func errorPresenter(ctx context.Context, err error) *gqlerror.Error {
	var gqlErr *gqlerror.Error
	if errors.As(err, &gqlErr) {
		if gqlErr.Extensions["status"] != nil {
			return gqlErr
		}
	}

	logger := log.Ctx(ctx).With().Stack().Logger()
	logger.WithLevel(zerolog.PanicLevel).
		Err(err).
		Msg("response not provided")

	return internalError
}
