package graphql

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/transport"
)

// SSEWithWriteDeadline wraps the built-in SSE transport and extends
// the connection's write deadline before delegating. Without this,
// the server's default WriteTimeout (10 s) kills long-running streams
// such as LLM token-by-token responses.
type SSEWithWriteDeadline struct {
	WriteDeadline time.Duration // deadline per SSE connection, default 5 min
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
	_ = rc.SetWriteDeadline(time.Now().Add(deadline))

	transport.SSE{}.Do(w, r, exec)
}
