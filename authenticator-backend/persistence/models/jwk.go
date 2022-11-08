package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
)

type Jwk struct {
	ID        int       `db:"id"`
	KeyData   string    `db:"key_data"`
	CreatedAt time.Time `db:"created_at"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (jwk *Jwk) Validate(_ *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Name: "KeyData", Field: jwk.KeyData},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: jwk.CreatedAt},
	), nil
}
