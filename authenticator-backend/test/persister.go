package test

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewPersister(user []models.User, passcodes []models.Passcode, jwks []models.Jwk, credentials []models.WebauthnCredential, sessionData []models.WebauthnSessionData, passwords []models.PasswordCredential, accessGrants []models.AccountAccessGrant, userGuestRelations []models.UserGuestRelation, loginAudits []models.LoginAuditLog) persistence.Persister {
	return &persister{
		userPersister:                          NewUserPersister(user),
		passcodePersister:                      NewPasscodePersister(passcodes),
		jwkPersister:                           NewJwkPersister(jwks),
		webauthnCredentialPersister:            NewWebauthnCredentialPersister(credentials),
		webauthnSessionDataPersister:           NewWebauthnSessionDataPersister(sessionData),
		passwordCredentialPersister:            NewPasswordCredentialPersister(passwords),
		accountAccessGrantPersister:            NewAccountAccessGrantPersister(accessGrants),
		userGuestRelationPersister:             NewUserGuestRelationPersister(userGuestRelations),
		loginAuditLogPersister:                 NewLoginAuditLogPersister(loginAudits),
		webauthnCredentialsPrivateKeyPersister: NewWebauthnCredentialsPrivateKeyPersister([]models.WebauthnCredentialsPrivateKey{}),
		postPersister:                          NewPostPersister(nil),
	}
}

type persister struct {
	userPersister                          persistence.UserPersister
	passcodePersister                      persistence.PasscodePersister
	jwkPersister                           persistence.JwkPersister
	webauthnCredentialPersister            persistence.WebauthnCredentialPersister
	webauthnCredentialsPrivateKeyPersister persistence.WebauthnCredentialsPrivateKeyPersister
	webauthnSessionDataPersister           persistence.WebauthnSessionDataPersister
	passwordCredentialPersister            persistence.PasswordCredentialPersister
	accountAccessGrantPersister            persistence.AccountAccessGrantPersister
	userGuestRelationPersister             persistence.UserGuestRelationPersister
	loginAuditLogPersister                 persistence.LoginAuditLogPersister
	postPersister                          persistence.PostPersister
}

func (p *persister) GetPasswordCredentialPersister() persistence.PasswordCredentialPersister {
	return p.passwordCredentialPersister
}

func (p *persister) GetPasswordCredentialPersisterWithConnection(_ *pop.Connection) persistence.PasswordCredentialPersister {
	return p.passwordCredentialPersister
}

func (*persister) GetConnection() *pop.Connection {
	return nil
}

func (*persister) Transaction(fn func(tx *pop.Connection) error) error {
	return fn(nil)
}

func (p *persister) GetUserPersister() persistence.UserPersister {
	return p.userPersister
}

func (p *persister) GetUserPersisterWithConnection(_ *pop.Connection) persistence.UserPersister {
	return p.userPersister
}

func (p *persister) GetPasscodePersister() persistence.PasscodePersister {
	return p.passcodePersister
}

func (p *persister) GetPasscodePersisterWithConnection(_ *pop.Connection) persistence.PasscodePersister {
	return p.passcodePersister
}

func (p *persister) GetWebauthnCredentialPersister() persistence.WebauthnCredentialPersister {
	return p.webauthnCredentialPersister
}

func (p *persister) GetWebauthnCredentialsPrivateKeyPersister() persistence.WebauthnCredentialsPrivateKeyPersister {
	return p.webauthnCredentialsPrivateKeyPersister
}

func (p *persister) GetWebauthnCredentialPersisterWithConnection(_ *pop.Connection) persistence.WebauthnCredentialPersister {
	return p.webauthnCredentialPersister
}

func (p *persister) GetWebauthnSessionDataPersister() persistence.WebauthnSessionDataPersister {
	return p.webauthnSessionDataPersister
}

func (p *persister) GetWebauthnSessionDataPersisterWithConnection(_ *pop.Connection) persistence.WebauthnSessionDataPersister {
	return p.webauthnSessionDataPersister
}

func (p *persister) GetAccountAccessGrantPersister() persistence.AccountAccessGrantPersister {
	return p.accountAccessGrantPersister
}

func (p *persister) GetJwkPersister() persistence.JwkPersister {
	return p.jwkPersister
}

func (p *persister) GetJwkPersisterWithConnection(_ *pop.Connection) persistence.JwkPersister {
	return p.jwkPersister
}

func (p *persister) GetUserGuestRelationPersister() persistence.UserGuestRelationPersister {
	return p.userGuestRelationPersister
}

func (p *persister) GetLoginAuditLogPersister() persistence.LoginAuditLogPersister {
	return p.loginAuditLogPersister
}

func (p *persister) GetPostPersister() persistence.PostPersister {
	return p.postPersister
}
