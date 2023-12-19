package grpc

import (
	"context"
	"errors"
	"net/http"

	"github.com/TheRafaBonin/roxy"

	thunderGraphql "github.com/gothunder/thunder/pkg/graphql"
	thunderLogger "github.com/gothunder/thunder/pkg/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleError handles any error into a thunder response
func HandleError(ctx context.Context, err error) error {
	// Declare some variables
	grpcResponse := roxy.GetDefaultGrpcResponse(err)
	grpcMessage := grpcResponse.Message
	grpcCode := grpcResponse.Code

	// Checks the OK case
	if grpcCode == codes.OK {
		return nil
	}

	// Log and return
	thunderLogger.LogError(ctx, err)
	return status.Error(grpcCode, grpcMessage)
}

func HandleGrpcErrorIgnoringNotFound(ctx context.Context, err error) error {
	var bareStatusErr error
	_, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		bareStatusErr = err
	} else {
		bareStatusErr = errors.Unwrap(err)
	}
	if statusCode, ok := status.FromError(bareStatusErr); ok {
		switch statusCode.Code() {
		case codes.NotFound:
			return nil
		case codes.InvalidArgument:
			return thunderGraphql.HandleError(ctx, roxy.SetDefaultHTTPResponse(err, roxy.HTTPResponse{
				Message: statusCode.Message(),
				Status:  http.StatusBadRequest,
			}))
		case codes.Internal:
			return thunderGraphql.HandleError(ctx, roxy.SetDefaultHTTPResponse(err, roxy.HTTPResponse{
				Message: "internal error",
				Status:  http.StatusInternalServerError,
			}))
		default:
			return thunderGraphql.HandleError(ctx, roxy.SetDefaultHTTPResponse(err, roxy.HTTPResponse{
				Message: "internal error",
				Status:  http.StatusInternalServerError,
			}))
		}
	} else {
		return thunderGraphql.HandleError(ctx, roxy.SetDefaultHTTPResponse(err, roxy.HTTPResponse{
			Message: "internal error",
			Status:  http.StatusInternalServerError,
		}))
	}
}

func GetStatusCodeFromRawError(err error) codes.Code {
	var bareStatusErr error
	_, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		bareStatusErr = err
	} else {
		bareStatusErr = errors.Unwrap(err)
	}

	if statusCode, ok := status.FromError(bareStatusErr); ok {
		return statusCode.Code()
	}

	return codes.Internal
}

func HandleGrpcError(ctx context.Context, err error) error {
	var bareStatusErr error
	_, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		bareStatusErr = err
	} else {
		bareStatusErr = errors.Unwrap(err)
	}
	if statusCode, ok := status.FromError(bareStatusErr); ok {
		switch statusCode.Code() {
		case codes.NotFound:
			return thunderGraphql.HandleError(ctx, roxy.SetDefaultHTTPResponse(err, roxy.HTTPResponse{
				Message: statusCode.Message(),
				Status:  http.StatusNotFound,
			}))
		case codes.InvalidArgument:
			return thunderGraphql.HandleError(ctx, roxy.SetDefaultHTTPResponse(err, roxy.HTTPResponse{
				Message: statusCode.Message(),
				Status:  http.StatusBadRequest,
			}))
		case codes.Internal:
			return thunderGraphql.HandleError(ctx, roxy.SetDefaultHTTPResponse(err, roxy.HTTPResponse{
				Message: "internal error",
				Status:  http.StatusInternalServerError,
			}))
		default:
			return thunderGraphql.HandleError(ctx, roxy.SetDefaultHTTPResponse(err, roxy.HTTPResponse{
				Message: "internal error",
				Status:  http.StatusInternalServerError,
			}))
		}
	} else {
		return thunderGraphql.HandleError(ctx, roxy.SetDefaultHTTPResponse(err, roxy.HTTPResponse{
			Message: "internal error",
			Status:  http.StatusInternalServerError,
		}))
	}
}
