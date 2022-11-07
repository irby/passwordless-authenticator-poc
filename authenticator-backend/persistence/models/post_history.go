package models

import (
	"github.com/gofrs/uuid"
	"time"
)

type PostHistory struct {
	ID                   uuid.UUID `db:"id" json:"id"`
	PostId               uuid.UUID `db:"post_id" json:"post_id"`
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
	CreatedByUserId      uuid.UUID `db:"created_by_user_id" json:"created_by_user_id"`
	CreatedBySurrogateId uuid.UUID `db:"created_by_surrogate_id" json:"created_by_surrogate_id"`
	UpdatedByUserId      uuid.UUID `db:"updated_by_user_id" json:"updated_by_user_id"`
	UpdatedBySurrogateId uuid.UUID `db:"updated_by_surrogate_id" json:"updated_by_surrogate_id"`
	Data                 string    `db:"data" json:"data"`
}
