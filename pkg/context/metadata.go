package context

import (
	"context"

	"github.com/samber/lo"
)

type metadataKey struct{}

// Metadata is a map[string]string to maintain a maximum of compatibility
// with broker's metadata implementation
type Metadata map[string]string

func (m Metadata) Get(key string) string {
	value, ok := m[key]
	if !ok {
		return ""
	}

	return value
}

func (m Metadata) Set(key, value string) {
	m[key] = value
}

func (m Metadata) Del(key string) {
	delete(m, key)
}

func (m Metadata) Keys() []string {
	return lo.Keys(m)
}

func (m Metadata) apply(metadata Metadata) Metadata {
	for k, v := range metadata {
		m[k] = v
	}

	return m
}

// ContextWithMetadata returns a new context with the given metadata
func ContextWithMetadata(ctx context.Context, metadata Metadata) context.Context {
	currentMetadata := MetadataFromContext(ctx)
	if currentMetadata == nil {
		return context.WithValue(ctx, metadataKey{}, metadata)
	}

	return context.WithValue(ctx, metadataKey{}, currentMetadata.apply(metadata))
}

// ContextWithMetadata returns a new context with the given metadata, replacing any existing metadata
func ContextReplaceMetadata(ctx context.Context, metadata Metadata) context.Context {
	return context.WithValue(ctx, metadataKey{}, metadata)
}

// MetadataFromContext returns the metadata from the given context
func MetadataFromContext(ctx context.Context) Metadata {
	if md, ok := ctx.Value(metadataKey{}).(Metadata); ok {
		return md
	}
	return nil
}
