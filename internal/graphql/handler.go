package graphql

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gothunder/thunder/pkg/router"
)

type graphqlHandler struct {
	handler *handler.Server
}

func NewGraphqlHandler(graphQLSchema graphql.ExecutableSchema) router.HTTPHandler {
	return &graphqlHandler{
		handler: CreateHandler(graphQLSchema),
	}
}

func (h *graphqlHandler) Method() string {
	return "POST"
}

func (h *graphqlHandler) Pattern() string {
	return "/query"
}

func (h *graphqlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}
