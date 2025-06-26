package context

import (
	"context"
)

const RequestIDMetadataKey = "x-request-id"

// RequestIDFromContext retrieves the request ID from a context.Context.
func RequestIDFromContext(ctx context.Context) string {
	m := MetadataFromContext(ctx)
	return m.Get(RequestIDMetadataKey)
}

// ContextWithRequestID returns a new context.Context that holds the given request ID.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	md := newMetadata()
	md.Set(RequestIDMetadataKey, requestID)
	return ContextWithMetadata(ctx, md)
}
