package chi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gothunder/thunder/pkg/router"
	"github.com/rs/zerolog"
)

func registerRoutes(handlers []router.HTTPHandler, logger *zerolog.Logger, r *chi.Mux) {
	// This never gets called, because the heartbeat middleware intercepts the request first
	// The reason this is here it's because if some services don't register any routes, the router always returns 404 so it breaks the health checks.
	r.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	// Register all provided instances
	for _, handler := range handlers {
		logger.Debug().Msgf("Registering %s %s", handler.Pattern(), handler.Method())

		r.Method(handler.Method(), handler.Pattern(), handler)
	}
}
