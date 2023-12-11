package outbox

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type OutboxMessageMixin struct {
	mixin.Schema
}

func (e OutboxMessageMixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.String("topic").NotEmpty().Immutable(),
		field.Bytes("payload").NotEmpty().Immutable(), // bytes to avoid encoding/decoding with arbitrary shapes
		field.JSON("headers", map[string]string{}).Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("delivered_at").Optional(),
	}
}

func (e OutboxMessageMixin) Edges() []ent.Edge {
	return nil
}

func (e OutboxMessageMixin) Hooks() []ent.Hook {
	return nil
}

func (e OutboxMessageMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at", "delivered_at", "topic"),
	}
}
