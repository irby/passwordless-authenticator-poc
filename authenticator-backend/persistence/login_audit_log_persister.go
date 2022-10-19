package persistence

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

type LoginAuditLogPersister interface {
	Create(log models.LoginAuditLog) error
	GetByPrimaryUserId(uuid uuid.UUID) ([]models.LoginAuditLog, error)
	GetByGuestUserId(uuid uuid.UUID) ([]models.LoginAuditLog, error)
	GetByGuestUserIdAndGrantId(guestUserId uuid.UUID, grantId uuid.UUID) ([]models.LoginAuditLog, error)
}

type loginAuditLogPersister struct {
	db *pop.Connection
}

func NewLoginAuditLogPersister(db *pop.Connection) LoginAuditLogPersister {
	return &loginAuditLogPersister{db: db}
}

func (p *loginAuditLogPersister) Create(log models.LoginAuditLog) error {
	if log.ID == uuid.Nil {
		uuId, _ := uuid.NewV4()
		log.ID = uuId
	}
	log.CreatedAt = time.Now().UTC()
	log.UpdatedAt = time.Now().UTC()
	vErr, err := p.db.ValidateAndCreate(&log)
	if err != nil {
		return fmt.Errorf("failed to store user audit login: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("accessGrant object validation failed: %w", vErr)
	}

	return nil
}

func (p *loginAuditLogPersister) GetByPrimaryUserId(uuid uuid.UUID) ([]models.LoginAuditLog, error) {
	var models []models.LoginAuditLog
	conn := p.db.RawQuery("select * from login_audit_logs WHERE user_id = ?", uuid)
	err := conn.All(&models)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve login audits by primary user id: %w", err)
	}
	return models, nil
}

func (p *loginAuditLogPersister) GetByGuestUserId(uuid uuid.UUID) ([]models.LoginAuditLog, error) {
	var models []models.LoginAuditLog
	conn := p.db.RawQuery("select * from login_audit_logs WHERE surrogate_user_id = ?", uuid)
	err := conn.All(&models)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve login audits by guest user id: %w", err)
	}
	return models, nil
}

func (p *loginAuditLogPersister) GetByGuestUserIdAndGrantId(guestUserId uuid.UUID, grantId uuid.UUID) ([]models.LoginAuditLog, error) {
	var models []models.LoginAuditLog
	conn := p.db.RawQuery("select * from login_audit_logs WHERE surrogate_user_id = ? AND user_guest_relation_id = ?", guestUserId, grantId)
	err := conn.All(&models)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve login audits by primary user id: %w", err)
	}
	return models, nil
}
