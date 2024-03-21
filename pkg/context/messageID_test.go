package context

import (
	"context"
	"testing"
)

func TestMessageIDFromContextAndContextWithMessageID(t *testing.T) {
	t.Parallel()
	t.Run("should return the message ID from the context", func(t *testing.T) {
		t.Parallel()
		ctx := ContextWithMessageID(context.Background(), "message-id")
		if MessageIDFromContext(ctx) != "message-id" {
			t.Error("should return the message ID from the context")
		}
	})
	t.Run("should return an empty string if the message ID is not in the context", func(t *testing.T) {
		t.Parallel()
		if MessageIDFromContext(context.Background()) != "" {
			t.Error("should return an empty string if the message ID is not in the context")
		}
	})
}
