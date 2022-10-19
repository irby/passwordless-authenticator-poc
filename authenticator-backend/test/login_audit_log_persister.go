package test

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewLoginAuditLogPersister(init []models.LoginAuditLog) persistence.LoginAuditLogPersister {
	return &loginAuditLogPersister{append([]models.LoginAuditLog{}, init...)}
}

type loginAuditLogPersister struct {
	logs []models.LoginAuditLog
}

func (p *loginAuditLogPersister) Create(log models.LoginAuditLog) error {
	p.logs = append(p.logs, log)
	return nil
}

func (p *loginAuditLogPersister) GetByPrimaryUserId(uuid uuid.UUID) ([]models.LoginAuditLog, error) {
	var results []models.LoginAuditLog
	for _, data := range p.logs {
		if data.UserId == uuid {
			results = append(results, data)
		}
	}
	return results, nil
}

func (p *loginAuditLogPersister) GetByGuestUserId(uuid uuid.UUID) ([]models.LoginAuditLog, error) {
	var results []models.LoginAuditLog
	for _, data := range p.logs {
		if *data.SurrogateUserId == uuid {
			results = append(results, data)
		}
	}
	return results, nil
}

func (p *loginAuditLogPersister) GetByGuestUserIdAndGrantId(guestUserId uuid.UUID, grantId uuid.UUID) ([]models.LoginAuditLog, error) {
	var results []models.LoginAuditLog
	for _, data := range p.logs {
		if *data.SurrogateUserId == guestUserId && *data.UserGuestRelationId == grantId {
			results = append(results, data)
		}
	}
	return results, nil
}
