package events

import (
	"context"

	"github.com/TheRafaBonin/roxy"
	"github.com/rs/zerolog/log"

	thunderLogger "github.com/gothunder/thunder/pkg/log"
)

// HandleError handles any error into a thunder response
func HandleError(ctx context.Context, err error) HandlerResponse {
	messageAction := roxy.GetDefaultMessageAction(err)
	logger := log.Ctx(ctx).With().Stack().Logger()
	var thunderResponse HandlerResponse

	switch messageAction {
	case roxy.SuccessMessageAction, roxy.DropMessageAction:
		thunderResponse = Success

	case roxy.RequeueMessageAction:
		logger.Info().Err(err).Msg("requeuing message")
		thunderResponse = Retry

	case roxy.DeadLetterMessageAction:
		logger.Info().Err(err).Msg("dead lettering message")
		thunderResponse = DeadLetter
	default:
		logger.Info().Err(err).Msg("dead lettering message")
		thunderResponse = DeadLetter
	}

	thunderLogger.LogError(ctx, err)
	return thunderResponse
}
