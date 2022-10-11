package models

import (
	"database/sql"
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
	ID             uuid.UUID     `db:"id"`
	UserId         uuid.UUID     `db:"user_id"`
	Ttl            int           `db:"ttl"` // in seconds
	Token          string        `db:"code"`
	IsActive       bool          `db:"is_active"`
	CreatedAt      time.Time     `db:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at"`
	ClaimedBy      *uuid.UUID    `db:"claimed_by"`
	ExpireByLogins bool          `db:"expire_by_logins"`
	LoginsAllowed  sql.NullInt32 `db:"logins_allowed"`
	ExpireByTime   bool          `db:"expire_by_time"`
	MinutesAllowed sql.NullInt32 `db:"minutes_allowed"`
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (grant *AccountAccessGrant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: grant.ID},
		&validators.UUIDIsPresent{Name: "UserID", Field: grant.UserId},
		&validators.StringLengthInRange{Name: "Code", Field: grant.Token, Min: 6},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: grant.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: grant.UpdatedAt},
		&validators.FuncValidator{Name: "LoginsAllowed", Fn: IsLoginsAllowedPopulated(grant)},
		&validators.FuncValidator{Name: "MinutesAllowed", Fn: IsMinutesAllowedPopulated(grant)},
		&validators.FuncValidator{Name: "MutualExclusiveness", Fn: LoginsAndMinutesMutuallyExclusive(grant)},
	), nil
}

func IsLoginsAllowedPopulated(grant *AccountAccessGrant) func() bool {
	return func() bool {
		if !grant.ExpireByLogins {
			return true
		}

		return grant.LoginsAllowed.Int32 > 0 && grant.LoginsAllowed.Valid
	}
}

func IsMinutesAllowedPopulated(grant *AccountAccessGrant) func() bool {
	return func() bool {
		if !grant.ExpireByTime {
			return true
		}

		return grant.MinutesAllowed.Int32 > 0 && grant.MinutesAllowed.Valid
	}
}

func LoginsAndMinutesMutuallyExclusive(grant *AccountAccessGrant) func() bool {
	return func() bool {
		return !(grant.ExpireByTime && grant.ExpireByLogins) && !(grant.MinutesAllowed.Valid && grant.LoginsAllowed.Valid)
	}
}
