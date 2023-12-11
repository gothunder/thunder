package schema

import (
	"entgo.io/ent"
	"github.com/gothunder/thunder/pkg/events/outbox"
)

type OutboxMessage struct {
	ent.Schema
}

func (OutboxMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		outbox.OutboxMessageMixin{},
	}
}
