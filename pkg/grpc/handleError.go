package grpc

import (
	"context"

	"github.com/TheRafaBonin/roxy"

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
