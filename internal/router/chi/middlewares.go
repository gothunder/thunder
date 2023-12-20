package chi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gothunder/thunder/internal/log"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

func defaultMiddlewares(r *chi.Mux, logger *zerolog.Logger) {
	// Heartbeat comes first to keep the healthcheck lean
	r.Use(middleware.Heartbeat("/health"))

	// Ggets sets the right RemoteAddr for the request
	r.Use(middleware.RealIP)

	// Adds a request id to the context of each request
	r.Use(middleware.RequestID)

	// Add a logger instance to the context with some default values
	r.Use(log.Middleware(logger))

	// Logs the request
	r.Use(requestDumpMiddleware())
}

func requestDumpMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var strBody string

			body := r.Body
			bytesBody, err := io.ReadAll(body)
			if err != nil {
				strBody = fmt.Sprintf("error reading body: %s", err.Error())
			} else {
				// Rewrites body to allow for further reading
				r.Body = io.NopCloser(bytes.NewBuffer(bytesBody))

				strBody = string(bytesBody)
			}

			logger := zlog.Ctx(r.Context())

			headerBuffer := bytes.Buffer{}
			var strHeader string

			err = r.Header.Write(&headerBuffer)
			if err != nil {
				strHeader = fmt.Sprintf("error writing headers to string: %s", err.Error())
			} else {
				strHeader = headerBuffer.String()
			}

			logger.Debug().
				Str("headers", strHeader).
				Str("body", strBody).
				Str("requestID", middleware.GetReqID(r.Context())).
				Msg("logging request")

			next.ServeHTTP(w, r)
		})
	}

}
