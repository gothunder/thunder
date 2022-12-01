package graphql

import (
	"context"

	"github.com/gothunder/thunder/internal/graphql"
	"github.com/gothunder/thunder/pkg/response"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func HandleResponse(ctx context.Context, res response.Response) *gqlerror.Error {
	return graphql.HandleResponse(ctx, res)
}
