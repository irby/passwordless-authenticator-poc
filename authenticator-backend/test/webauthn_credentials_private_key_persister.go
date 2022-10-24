package test

import (
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewWebauthnCredentialsPrivateKeyPersister(init []models.WebauthnCredentialsPrivateKey) persistence.WebauthnCredentialsPrivateKeyPersister {
	return &webauthnCredentialsPrivateKeyPersister{append([]models.WebauthnCredentialsPrivateKey{}, init...)}
}

type webauthnCredentialsPrivateKeyPersister struct {
	credentials []models.WebauthnCredentialsPrivateKey
}

func (p *webauthnCredentialsPrivateKeyPersister) Get(id string) (*models.WebauthnCredentialsPrivateKey, error) {
	var found *models.WebauthnCredentialsPrivateKey
	for _, data := range p.credentials {
		if data.ID == id {
			d := data
			found = &d
		}
	}
	return found, nil
}

func (p *webauthnCredentialsPrivateKeyPersister) Create(credential models.WebauthnCredentialsPrivateKey) error {
	p.credentials = append(p.credentials, credential)
	return nil
}
