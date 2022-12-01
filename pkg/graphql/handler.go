package graphql

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"

	thunderGraphql "github.com/gothunder/thunder/internal/graphql"
	"github.com/gothunder/thunder/pkg/router"
)

func CreateHandler(graphQLSchema graphql.ExecutableSchema) *handler.Server {
	return thunderGraphql.CreateHandler(graphQLSchema)
}

func newGraphqlHandler(graphQLSchema graphql.ExecutableSchema) router.HandlerOutput {
	return router.HandlerOutput{
		Handler: thunderGraphql.NewGraphqlHandler(graphQLSchema),
	}
}
