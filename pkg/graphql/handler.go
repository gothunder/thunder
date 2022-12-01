package graphql

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"

	thunderGraphql "github.com/gothunder/thunder/internal/graphql"
)

func CreateHandler(graphQLSchema graphql.ExecutableSchema) *handler.Server {
	return thunderGraphql.CreateHandler(graphQLSchema)
}
