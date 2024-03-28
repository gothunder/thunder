package context

import (
	"context"
	"testing"
)

func TestMetadata(t *testing.T) {
	t.Parallel()
	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		t.Run("should return the value of the key", func(t *testing.T) {
			t.Parallel()
			metadata := &Metadata{
				m: map[string]string{
					"x-thunder-metadata-key": "value",
				},
			}
			if metadata.Get("key") != "value" {
				t.Error("Get should return the value of the key")
			}
		})
		t.Run("should return an empty string if the key does not exist", func(t *testing.T) {
			t.Parallel()
			metadata := &Metadata{}
			if metadata.Get("key") != "" {
				t.Error("Get should return an empty string if the key does not exist")
			}
		})
	})
	t.Run("Set", func(t *testing.T) {
		t.Parallel()
		t.Run("should set the value of the key", func(t *testing.T) {
			t.Parallel()
			metadata := &Metadata{
				m: map[string]string{},
			}
			metadata.Set("key", "value")
			if metadata.Get("key") != "value" {
				t.Error("Set should set the value of the key")
			}
		})
	})
	t.Run("Del", func(t *testing.T) {
		t.Parallel()
		t.Run("should delete the key", func(t *testing.T) {
			t.Parallel()
			metadata := &Metadata{
				m: map[string]string{
					"x-thunder-metadata-key": "value",
				},
			}
			metadata.Del("key")
			if metadata.Get("key") != "" {
				t.Error("Del should delete the key")
			}
		})
	})
	t.Run("Keys", func(t *testing.T) {
		t.Parallel()
		t.Run("should return the keys", func(t *testing.T) {
			t.Parallel()
			metadata := &Metadata{
				m: map[string]string{
					"x-thunder-metadata-key": "value",
				},
			}
			if len(metadata.Keys()) != 1 || metadata.Keys()[0] != "key" {
				t.Error("Keys should return the keys")
			}
		})
	})

	t.Run("SetMap", func(t *testing.T) {
		t.Parallel()
		t.Run("should set the map", func(t *testing.T) {
			t.Parallel()
			metadata := &Metadata{
				m: map[string]string{},
			}
			metadata.SetMap(map[string]string{
				"key": "value",
			})
			if metadata.Get("key") != "value" {
				t.Error("SetMap should set the map")
			}
		})
	})

	t.Run("MarshalMap", func(t *testing.T) {
		t.Parallel()
		t.Run("should return the map", func(t *testing.T) {
			t.Parallel()
			metadata := &Metadata{
				m: map[string]string{
					"x-thunder-metadata-key": "value",
				},
			}

			stringMap := metadata.MarshalMap()
			if len(stringMap) != 1 || stringMap["x-thunder-metadata-key"] != "value" {
				t.Error("MarshalMap should return the map")
			}
		})
	})

	t.Run("UnmarshalMap", func(t *testing.T) {
		t.Parallel()
		t.Run("should set the map", func(t *testing.T) {
			t.Parallel()
			metadata := &Metadata{
				m: map[string]string{},
			}
			stringMap := map[string]string{
				"x-thunder-metadata-key": "value",
				"x-thunder-topic":        "some-topic",
				"x-delivery-count":       "1",
			}

			metadata.UnmarshalMap(stringMap)

			if metadata.Get("key") != "value" {
				t.Error("UnmarshalMap should set the map")
			}

			if len(metadata.Keys()) > 1 {
				t.Error("UnmarshalMap should set the map")
			}
		})
	})
}

func TestContextWithMetadata(t *testing.T) {
	t.Parallel()
	t.Run("should return a new context with the given metadata", func(t *testing.T) {
		t.Parallel()
		ctx := ContextWithMetadata(context.Background(), &Metadata{
			m: map[string]string{
				"x-thunder-metadata-key": "value",
			},
		})
		if ctx.Value(metadataKey{}).(*Metadata).m["x-thunder-metadata-key"] != "value" {
			t.Error("ContextWithMetadata should return a new context with the given metadata")
		}

		ctx = ContextWithMetadata(ctx, &Metadata{
			m: map[string]string{
				"x-thunder-metadata-key2": "value2",
			},
		})

		if ctx.Value(metadataKey{}).(*Metadata).m["x-thunder-metadata-key"] != "value" ||
			ctx.Value(metadataKey{}).(*Metadata).m["x-thunder-metadata-key2"] != "value2" {
			t.Error("ContextWithMetadata should return a new context with the given metadata")
		}
	})
}

func TestContextReplaceMetadata(t *testing.T) {
	t.Parallel()
	t.Run("should return a new context with the given metadata, replacing any existing metadata", func(t *testing.T) {
		t.Parallel()
		ctx := ContextReplaceMetadata(context.Background(), &Metadata{
			m: map[string]string{
				"x-thunder-metadata-key": "value",
			},
		})
		if ctx.Value(metadataKey{}).(*Metadata).m["x-thunder-metadata-key"] != "value" {
			t.Error("ContextReplaceMetadata should return a new context with the given metadata, replacing any existing metadata")
		}
	})
}

func TestMetadataFromContext(t *testing.T) {
	t.Parallel()
	t.Run("should return the metadata from the given context", func(t *testing.T) {
		t.Parallel()
		ctx := context.WithValue(context.Background(), metadataKey{}, &Metadata{
			m: map[string]string{
				"x-thunder-metadata-key": "value",
			},
		})
		if MetadataFromContext(ctx).m["x-thunder-metadata-key"] != "value" {
			t.Error("MetadataFromContext should return the metadata from the given context")
		}
	})

}
