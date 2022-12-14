package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// User is used by pop to map your users database table to your go code.
type User struct {
	ID                  uuid.UUID            `db:"id" json:"id"`
	Email               string               `db:"email" json:"email"`
	Verified            bool                 `db:"verified" json:"verified"`
	WebauthnCredentials []WebauthnCredential `has_many:"webauthn_credentials" json:"webauthn_credentials,omitempty"`
	CreatedAt           time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time            `db:"updated_at" json:"updated_at"`
	IsActive            bool                 `db:"is_active" json:"is_active"`
	IsAdmin             bool                 `db:"is_admin" json:"is_admin"`
}

func NewUser(email string) User {
	id, _ := uuid.NewV4()
	return User{
		ID:        id,
		Email:     email,
		Verified:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
		IsAdmin:   false,
	}
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (user *User) Validate(_ *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: user.ID},
		&validators.EmailLike{Name: "Email", Field: user.Email},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: user.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: user.CreatedAt},
	), nil
}
