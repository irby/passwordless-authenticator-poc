package models

import (
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

type LoginAuditLog struct {
	ID                  uuid.UUID  `db:"id"`
	UserId              uuid.UUID  `db:"user_id"`
	SurrogateUserId     *uuid.UUID `db:"surrogate_user_id"`
	UserGuestRelationId *uuid.UUID `db:"user_guest_relation_id"`
	ClientIpAddress     string     `db:"client_ip_address"`
	ClientUserAgent     string     `db:"client_user_agent"`
	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
	LoginMethod         int        `db:"login_method"`
}

func (log *LoginAuditLog) Validate(_ *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: log.ID},
		&validators.UUIDIsPresent{Name: "UserId", Field: log.UserId},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: log.CreatedAt},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: log.UpdatedAt},
		&validators.StringIsPresent{Name: "ClientIp", Field: log.ClientIpAddress},
		&validators.StringIsPresent{Name: "ClientUserAgent", Field: log.ClientUserAgent},
		&validators.IntIsPresent{Name: "LoginMethod", Field: log.LoginMethod},
		&validators.FuncValidator{Name: "SurrogateUserIdAndUserGuestRelation", Fn: IsSurrogateAndGuestPopulated(log)},
	), nil
}

func IsSurrogateAndGuestPopulated(log *LoginAuditLog) func() bool {
	return func() bool {
		if log.SurrogateUserId == nil {
			return log.UserGuestRelationId == nil
		}
		if log.UserGuestRelationId == nil {
			return log.SurrogateUserId == nil
		}
		return log.UserGuestRelationId != nil && log.SurrogateUserId != nil
	}
}
