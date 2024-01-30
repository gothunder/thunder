package graphql

import (
	"context"
	"errors"
	"io"
	"strings"

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

	lvl := zerolog.PanicLevel
	if errors.Is(err, io.ErrUnexpectedEOF) || strings.Contains(err.Error(), "unexpected EOF") {
		lvl = zerolog.ErrorLevel
	}

	logger := log.Ctx(ctx).With().Stack().Logger()
	logger.WithLevel(lvl).
		Err(err).
		Str("requestID", requestID).
		Msg("response not provided")

	return internalError
}
