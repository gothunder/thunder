package outboxent

import (
	"context"
	"testing"

	"github.com/gothunder/thunder/internal/events/outbox/ent/entInit"
	"github.com/gothunder/thunder/internal/events/outbox/ent/entInit/enttest"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
)

func setupEnt(t *testing.T) *entInit.Client {
	return enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
}

func populateOutboxMessages(ctx context.Context, client *entInit.Client, messages []entInit.OutboxMessage) error {
	return client.OutboxMessage.MapCreateBulk(messages, func(omc *entInit.OutboxMessageCreate, i int) {
		omc.SetID(messages[i].ID)
		omc.SetPayload(messages[i].Payload)
		omc.SetTopic(messages[i].Topic)
		omc.SetCreatedAt(messages[i].CreatedAt)
		if !messages[i].DeliveredAt.IsZero() {
			omc.SetDeliveredAt(messages[i].DeliveredAt)
		}
	}).Exec(ctx)
}
