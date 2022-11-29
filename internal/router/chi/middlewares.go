package chi

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gothunder/thunder/internal/log"
	"github.com/rs/zerolog"
)

func defaultMiddlewares(r *chi.Mux, logger *zerolog.Logger) {
	// Heartbeat comes first to keep the healthcheck lean
	r.Use(middleware.Heartbeat("/health"))

	// Ggets sets the right RemoteAddr for the request
	r.Use(middleware.RealIP)

	// Add a logger instance to the context with some default values
	r.Use(log.Middleware(logger))
}
