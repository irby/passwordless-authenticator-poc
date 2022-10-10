package models

import (
	"database/sql"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type UserGuestRelation struct {
	ID              uuid.UUID     `db:"id"`
	GuestUserID     uuid.UUID     `db:"guestUserId"`
	ParentUserID    uuid.UUID     `db:"parentUserId"`
	CreatedAt       time.Time     `db:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at"`
	IsActive        bool          `db:"is_active"`
	LoginsRemaining sql.NullInt32 `db:"login_count"`
	ExpireTime      sql.NullTime  `db:"expireTime"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (relation *UserGuestRelation) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: relation.ID},
		&validators.UUIDIsPresent{Name: "GuestUserID", Field: relation.GuestUserID},
		&validators.UUIDIsPresent{Name: "ParentUserID", Field: relation.ParentUserID},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: relation.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: relation.UpdatedAt},
	), nil
}
