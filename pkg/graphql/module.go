package graphql

import (
	"github.com/gothunder/thunder/internal/graphql"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		graphql.NewGraphqlHandler,
	),
)
