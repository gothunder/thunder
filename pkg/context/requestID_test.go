package context

import (
	"context"
	"testing"
)

func TestContextWithRequestIDAndRequestIDFromContext(t *testing.T) {
	t.Parallel()
	t.Run("should return the request ID from the context", func(t *testing.T) {
		t.Parallel()
		ctx := ContextWithRequestID(context.Background(), "request-id")
		if RequestIDFromContext(ctx) != "request-id" {
			t.Error("should return the request ID from the context")
		}
	})
	t.Run("should return empty when not set", func(t *testing.T) {
		t.Parallel()
		if RequestIDFromContext(context.Background()) != "" {
			t.Error("should return an empty string if the request ID is not in the context")
		}
	})
}
