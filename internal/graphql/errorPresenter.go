package graphql

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/99designs/gqlgen/graphql"
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

	var query string
	logger := log.Ctx(ctx).With().Stack().Logger()
	if graphql.HasOperationContext(ctx) {
		oc := graphql.GetOperationContext(ctx)
		query = oc.RawQuery
	}

	logger.WithLevel(lvl).
		Err(err).
		Str("requestID", requestID).
		Str("query", query).
		Msg("response not provided")

	return internalError
}
