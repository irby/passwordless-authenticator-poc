package persistence

import (
	"database/sql"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebauthnCredentialsPrivateKeyPersister interface {
	Get(string) (*models.WebauthnCredentialsPrivateKey, error)
	Create(models.WebauthnCredentialsPrivateKey) error
}

type webauthnCredentialsPrivateKeyPersister struct {
	db *pop.Connection
}

func NewWebauthnCredentialsPrivateKeyPersister(db *pop.Connection) WebauthnCredentialsPrivateKeyPersister {
	return &webauthnCredentialsPrivateKeyPersister{db: db}
}

func (p *webauthnCredentialsPrivateKeyPersister) Get(id string) (*models.WebauthnCredentialsPrivateKey, error) {
	credential := models.WebauthnCredentialsPrivateKey{}
	err := p.db.Find(&credential, id)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	return &credential, nil
}

// Create stores a new `WebauthnCredentialsPrivateKey`. Please run inside a transaction, since `Transports` associated with the
// credential are stored separately in another table.
func (p *webauthnCredentialsPrivateKeyPersister) Create(credential models.WebauthnCredentialsPrivateKey) error {
	vErr, err := p.db.ValidateAndCreate(&credential)
	if err != nil {
		return fmt.Errorf("failed to store credential private key: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("credential object validation failed: %w", vErr)
	}

	return nil
}
