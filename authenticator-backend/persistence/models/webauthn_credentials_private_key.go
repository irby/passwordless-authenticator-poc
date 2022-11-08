package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
)

type WebauthnCredentialsPrivateKey struct {
	ID         string    `db:"id" json:"id"`
	PrivateKey string    `db:"private_key" json:"-"`
	CreatedAt  time.Time `db:"created_at" json:"-"`
	UpdatedAt  time.Time `db:"updated_at" json:"-"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (key *WebauthnCredentialsPrivateKey) Validate(_ *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Name: "ID", Field: key.ID},
		&validators.StringIsPresent{Name: "ID", Field: key.PrivateKey},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: key.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: key.UpdatedAt},
	), nil
}
