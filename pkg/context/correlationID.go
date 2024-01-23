package context

import (
	"context"

	"github.com/google/uuid"
)

type correlationIDKey struct{}

// CorrelationIDFromContext retrieves the correlation ID from a context.Context.
func CorrelationIDFromContext(ctx context.Context) string {
	correlationID, ok := ctx.Value(correlationIDKey{}).(string)
	if !ok {
		return ""
	}
	return correlationID
}

// ContextWithCorrelationID returns a new context.Context that holds the given correlation ID.
func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		return context.WithValue(ctx, correlationIDKey{}, uuid.NewString())
	}
	return context.WithValue(ctx, correlationIDKey{}, correlationID)
}
