package context

import (
	"context"
)

const ThunderIDMetadataKey = "x-thunder-id"

// MessageIDFromContext retrieves the message ID from a context.Context.
func MessageIDFromContext(ctx context.Context) string {
	m := MetadataFromContext(ctx)
	return m.Get(ThunderIDMetadataKey)
}

// ContextWithMessageID returns a new context.Context that holds the given message ID.
func ContextWithMessageID(ctx context.Context, messageID string) context.Context {
	md := make(Metadata, 1)
	md.Set(ThunderIDMetadataKey, messageID)
	return ContextWithMetadata(ctx, md)
}
