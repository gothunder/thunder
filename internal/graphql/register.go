package graphql

import (
	"context"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gothunder/thunder/internal/recoverer"
	"github.com/ravilushqa/otelgqlgen"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func CreateHandler(graphQLSchema graphql.ExecutableSchema) *handler.Server {
	// Create handler with explicit transport registration
	graphqlHandler := handler.New(graphQLSchema)

	// Register transports (order matters — first match wins)
	// SSE must be before POST since both match POST+application/json,
	// but SSE additionally requires Accept: text/event-stream.
	graphqlHandler.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})
	graphqlHandler.AddTransport(transport.Options{})
	graphqlHandler.AddTransport(transport.GET{})
	graphqlHandler.AddTransport(SSEWithWriteDeadline{
		WriteDeadline: 5 * time.Minute,
	})
	graphqlHandler.AddTransport(transport.POST{})
	graphqlHandler.AddTransport(transport.MultipartForm{})

	// Query caching and extensions (same as NewDefaultServer)
	graphqlHandler.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	graphqlHandler.Use(extension.Introspection{})
	graphqlHandler.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

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
