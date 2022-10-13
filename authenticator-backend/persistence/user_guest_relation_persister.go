package persistence

import (
	"database/sql"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type UserGuestRelationPersister interface {
	Get(uuid uuid.UUID) (*models.UserGuestRelation, error)
	Create(model models.UserGuestRelation) error
	Update(model models.UserGuestRelation) error
	GetByGuestUserId(guestUserId *uuid.UUID) ([]models.UserGuestRelation, error)
	GetByParentUserId(parentUserId *uuid.UUID) ([]models.UserGuestRelation, error)
}

type userGuestRelationPersister struct {
	db *pop.Connection
}

func NewUserGuestRelationPersister(db *pop.Connection) UserGuestRelationPersister {
	return &userGuestRelationPersister{db: db}
}

func (p *userGuestRelationPersister) Get(id uuid.UUID) (*models.UserGuestRelation, error) {
	model := models.UserGuestRelation{}
	err := p.db.Find(&model, id)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user guest relation: %w", err)
	}
	return &model, nil
}

func (p *userGuestRelationPersister) Create(model models.UserGuestRelation) error {
	vErr, err := p.db.ValidateAndCreate(&model)
	if err != nil {
		return fmt.Errorf("failed to store user guest relation: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("user guest relation object validation failed: %w", vErr)
	}

	return nil
}

func (p *userGuestRelationPersister) Update(model models.UserGuestRelation) error {
	vErr, err := p.db.ValidateAndUpdate(&model)
	if err != nil {
		return fmt.Errorf("failed to update user guest relation: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("user guest relation object validation failed: %w", vErr)
	}

	return nil
}

func (p *userGuestRelationPersister) GetByGuestUserId(guestUserId *uuid.UUID) ([]models.UserGuestRelation, error) {
	models := []models.UserGuestRelation{}
	conn := p.db.RawQuery("select * from user_guest_relations where guest_user_id = ? AND is_active = true", guestUserId)
	err := conn.All(&models)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user guest relations by guest id: %w", err)
	}
	return models, nil
}

func (p *userGuestRelationPersister) GetByParentUserId(parentUserId *uuid.UUID) ([]models.UserGuestRelation, error) {
	models := []models.UserGuestRelation{}
	conn := p.db.RawQuery("select * from user_guest_relations where parent_user_id = ? AND is_active = true", &parentUserId)
	err := conn.All(&models)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user guest relations by parent id: %w", err)
	}
	return models, nil
}
