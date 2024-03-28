package context

import (
	"context"

	"github.com/google/uuid"
)

const ThunderCorrelationIDMetadataKey = "x-thunder-correlation-id"

// CorrelationIDFromContext retrieves the correlation ID from a context.Context.
func CorrelationIDFromContext(ctx context.Context) string {
	md := MetadataFromContext(ctx)

	return md.Get(ThunderCorrelationIDMetadataKey)
}

// ContextWithCorrelationID returns a new context.Context that holds the given correlation ID.
func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	md := newMetadata()
	if correlationID == "" {
		md.Set(ThunderCorrelationIDMetadataKey, uuid.Must(uuid.NewV7()).String())
	} else {
		md.Set(ThunderCorrelationIDMetadataKey, correlationID)
	}
	return ContextWithMetadata(ctx, md)
}
