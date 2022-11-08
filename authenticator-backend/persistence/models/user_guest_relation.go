package models

import (
	"database/sql"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

type UserGuestRelation struct {
	ID                      uuid.UUID     `db:"id" json:"id"`
	GuestUserID             uuid.UUID     `db:"guest_user_id" json:"guestUserId"`
	ParentUserID            uuid.UUID     `db:"parent_user_id" json:"parentUserId"`
	CreatedAt               time.Time     `db:"created_at" json:"createdAt"`
	UpdatedAt               time.Time     `db:"updated_at" json:"updatedAt"`
	IsActive                bool          `db:"is_active" json:"isActive"`
	ExpireByLogins          bool          `db:"expire_by_logins" json:"-"`
	LoginsAllowed           sql.NullInt32 `db:"logins_allowed" json:"-"`
	ExpireByTime            bool          `db:"expire_by_time" json:"-"`
	MinutesAllowed          sql.NullInt32 `db:"minutes_allowed" json:"-"`
	AssociatedAccessGrantId uuid.UUID     `db:"associated_access_grant_id" json:"-"`
	GrantHash               *[]byte       `db:"grant_hash" json:"-"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (relation *UserGuestRelation) Validate(_ *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: relation.ID},
		&validators.UUIDIsPresent{Name: "GuestUserID", Field: relation.GuestUserID},
		&validators.UUIDIsPresent{Name: "ParentUserID", Field: relation.ParentUserID},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: relation.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: relation.UpdatedAt},
		&validators.FuncValidator{Name: "LoginsRemaining", Fn: isLoginsRemainingPopulated(relation)},
		&validators.FuncValidator{Name: "ExpireTime", Fn: isMinutesAllowedPopulated(relation)},
	), nil
}

func isLoginsRemainingPopulated(relation *UserGuestRelation) func() bool {
	return func() bool {
		if !relation.ExpireByLogins {
			return true
		}

		return relation.LoginsAllowed.Int32 > 0 && relation.LoginsAllowed.Valid
	}
}

func isMinutesAllowedPopulated(relation *UserGuestRelation) func() bool {
	return func() bool {
		if !relation.ExpireByTime {
			return true
		}

		return relation.MinutesAllowed.Int32 > 0 && relation.MinutesAllowed.Valid
	}
}
