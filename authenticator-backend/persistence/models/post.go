package models

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"time"
)

type Post struct {
	ID                   uuid.UUID `db:"id" json:"id"`
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
	CreatedByUserId      uuid.UUID `db:"created_by_user_id" json:"created_by_user_id"`
	CreatedBySurrogateId uuid.UUID `db:"created_by_surrogate_id" json:"created_by_surrogate_id"`
	UpdatedByUserId      uuid.UUID `db:"updated_by_user_id" json:"updated_by_user_id"`
	UpdatedBySurrogateId uuid.UUID `db:"updated_by_surrogate_id" json:"updated_by_surrogate_id"`
	Data                 string    `db:"data" json:"data"`
	IsActive             bool      `db:"is_active" json:"is_active"`
}

func (post *Post) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: post.ID},
		&validators.StringIsPresent{Name: "Data", Field: post.Data},
		&validators.TimeIsPresent{Name: "UpdatedAt", Field: post.UpdatedAt},
		&validators.TimeIsPresent{Name: "CreatedAt", Field: post.CreatedAt},
		&validators.UUIDIsPresent{Name: "CreatedByUserId", Field: post.CreatedByUserId},
		&validators.UUIDIsPresent{Name: "CreatedBySurrogateId", Field: post.CreatedBySurrogateId},
		&validators.UUIDIsPresent{Name: "UpdatedByUserId", Field: post.UpdatedByUserId},
		&validators.UUIDIsPresent{Name: "UpdatedBySurrogateId", Field: post.UpdatedBySurrogateId},
	), nil
}
