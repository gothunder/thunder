// Code generated by ent, DO NOT EDIT.

package entInit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/gothunder/thunder/tests/entInit/outboxmessage"
	"github.com/gothunder/thunder/tests/entInit/predicate"
)

// OutboxMessageUpdate is the builder for updating OutboxMessage entities.
type OutboxMessageUpdate struct {
	config
	hooks    []Hook
	mutation *OutboxMessageMutation
}

// Where appends a list predicates to the OutboxMessageUpdate builder.
func (omu *OutboxMessageUpdate) Where(ps ...predicate.OutboxMessage) *OutboxMessageUpdate {
	omu.mutation.Where(ps...)
	return omu
}

// SetHeaders sets the "headers" field.
func (omu *OutboxMessageUpdate) SetHeaders(m map[string]string) *OutboxMessageUpdate {
	omu.mutation.SetHeaders(m)
	return omu
}

// ClearHeaders clears the value of the "headers" field.
func (omu *OutboxMessageUpdate) ClearHeaders() *OutboxMessageUpdate {
	omu.mutation.ClearHeaders()
	return omu
}

// SetCreatedAt sets the "created_at" field.
func (omu *OutboxMessageUpdate) SetCreatedAt(t time.Time) *OutboxMessageUpdate {
	omu.mutation.SetCreatedAt(t)
	return omu
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (omu *OutboxMessageUpdate) SetNillableCreatedAt(t *time.Time) *OutboxMessageUpdate {
	if t != nil {
		omu.SetCreatedAt(*t)
	}
	return omu
}

// SetDeliveredAt sets the "delivered_at" field.
func (omu *OutboxMessageUpdate) SetDeliveredAt(t time.Time) *OutboxMessageUpdate {
	omu.mutation.SetDeliveredAt(t)
	return omu
}

// SetNillableDeliveredAt sets the "delivered_at" field if the given value is not nil.
func (omu *OutboxMessageUpdate) SetNillableDeliveredAt(t *time.Time) *OutboxMessageUpdate {
	if t != nil {
		omu.SetDeliveredAt(*t)
	}
	return omu
}

// ClearDeliveredAt clears the value of the "delivered_at" field.
func (omu *OutboxMessageUpdate) ClearDeliveredAt() *OutboxMessageUpdate {
	omu.mutation.ClearDeliveredAt()
	return omu
}

// Mutation returns the OutboxMessageMutation object of the builder.
func (omu *OutboxMessageUpdate) Mutation() *OutboxMessageMutation {
	return omu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (omu *OutboxMessageUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, omu.sqlSave, omu.mutation, omu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (omu *OutboxMessageUpdate) SaveX(ctx context.Context) int {
	affected, err := omu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (omu *OutboxMessageUpdate) Exec(ctx context.Context) error {
	_, err := omu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (omu *OutboxMessageUpdate) ExecX(ctx context.Context) {
	if err := omu.Exec(ctx); err != nil {
		panic(err)
	}
}

func (omu *OutboxMessageUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(outboxmessage.Table, outboxmessage.Columns, sqlgraph.NewFieldSpec(outboxmessage.FieldID, field.TypeUUID))
	if ps := omu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := omu.mutation.Headers(); ok {
		_spec.SetField(outboxmessage.FieldHeaders, field.TypeJSON, value)
	}
	if omu.mutation.HeadersCleared() {
		_spec.ClearField(outboxmessage.FieldHeaders, field.TypeJSON)
	}
	if value, ok := omu.mutation.CreatedAt(); ok {
		_spec.SetField(outboxmessage.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := omu.mutation.DeliveredAt(); ok {
		_spec.SetField(outboxmessage.FieldDeliveredAt, field.TypeTime, value)
	}
	if omu.mutation.DeliveredAtCleared() {
		_spec.ClearField(outboxmessage.FieldDeliveredAt, field.TypeTime)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, omu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{outboxmessage.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	omu.mutation.done = true
	return n, nil
}

// OutboxMessageUpdateOne is the builder for updating a single OutboxMessage entity.
type OutboxMessageUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *OutboxMessageMutation
}

// SetHeaders sets the "headers" field.
func (omuo *OutboxMessageUpdateOne) SetHeaders(m map[string]string) *OutboxMessageUpdateOne {
	omuo.mutation.SetHeaders(m)
	return omuo
}

// ClearHeaders clears the value of the "headers" field.
func (omuo *OutboxMessageUpdateOne) ClearHeaders() *OutboxMessageUpdateOne {
	omuo.mutation.ClearHeaders()
	return omuo
}

// SetCreatedAt sets the "created_at" field.
func (omuo *OutboxMessageUpdateOne) SetCreatedAt(t time.Time) *OutboxMessageUpdateOne {
	omuo.mutation.SetCreatedAt(t)
	return omuo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (omuo *OutboxMessageUpdateOne) SetNillableCreatedAt(t *time.Time) *OutboxMessageUpdateOne {
	if t != nil {
		omuo.SetCreatedAt(*t)
	}
	return omuo
}

// SetDeliveredAt sets the "delivered_at" field.
func (omuo *OutboxMessageUpdateOne) SetDeliveredAt(t time.Time) *OutboxMessageUpdateOne {
	omuo.mutation.SetDeliveredAt(t)
	return omuo
}

// SetNillableDeliveredAt sets the "delivered_at" field if the given value is not nil.
func (omuo *OutboxMessageUpdateOne) SetNillableDeliveredAt(t *time.Time) *OutboxMessageUpdateOne {
	if t != nil {
		omuo.SetDeliveredAt(*t)
	}
	return omuo
}

// ClearDeliveredAt clears the value of the "delivered_at" field.
func (omuo *OutboxMessageUpdateOne) ClearDeliveredAt() *OutboxMessageUpdateOne {
	omuo.mutation.ClearDeliveredAt()
	return omuo
}

// Mutation returns the OutboxMessageMutation object of the builder.
func (omuo *OutboxMessageUpdateOne) Mutation() *OutboxMessageMutation {
	return omuo.mutation
}

// Where appends a list predicates to the OutboxMessageUpdate builder.
func (omuo *OutboxMessageUpdateOne) Where(ps ...predicate.OutboxMessage) *OutboxMessageUpdateOne {
	omuo.mutation.Where(ps...)
	return omuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (omuo *OutboxMessageUpdateOne) Select(field string, fields ...string) *OutboxMessageUpdateOne {
	omuo.fields = append([]string{field}, fields...)
	return omuo
}

// Save executes the query and returns the updated OutboxMessage entity.
func (omuo *OutboxMessageUpdateOne) Save(ctx context.Context) (*OutboxMessage, error) {
	return withHooks(ctx, omuo.sqlSave, omuo.mutation, omuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (omuo *OutboxMessageUpdateOne) SaveX(ctx context.Context) *OutboxMessage {
	node, err := omuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (omuo *OutboxMessageUpdateOne) Exec(ctx context.Context) error {
	_, err := omuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (omuo *OutboxMessageUpdateOne) ExecX(ctx context.Context) {
	if err := omuo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (omuo *OutboxMessageUpdateOne) sqlSave(ctx context.Context) (_node *OutboxMessage, err error) {
	_spec := sqlgraph.NewUpdateSpec(outboxmessage.Table, outboxmessage.Columns, sqlgraph.NewFieldSpec(outboxmessage.FieldID, field.TypeUUID))
	id, ok := omuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`entInit: missing "OutboxMessage.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := omuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, outboxmessage.FieldID)
		for _, f := range fields {
			if !outboxmessage.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("entInit: invalid field %q for query", f)}
			}
			if f != outboxmessage.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := omuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := omuo.mutation.Headers(); ok {
		_spec.SetField(outboxmessage.FieldHeaders, field.TypeJSON, value)
	}
	if omuo.mutation.HeadersCleared() {
		_spec.ClearField(outboxmessage.FieldHeaders, field.TypeJSON)
	}
	if value, ok := omuo.mutation.CreatedAt(); ok {
		_spec.SetField(outboxmessage.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := omuo.mutation.DeliveredAt(); ok {
		_spec.SetField(outboxmessage.FieldDeliveredAt, field.TypeTime, value)
	}
	if omuo.mutation.DeliveredAtCleared() {
		_spec.ClearField(outboxmessage.FieldDeliveredAt, field.TypeTime)
	}
	_node = &OutboxMessage{config: omuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, omuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{outboxmessage.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	omuo.mutation.done = true
	return _node, nil
}
