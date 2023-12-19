package events

import (
	"context"

	"github.com/TheRafaBonin/roxy"
	"github.com/rs/zerolog/log"

	thunderLogger "github.com/gothunder/thunder/pkg/log"
)

type ErrorMap map[roxy.MessageAction]HandlerResponse

var (
	defaultErrorMap = ErrorMap{
		roxy.SuccessMessageAction: Success,
		roxy.DropMessageAction:    Success,
		roxy.RequeueMessageAction: Retry,
	}

	retryBackoffErrorMap = ErrorMap{
		roxy.SuccessMessageAction: Success,
		roxy.DropMessageAction:    Success,
		roxy.RequeueMessageAction: RetryBackoff,
	}
)

var (
	HandlerResponseLogActionMap = map[HandlerResponse]string{
		Success:      "message processed successfully",
		DeadLetter:   "dead lettering message",
		Retry:        "requeuing message",
		RetryBackoff: "requeuing message with backoff",
	}
)

func handleError(ctx context.Context, err error, errorMap ErrorMap) HandlerResponse {
	messageAction := roxy.GetDefaultMessageAction(err)
	logger := log.Ctx(ctx).With().Stack().Logger()

	thunderResponse, ok := errorMap[messageAction]
	if !ok {
		thunderResponse = DeadLetter
	}

	messageActionLog, ok := HandlerResponseLogActionMap[thunderResponse]
	if !ok {
		messageActionLog = "message handled with unknown action"
	}

	logger.Info().Err(err).Msg(messageActionLog)

	thunderLogger.LogError(ctx, err)
	return thunderResponse
}

// HandleError handles any error into a thunder response
func HandleError(ctx context.Context, err error) HandlerResponse {
	return handleError(ctx, err, defaultErrorMap)
}

func HandleErrorBackoff(ctx context.Context, err error) HandlerResponse {
	return handleError(ctx, err, retryBackoffErrorMap)
}

func HandleErrorWithCustomMap(ctx context.Context, err error, errorMap ErrorMap) HandlerResponse {
	return handleError(ctx, err, errorMap)
}
