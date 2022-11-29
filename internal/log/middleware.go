package log

import (
	"net/http"

	"github.com/rs/zerolog"
)

func Middleware(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	fn := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			log := logger.
				With().
				Str("ip", r.RemoteAddr).
				Logger()

			// Add logger to context
			ctx := log.WithContext(r.Context())
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(rw, r)
		})
	}
	return fn
}
