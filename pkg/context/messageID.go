package context

import "context"

type messageIDKey struct{}

// MessageIDFromContext retrieves the message ID from a context.Context.
func MessageIDFromContext(ctx context.Context) string {
	messageID, ok := ctx.Value(messageIDKey{}).(string)
	if !ok {
		return ""
	}
	return messageID
}

// ContextWithMessageID returns a new context.Context that holds the given message ID.
func ContextWithMessageID(ctx context.Context, messageID string) context.Context {
	return context.WithValue(ctx, messageIDKey{}, messageID)
}
