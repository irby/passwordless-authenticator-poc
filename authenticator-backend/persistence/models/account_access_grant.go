package models

import (
	"github.com/gobuffalo/pop/v6"
	_ "github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	_ "github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	_ "github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type AccountAccessGrant struct {
	ID        uuid.UUID `db:"id"`
	UserId    uuid.UUID `db:"user_id"`
	Ttl       int       `db:"ttl"` // in seconds
	Token     string    `db:"code"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (grant *AccountAccessGrant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: grant.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: grant.UserId},
		&validators.StringLengthInRange{Name: "Code", Field: grant.Token, Min: 6},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: grant.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: grant.UpdatedAt},
	), nil
}
