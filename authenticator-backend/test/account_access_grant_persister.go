package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewAccountAccessGrantPersister(init []models.AccountAccessGrant) persistence.AccountAccessGrantPersister {
	return &accessGrantPersister{append([]models.AccountAccessGrant{}, init...)}
}

type accessGrantPersister struct {
	grants []models.AccountAccessGrant
}

func (p *accessGrantPersister) Get(id uuid.UUID) (*models.AccountAccessGrant, error) {
	var found *models.AccountAccessGrant
	for _, data := range p.grants {
		if data.ID == id {
			d := data
			found = &d
		}
	}

	return found, nil
}

func (p *accessGrantPersister) Create(accessGrant models.AccountAccessGrant) error {
	p.grants = append(p.grants, accessGrant)
	return nil
}

func (p *accessGrantPersister) Update(accessGrant models.AccountAccessGrant) error {
	for i, data := range p.grants {
		if data.ID == accessGrant.ID {
			p.grants[i] = accessGrant
		}
	}
	return nil
}
