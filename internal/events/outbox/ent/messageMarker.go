package outboxent

import (
	"context"
	"reflect"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/TheRafaBonin/roxy"
	"github.com/google/uuid"
	"github.com/gothunder/thunder/internal/events/outbox"
	"github.com/gothunder/thunder/internal/utils"
)

const (
	// fieldID is the id field for the OutboxMessage entity.
	fieldID = "id"
	// fieldDeliveredAt holds the string denoting the delivered_at field in the database.
	fieldDeliveredAt = "delivered_at"

	// methods
	methodUpdate         = "Update"
	methodSetDeliveredAt = "SetDeliveredAt"
	methodUpdateWhere    = "Where"
	methodUpdateExec     = "Exec"
)

func NewEntMessageMarker(outboxMessageClient interface{}) (outbox.MessageMarker, error) {
	if outboxMessageClient == nil || !utils.HasMethod(outboxMessageClient, "Update") {
		return nil, roxy.Wrap(utils.ErrMethodNotFound, "creating ent message marker")
	}

	return &entMessageMarker{
		client: outboxMessageClient,
	}, nil
}

type entMessageMarker struct {
	client interface{}
}

// MarkAsPublished implements outbox.MessageMarker.
func (e entMessageMarker) MarkAsPublished(ctx context.Context, msgPack []outbox.Message) error {
	ids := make([]uuid.UUID, len(msgPack))
	for i, msg := range msgPack {
		ids[i] = msg.ID
	}

	updateBuilder, err := newUpdateBuilder(e.client)
	if err != nil {
		return roxy.Wrap(err, "creating update builder")
	}

	if err = updateBuilder.SetDeliveredAt(time.Now()); err != nil {
		return roxy.Wrap(err, "setting delivered_at field")
	}

	if err = updateBuilder.WhereDeliveryAtIsNilAndIDIn(ids); err != nil {
		return roxy.Wrap(err, "setting where clause")
	}

	return updateBuilder.Exec(ctx)
}

type outboxMessageUpdateBuilder struct {
	builder interface{}
}

func newUpdateBuilder(client interface{}) (*outboxMessageUpdateBuilder, error) {
	// initialize query builder
	updateBuilder, err := utils.SafeCallMethod(client, methodUpdate, []reflect.Value{})
	if err != nil {
		return nil, roxy.Wrap(err, "calling update method on ent client")
	}

	return &outboxMessageUpdateBuilder{
		builder: updateBuilder[0].Interface(),
	}, nil
}

func (q *outboxMessageUpdateBuilder) SetDeliveredAt(deliveredAt time.Time) error {
	_, err := utils.SafeCallMethod(q.builder, methodSetDeliveredAt, []reflect.Value{
		reflect.ValueOf(deliveredAt),
	})
	if err != nil {
		return roxy.Wrap(err, "calling SetDeliveredAt method on OutboxMessageUpdate")
	}
	return nil
}

func (q *outboxMessageUpdateBuilder) WhereDeliveryAtIsNilAndIDIn(ids []uuid.UUID) error {
	method, ok := reflect.TypeOf(q.builder).MethodByName(methodUpdateWhere)
	if !ok {
		return roxy.Wrap(utils.ErrMethodNotFound, "getting method Where of OutboxMessageUpdate")
	}

	// Where
	elemType := method.Type.In(1).Elem()
	IDInClause := sql.FieldIn(fieldID, ids...)
	deliveredAtNilClause := sql.FieldIsNull(fieldDeliveredAt)
	whereIDInAndDeliveredAtNil := reflect.ValueOf(
		sql.AndPredicates(IDInClause, deliveredAtNilClause),
	).Convert(elemType)

	_, err := utils.SafeCallMethod(q.builder, methodUpdateWhere, []reflect.Value{
		whereIDInAndDeliveredAtNil,
	})
	if err != nil {
		return roxy.Wrap(err, "calling Where method on OutboxMessageUpdate")
	}

	return nil
}

func (q *outboxMessageUpdateBuilder) Exec(ctx context.Context) error {
	result, err := utils.SafeCallMethod(q.builder, methodUpdateExec, []reflect.Value{
		reflect.ValueOf(ctx),
	})
	if err != nil {
		return roxy.Wrap(err, "calling Exec method on OutboxMessageUpdate")
	}

	if result[0].Interface() != nil {
		return result[0].Interface().(error)
	}

	return nil
}
