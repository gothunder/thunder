package graphql

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gothunder/thunder/internal/recoverer"
	"github.com/ravilushqa/otelgqlgen"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func CreateHandler(graphQLSchema graphql.ExecutableSchema) *handler.Server {
	// Create a new handler
	graphqlHandler := handler.NewDefaultServer(graphQLSchema)

	// Set the error handler
	graphqlHandler.SetErrorPresenter(errorPresenter)

	// Set the panic handler
	graphqlHandler.SetRecoverFunc(func(ctx context.Context, p interface{}) error {
		recoverer.Recoverer(ctx, p)
		return internalError
	})

	// Add a middleware to log the request
	graphqlHandler.AroundOperations(aroundOperations)

	// Add otel middleware
	// It will create traces for each request making possible to see the whole request lifecycle
	graphqlHandler.Use(otelgqlgen.Middleware())

	return graphqlHandler
}

// A default internal error when something goes wrong
var internalError *gqlerror.Error = &gqlerror.Error{
	Message: http.StatusText(http.StatusInternalServerError),
	Extensions: map[string]interface{}{
		"status": http.StatusInternalServerError,
	},
}
