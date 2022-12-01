package graphql

import (
	"context"
	"net/http"

	"github.com/gothunder/thunder/pkg/response"
	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func HandleResponse(ctx context.Context, res response.Response) *gqlerror.Error {
	logger := log.Ctx(ctx)

	if res.Status == 0 || res.Message == "" {
		res.Status = http.StatusInternalServerError
		res.Message = http.StatusText(http.StatusInternalServerError)

		logger.Warn().
			Msg("response status or message is empty")
	}

	logger.Info().
		Int("status", res.Status).
		Str("response", res.Message).
		Msg("response for request")

	if res.Status == http.StatusOK {
		return nil
	}

	return &gqlerror.Error{
		Message: res.Message,
		Extensions: map[string]interface{}{
			"status": res.Status,
		},
	}
}
