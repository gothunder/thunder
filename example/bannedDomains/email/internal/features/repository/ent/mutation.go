// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gothunder/thunder/example/email/internal/features/repository/ent/email"
	"github.com/gothunder/thunder/example/email/internal/features/repository/ent/predicate"

	"entgo.io/ent"
)

const (
	// Operation types.
	OpCreate    = ent.OpCreate
	OpDelete    = ent.OpDelete
	OpDeleteOne = ent.OpDeleteOne
	OpUpdate    = ent.OpUpdate
	OpUpdateOne = ent.OpUpdateOne

	// Node types.
	TypeEmail = "Email"
)

// EmailMutation represents an operation that mutates the Email nodes in the graph.
type EmailMutation struct {
	config
	op            Op
	typ           string
	id            *int
	email         *string
	createdAt     *time.Time
	updatedAt     *time.Time
	clearedFields map[string]struct{}
	done          bool
	oldValue      func(context.Context) (*Email, error)
	predicates    []predicate.Email
}

var _ ent.Mutation = (*EmailMutation)(nil)

// emailOption allows management of the mutation configuration using functional options.
type emailOption func(*EmailMutation)

// newEmailMutation creates new mutation for the Email entity.
func newEmailMutation(c config, op Op, opts ...emailOption) *EmailMutation {
	m := &EmailMutation{
		config:        c,
		op:            op,
		typ:           TypeEmail,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withEmailID sets the ID field of the mutation.
func withEmailID(id int) emailOption {
	return func(m *EmailMutation) {
		var (
			err   error
			once  sync.Once
			value *Email
		)
		m.oldValue = func(ctx context.Context) (*Email, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Email.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withEmail sets the old Email of the mutation.
func withEmail(node *Email) emailOption {
	return func(m *EmailMutation) {
		m.oldValue = func(context.Context) (*Email, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m EmailMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m EmailMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *EmailMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *EmailMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Email.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetEmail sets the "email" field.
func (m *EmailMutation) SetEmail(s string) {
	m.email = &s
}

// Email returns the value of the "email" field in the mutation.
func (m *EmailMutation) Email() (r string, exists bool) {
	v := m.email
	if v == nil {
		return
	}
	return *v, true
}

// OldEmail returns the old "email" field's value of the Email entity.
// If the Email object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *EmailMutation) OldEmail(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldEmail is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldEmail requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldEmail: %w", err)
	}
	return oldValue.Email, nil
}

// ResetEmail resets all changes to the "email" field.
func (m *EmailMutation) ResetEmail() {
	m.email = nil
}

// SetCreatedAt sets the "createdAt" field.
func (m *EmailMutation) SetCreatedAt(t time.Time) {
	m.createdAt = &t
}

// CreatedAt returns the value of the "createdAt" field in the mutation.
func (m *EmailMutation) CreatedAt() (r time.Time, exists bool) {
	v := m.createdAt
	if v == nil {
		return
	}
	return *v, true
}

// OldCreatedAt returns the old "createdAt" field's value of the Email entity.
// If the Email object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *EmailMutation) OldCreatedAt(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldCreatedAt is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldCreatedAt requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldCreatedAt: %w", err)
	}
	return oldValue.CreatedAt, nil
}

// ResetCreatedAt resets all changes to the "createdAt" field.
func (m *EmailMutation) ResetCreatedAt() {
	m.createdAt = nil
}

// SetUpdatedAt sets the "updatedAt" field.
func (m *EmailMutation) SetUpdatedAt(t time.Time) {
	m.updatedAt = &t
}

// UpdatedAt returns the value of the "updatedAt" field in the mutation.
func (m *EmailMutation) UpdatedAt() (r time.Time, exists bool) {
	v := m.updatedAt
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdatedAt returns the old "updatedAt" field's value of the Email entity.
// If the Email object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *EmailMutation) OldUpdatedAt(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldUpdatedAt is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldUpdatedAt requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldUpdatedAt: %w", err)
	}
	return oldValue.UpdatedAt, nil
}

// ResetUpdatedAt resets all changes to the "updatedAt" field.
func (m *EmailMutation) ResetUpdatedAt() {
	m.updatedAt = nil
}

// Where appends a list predicates to the EmailMutation builder.
func (m *EmailMutation) Where(ps ...predicate.Email) {
	m.predicates = append(m.predicates, ps...)
}

// Op returns the operation name.
func (m *EmailMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (Email).
func (m *EmailMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *EmailMutation) Fields() []string {
	fields := make([]string, 0, 3)
	if m.email != nil {
		fields = append(fields, email.FieldEmail)
	}
	if m.createdAt != nil {
		fields = append(fields, email.FieldCreatedAt)
	}
	if m.updatedAt != nil {
		fields = append(fields, email.FieldUpdatedAt)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *EmailMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case email.FieldEmail:
		return m.Email()
	case email.FieldCreatedAt:
		return m.CreatedAt()
	case email.FieldUpdatedAt:
		return m.UpdatedAt()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *EmailMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case email.FieldEmail:
		return m.OldEmail(ctx)
	case email.FieldCreatedAt:
		return m.OldCreatedAt(ctx)
	case email.FieldUpdatedAt:
		return m.OldUpdatedAt(ctx)
	}
	return nil, fmt.Errorf("unknown Email field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *EmailMutation) SetField(name string, value ent.Value) error {
	switch name {
	case email.FieldEmail:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetEmail(v)
		return nil
	case email.FieldCreatedAt:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreatedAt(v)
		return nil
	case email.FieldUpdatedAt:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdatedAt(v)
		return nil
	}
	return fmt.Errorf("unknown Email field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *EmailMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *EmailMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *EmailMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Email numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *EmailMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *EmailMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *EmailMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Email nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *EmailMutation) ResetField(name string) error {
	switch name {
	case email.FieldEmail:
		m.ResetEmail()
		return nil
	case email.FieldCreatedAt:
		m.ResetCreatedAt()
		return nil
	case email.FieldUpdatedAt:
		m.ResetUpdatedAt()
		return nil
	}
	return fmt.Errorf("unknown Email field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *EmailMutation) AddedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *EmailMutation) AddedIDs(name string) []ent.Value {
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *EmailMutation) RemovedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *EmailMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *EmailMutation) ClearedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *EmailMutation) EdgeCleared(name string) bool {
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *EmailMutation) ClearEdge(name string) error {
	return fmt.Errorf("unknown Email unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *EmailMutation) ResetEdge(name string) error {
	return fmt.Errorf("unknown Email edge %s", name)
}
