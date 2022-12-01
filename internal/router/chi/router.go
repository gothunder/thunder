package chi

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func NewRouter(logger *zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Apply default middlewares
	defaultMiddlewares(r, logger)

	return r
}
