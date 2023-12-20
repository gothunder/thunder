package graphql

import (
	"context"
	"errors"

	"github.com/go-chi/chi/v5/middleware"
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

	requestID := middleware.GetReqID(ctx)

	logger := log.Ctx(ctx).With().Stack().Logger()
	logger.WithLevel(zerolog.PanicLevel).
		Err(err).
		Str("requestID", requestID).
		Msg("response not provided")

	return internalError
}
