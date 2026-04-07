package graphql

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/rs/zerolog/log"
)

// SSEWithWriteDeadline wraps the built-in SSE transport and raises
// the connection's absolute write deadline before delegating. Without
// this, the server's default WriteTimeout (10 s) cuts SSE streams short.
//
// Note: this sets an absolute deadline, not a per-write reset. Streams
// are still capped at WriteDeadline (default 5 min) — it raises the
// ceiling, it does not make streaming unbounded.
type SSEWithWriteDeadline struct {
	WriteDeadline time.Duration // absolute deadline per SSE connection, default 5 min
}

var _ graphql.Transport = SSEWithWriteDeadline{}

func (t SSEWithWriteDeadline) Supports(r *http.Request) bool {
	return transport.SSE{}.Supports(r)
}

func (t SSEWithWriteDeadline) Do(w http.ResponseWriter, r *http.Request, exec graphql.GraphExecutor) {
	deadline := t.WriteDeadline
	if deadline == 0 {
		deadline = 5 * time.Minute
	}

	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Now().Add(deadline)); err != nil {
		log.Ctx(r.Context()).Warn().Err(err).
			Msg("graphql sse: failed to set write deadline; falling back to server WriteTimeout")
	}

	transport.SSE{}.Do(w, r, exec)
}
