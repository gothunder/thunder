package graphql

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
)

func aroundOperations(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	oc := graphql.GetOperationContext(ctx)

	logger := log.Ctx(ctx)
	logger.Info().
		Str("operation", oc.RawQuery).
		Msg("processing request")

	start := time.Now()
	defer func() {
		logger.Info().
			Dur("latency", time.Since(start)).
			Msg("request processed")
	}()

	return next(ctx)
}
