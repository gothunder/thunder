package schema

import (
	"entgo.io/ent"
	outboxent "github.com/gothunder/thunder/pkg/events/outbox/ent"
)

type OutboxMessage struct {
	ent.Schema
}

func (OutboxMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		outboxent.OutboxMessageMixin{},
	}
}
