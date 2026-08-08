package log

import (
	"net/http"

	thunderCtx "github.com/gothunder/thunder/pkg/context"
	"github.com/rs/zerolog"
)

func Middleware(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	fn := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			log := logger.
				With().
				Str("ip", r.RemoteAddr).
				Logger()

			requestID := thunderCtx.RequestIDFromContext(r.Context())
			if requestID != "" {
				log = log.With().Str("requestID", requestID).Logger()
			}

			// Add logger to context
			ctx := log.WithContext(r.Context())
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(rw, r)
		})
	}
	return fn
}
