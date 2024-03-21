package context

import (
	"context"
	"testing"
)

func TestContextWithCorrelationIDAndCorrelationIDFromContext(t *testing.T) {
	t.Parallel()
	t.Run("should return the correlation ID from the context", func(t *testing.T) {
		t.Parallel()
		ctx := ContextWithCorrelationID(context.Background(), "correlation-id")
		if CorrelationIDFromContext(ctx) != "correlation-id" {
			t.Error("should return the correlation ID from the context")
		}
	})
	t.Run("should generate a new correlation ID when no correlation ID is provided", func(t *testing.T) {
		t.Parallel()
		ctx := ContextWithCorrelationID(context.Background(), "")
		if CorrelationIDFromContext(ctx) == "" {
			t.Error("should return a new context with the given correlation ID")
		}
	})
	t.Run("should generate a new correlation ID if none in the context", func(t *testing.T) {
		t.Parallel()
		if CorrelationIDFromContext(context.Background()) == "" {
			t.Error("should return an empty string if the correlation ID is not in the context")
		}
	})
}
