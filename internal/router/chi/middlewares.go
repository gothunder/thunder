package chi

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gothunder/thunder/internal/log"
	thunderMiddlewares "github.com/gothunder/thunder/internal/router/chi/middlewares"
	"github.com/riandyrn/otelchi"
	"github.com/rs/zerolog"
)

func defaultMiddlewares(r *chi.Mux, logger *zerolog.Logger) {
	// Heartbeat comes first to keep the healthcheck lean
	r.Use(middleware.Heartbeat("/health"))

	// Add a correlation ID to the context
	r.Use(thunderMiddlewares.CorrelationID)

	// Add tracing to the context, making possible to see the whole request lifecycle
	r.Use(otelchi.Middleware(os.Getenv("SERVICE_NAME"), otelchi.WithChiRoutes(r)))

	// Ggets sets the right RemoteAddr for the request
	r.Use(middleware.RealIP)

	// Adds a request id to the context of each request
	r.Use(middleware.RequestID)

	// Add a logger instance to the context with some default values
	r.Use(log.Middleware(logger))
}
