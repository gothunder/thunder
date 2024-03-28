package context

import (
	"context"
	"strings"
	"sync"

	"github.com/samber/lo"
)

const (
	metadataPrepend = "x-thunder-metadata-"
)

type metadataKey struct{}

// Metadata is a map[string]string to maintain a maximum of compatibility
// with broker's metadata implementation
type Metadata struct {
	lock sync.RWMutex
	m    map[string]string
}

func NewMetadataFromMap(m map[string]string) *Metadata {
	metadata := newMetadata()
	metadata.SetMap(m)
	return metadata
}

func NewMetadata() *Metadata {
	return newMetadata()
}

func (*Metadata) buildKey(key string) string {
	var keyBuilder strings.Builder
	keyBuilder.Grow(len(metadataPrepend) + len(key))
	keyBuilder.WriteString(metadataPrepend)
	keyBuilder.WriteString(strings.ToLower(key))

	return keyBuilder.String()
}

func (metadata *Metadata) Get(key string) string {
	metadata.lock.RLock()
	defer metadata.lock.RUnlock()

	value, ok := metadata.m[metadata.buildKey(key)]
	if !ok {
		return ""
	}

	return value
}

func (metadata *Metadata) Set(key, value string) {
	metadata.lock.Lock()
	metadata.m[metadata.buildKey(key)] = value
	metadata.lock.Unlock()
}

func (metadata *Metadata) SetMap(m map[string]string) {
	metadata.lock.Lock()
	for k, v := range m {
		metadata.m[metadata.buildKey(k)] = v
	}
	metadata.lock.Unlock()
}

func (metadata *Metadata) Del(key string) {
	metadata.lock.Lock()
	delete(metadata.m, metadata.buildKey(key))
	metadata.lock.Unlock()
}

func (metadata *Metadata) Keys() []string {
	metadata.lock.RLock()
	defer metadata.lock.RUnlock()
	return lo.Map(lo.Keys(metadata.m), func(k string, _ int) string {
		return strings.TrimPrefix(k, metadataPrepend)
	})
}

// UnmarshalMap parse the metadata from the map and adds it to the metadata
// It expects the metadata keys to be prefixed with "x-thunder-metadata-"
// e.g. "x-thunder-metadata-key"
// All keys not prefixed with "x-thunder-metadata-" will be ignored
func (metadata *Metadata) UnmarshalMap(m map[string]string) {
	metadata.lock.Lock()
	for k, v := range m {
		if strings.HasPrefix(k, metadataPrepend) {
			metadata.m[k] = v
		}
	}
	metadata.lock.Unlock()
}

// MarshalMap writes the representation of a metadata to the given map
// It will write all keys with the "x-thunder-metadata-" prefix
// since this is expected to pair with UnmarshalMap
// e.g. "x-thunder-metadata-key"
func (metadata *Metadata) MarshalMap() map[string]string {
	metadata.lock.RLock()
	defer metadata.lock.RUnlock()

	m := make(map[string]string, len(metadata.m))

	for k, v := range metadata.m {
		m[k] = v
	}

	return m
}

// apply should return a copy to avoid change metadata in parent contexts
// since it is a pointer
func (md *Metadata) apply(metadata *Metadata) *Metadata {
	newMetadata := md.clone()
	metadata.lock.RLock()
	for k, v := range metadata.m {
		newMetadata.m[k] = v
	}
	metadata.lock.RUnlock()
	return newMetadata
}

func (metadata *Metadata) clone() *Metadata {
	metadata.lock.RLock()
	defer metadata.lock.RUnlock()

	metadataMap := make(map[string]string, len(metadata.m))
	for k, v := range metadata.m {
		metadataMap[k] = v
	}

	return makeMetadata(metadataMap)
}

func newMetadata() *Metadata {
	return &Metadata{
		m: make(map[string]string),
	}
}

func makeMetadata(m map[string]string) *Metadata {
	return &Metadata{m: m}
}

// ContextWithMetadata returns a new context with the given metadata
func ContextWithMetadata(ctx context.Context, metadata *Metadata) context.Context {
	currentMetadata := MetadataFromContext(ctx)
	if currentMetadata == nil {
		return context.WithValue(ctx, metadataKey{}, metadata)
	}

	return context.WithValue(ctx, metadataKey{}, currentMetadata.apply(metadata))
}

// ContextWithMetadata returns a new context with the given metadata, replacing any existing metadata
func ContextReplaceMetadata(ctx context.Context, metadata *Metadata) context.Context {
	return context.WithValue(ctx, metadataKey{}, metadata)
}

// MetadataFromContext returns the metadata from the given context
func MetadataFromContext(ctx context.Context) *Metadata {
	if md, ok := ctx.Value(metadataKey{}).(*Metadata); ok {
		return md
	}
	return newMetadata()
}
