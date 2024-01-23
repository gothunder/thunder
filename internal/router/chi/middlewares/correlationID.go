package middlewares

import (
	"net/http"

	thunderContext "github.com/gothunder/thunder/pkg/context"
)

func CorrelationID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		correlationID := r.Header.Get("X-Correlation-ID")
		ctx = thunderContext.ContextWithCorrelationID(ctx, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
