package persistence

import (
	"database/sql"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type AccountAccessGrantPersister interface {
	Get(uuid uuid.UUID) (*models.AccountAccessGrant, error)
	Create(grant models.AccountAccessGrant) error
	Update(grant models.AccountAccessGrant) error
}

type accessGrantPersister struct {
	db *pop.Connection
}

func NewAccountAccessGrantPersister(db *pop.Connection) AccountAccessGrantPersister {
	return &accessGrantPersister{db: db}
}

func (p *accessGrantPersister) Get(id uuid.UUID) (*models.AccountAccessGrant, error) {
	accessGrant := models.AccountAccessGrant{}
	err := p.db.Find(&accessGrant, id)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get access grant: %w", err)
	}

	return &accessGrant, nil
}

func (p *accessGrantPersister) Create(accessGrant models.AccountAccessGrant) error {
	vErr, err := p.db.ValidateAndCreate(&accessGrant)
	if err != nil {
		return fmt.Errorf("failed to store accessGrant: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("accessGrant object validation failed: %w", vErr)
	}

	return nil
}

func (p *accessGrantPersister) Update(accessGrant models.AccountAccessGrant) error {
	vErr, err := p.db.ValidateAndUpdate(&accessGrant)
	if err != nil {
		return fmt.Errorf("failed to update accessGrant: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("accessGrant object validation failed: %w", vErr)
	}

	return nil
}
