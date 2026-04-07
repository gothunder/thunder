package graphql

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deadlineRecorder wraps httptest.ResponseRecorder and records calls to
// SetWriteDeadline so http.NewResponseController can find it via the
// rwUnwrapper / deadline interface.
type deadlineRecorder struct {
	*httptest.ResponseRecorder
	deadline    time.Time
	deadlineSet bool
}

func (d *deadlineRecorder) SetWriteDeadline(t time.Time) error {
	d.deadline = t
	d.deadlineSet = true
	return nil
}

func (d *deadlineRecorder) Flush() {}

func TestSSEWithWriteDeadline_Supports(t *testing.T) {
	tr := SSEWithWriteDeadline{}

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")
	assert.True(t, tr.Supports(req), "should support SSE requests")

	req2 := httptest.NewRequest(http.MethodPost, "/query", nil)
	req2.Header.Set("Accept", "application/json")
	req2.Header.Set("Content-Type", "application/json")
	assert.False(t, tr.Supports(req2), "should not support non-SSE requests")
}

func TestSSEWithWriteDeadline_Do_SetsDeadline(t *testing.T) {
	tr := SSEWithWriteDeadline{WriteDeadline: 2 * time.Minute}

	rec := &deadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")

	before := time.Now()
	// exec is nil — gqlgen will try to run the request and likely write an
	// error response. We only care that SetWriteDeadline was invoked before
	// delegating, so swallow any panic from the inner transport.
	func() {
		defer func() { _ = recover() }()
		tr.Do(rec, req, nil)
	}()

	require.True(t, rec.deadlineSet, "SetWriteDeadline should have been called")
	delta := rec.deadline.Sub(before)
	assert.GreaterOrEqual(t, delta, 2*time.Minute-time.Second)
	assert.LessOrEqual(t, delta, 2*time.Minute+5*time.Second)
}

func TestSSEWithWriteDeadline_Do_DefaultDeadline(t *testing.T) {
	tr := SSEWithWriteDeadline{}

	rec := &deadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")

	before := time.Now()
	func() {
		defer func() { _ = recover() }()
		tr.Do(rec, req, nil)
	}()

	require.True(t, rec.deadlineSet)
	delta := rec.deadline.Sub(before)
	assert.GreaterOrEqual(t, delta, 5*time.Minute-time.Second, "default should be 5min")
}
