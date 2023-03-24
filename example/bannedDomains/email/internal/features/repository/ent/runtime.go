// Code generated by ent, DO NOT EDIT.

package ent

import (
	"time"

	"github.com/gothunder/thunder/example/email/internal/features/repository/ent/email"
	"github.com/gothunder/thunder/example/email/internal/features/repository/ent/schema"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	emailFields := schema.Email{}.Fields()
	_ = emailFields
	// emailDescCreatedAt is the schema descriptor for createdAt field.
	emailDescCreatedAt := emailFields[1].Descriptor()
	// email.DefaultCreatedAt holds the default value on creation for the createdAt field.
	email.DefaultCreatedAt = emailDescCreatedAt.Default.(func() time.Time)
	// emailDescUpdatedAt is the schema descriptor for updatedAt field.
	emailDescUpdatedAt := emailFields[2].Descriptor()
	// email.DefaultUpdatedAt holds the default value on creation for the updatedAt field.
	email.DefaultUpdatedAt = emailDescUpdatedAt.Default.(func() time.Time)
	// email.UpdateDefaultUpdatedAt holds the default value on update for the updatedAt field.
	email.UpdateDefaultUpdatedAt = emailDescUpdatedAt.UpdateDefault.(func() time.Time)
}
