package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewUserGuestRelationPersister(init []models.UserGuestRelation) persistence.UserGuestRelationPersister {
	return &userGuestRelationPersister{append([]models.UserGuestRelation{}, init...)}
}

type userGuestRelationPersister struct {
	relations []models.UserGuestRelation
}

func (p *userGuestRelationPersister) Get(id uuid.UUID) (*models.UserGuestRelation, error) {
	var found *models.UserGuestRelation
	for _, data := range p.relations {
		if data.ID == id {
			d := data
			found = &d
		}
	}

	return found, nil
}

func (p *userGuestRelationPersister) Create(model models.UserGuestRelation) error {
	p.relations = append(p.relations, model)
	return nil
}

func (p *userGuestRelationPersister) Update(model models.UserGuestRelation) error {
	for i, data := range p.relations {
		if data.ID == model.ID {
			p.relations[i] = model
		}
	}
	return nil
}

func (p *userGuestRelationPersister) GetByGuestUserId(guestUserId *uuid.UUID) ([]models.UserGuestRelation, error) {
	var results []models.UserGuestRelation
	for _, data := range p.relations {
		if data.GuestUserID == *guestUserId {
			results = append(p.relations, data)
		}
	}
	return results, nil
}

func (p *userGuestRelationPersister) GetByParentUserId(parentUserId *uuid.UUID) ([]models.UserGuestRelation, error) {
	var results []models.UserGuestRelation
	for _, data := range p.relations {
		if data.ParentUserID == *parentUserId {
			results = append(p.relations, data)
		}
	}
	return results, nil
}
