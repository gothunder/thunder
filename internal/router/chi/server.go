package chi

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gothunder/thunder/pkg/router"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

func CreateServer(handlers []router.HTTPHandler, logger *zerolog.Logger, r *chi.Mux) (*http.Server, *net.Listener, error) {
	// Register all handlers before starting the server
	registerRoutes(handlers, logger, r)

	// Create the server and listener on the port
	addr := ":" + os.Getenv("PORT")
	server := &http.Server{
		Addr:    addr, // configure the bind address
		Handler: r,    // set the default handler

		// Good practice to set timeouts to avoid Slowloris attacks.
		ReadTimeout:  5 * time.Second,   // max time to read request from the client
		WriteTimeout: 10 * time.Second,  // max time to write response to the client
		IdleTimeout:  120 * time.Second, // max time for connections using TCP Keep-Alive
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, eris.Wrap(err, "failed to listen on port")
	}

	return server, &ln, nil
}
