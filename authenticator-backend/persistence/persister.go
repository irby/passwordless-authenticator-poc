package persistence

import (
	"embed"

	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/config"
)

//go:embed migrations/*
var migrations embed.FS

// Persister is the persistence interface connecting to the database and capable of doing migrations
type persister struct {
	DB *pop.Connection
}

type Persister interface {
	GetConnection() *pop.Connection
	Transaction(func(tx *pop.Connection) error) error
	GetUserPersister() UserPersister
	GetUserPersisterWithConnection(tx *pop.Connection) UserPersister
	GetPasscodePersister() PasscodePersister
	GetPasscodePersisterWithConnection(tx *pop.Connection) PasscodePersister
	GetPasswordCredentialPersister() PasswordCredentialPersister
	GetPasswordCredentialPersisterWithConnection(tx *pop.Connection) PasswordCredentialPersister
	GetWebauthnCredentialPersister() WebauthnCredentialPersister
	GetWebauthnCredentialsPrivateKeyPersister() WebauthnCredentialsPrivateKeyPersister
	GetWebauthnCredentialPersisterWithConnection(tx *pop.Connection) WebauthnCredentialPersister
	GetWebauthnSessionDataPersister() WebauthnSessionDataPersister
	GetWebauthnSessionDataPersisterWithConnection(tx *pop.Connection) WebauthnSessionDataPersister
	GetJwkPersister() JwkPersister
	GetJwkPersisterWithConnection(tx *pop.Connection) JwkPersister
	GetAccountAccessGrantPersister() AccountAccessGrantPersister
	GetUserGuestRelationPersister() UserGuestRelationPersister
	GetLoginAuditLogPersister() LoginAuditLogPersister
	GetPostPersister() PostPersister
}

type Migrator interface {
	MigrateUp() error
	MigrateDown(int) error
}

type Storage interface {
	Migrator
	Persister
}

// New return a new Persister Object with given configuration
func New(config config.Database) (Storage, error) {
	DB, err := pop.NewConnection(&pop.ConnectionDetails{
		Dialect:  config.Dialect,
		Database: config.Database,
		Host:     config.Host,
		Port:     config.Port,
		User:     config.User,
		Password: config.Password,
		Pool:     5,
		IdlePool: 0,
	})

	if err != nil {
		return nil, err
	}

	if err := DB.Open(); err != nil {
		return nil, err
	}

	return &persister{
		DB: DB,
	}, nil
}

// MigrateUp applies all pending up migrations to the Database
func (p *persister) MigrateUp() error {
	migrationBox, err := pop.NewMigrationBox(migrations, p.DB)
	if err != nil {
		return err
	}
	err = migrationBox.Up()
	if err != nil {
		return err
	}
	return nil
}

// MigrateDown migrates the Database down by the given number of steps
func (p *persister) MigrateDown(steps int) error {
	migrationBox, err := pop.NewMigrationBox(migrations, p.DB)
	if err != nil {
		return err
	}
	err = migrationBox.Down(steps)
	if err != nil {
		return err
	}
	return nil
}

func (p *persister) GetConnection() *pop.Connection {
	return p.DB
}

func (p *persister) GetUserPersister() UserPersister {
	return NewUserPersister(p.DB)
}

func (*persister) GetUserPersisterWithConnection(tx *pop.Connection) UserPersister {
	return NewUserPersister(tx)
}

func (p *persister) GetPasscodePersister() PasscodePersister {
	return NewPasscodePersister(p.DB)
}

func (*persister) GetPasscodePersisterWithConnection(tx *pop.Connection) PasscodePersister {
	return NewPasscodePersister(tx)
}

func (p *persister) GetPasswordCredentialPersister() PasswordCredentialPersister {
	return NewPasswordCredentialPersister(p.DB)
}

func (*persister) GetPasswordCredentialPersisterWithConnection(tx *pop.Connection) PasswordCredentialPersister {
	return NewPasswordCredentialPersister(tx)
}

func (p *persister) GetWebauthnCredentialPersister() WebauthnCredentialPersister {
	return NewWebauthnCredentialPersister(p.DB)
}

func (p *persister) GetWebauthnCredentialsPrivateKeyPersister() WebauthnCredentialsPrivateKeyPersister {
	return NewWebauthnCredentialsPrivateKeyPersister(p.DB)
}

func (*persister) GetWebauthnCredentialPersisterWithConnection(tx *pop.Connection) WebauthnCredentialPersister {
	return NewWebauthnCredentialPersister(tx)
}

func (p *persister) GetWebauthnSessionDataPersister() WebauthnSessionDataPersister {
	return NewWebauthnSessionDataPersister(p.DB)
}

func (*persister) GetWebauthnSessionDataPersisterWithConnection(tx *pop.Connection) WebauthnSessionDataPersister {
	return NewWebauthnSessionDataPersister(tx)
}

func (p *persister) GetAccountAccessGrantPersister() AccountAccessGrantPersister {
	return NewAccountAccessGrantPersister(p.DB)
}

func (p *persister) GetJwkPersister() JwkPersister {
	return NewJwkPersister(p.DB)
}

func (*persister) GetJwkPersisterWithConnection(tx *pop.Connection) JwkPersister {
	return NewJwkPersister(tx)
}

func (p *persister) Transaction(fn func(tx *pop.Connection) error) error {
	return p.DB.Transaction(fn)
}

func (p *persister) GetUserGuestRelationPersister() UserGuestRelationPersister {
	return NewUserGuestRelationPersister(p.DB)
}

func (p *persister) GetLoginAuditLogPersister() LoginAuditLogPersister {
	return NewLoginAuditLogPersister(p.DB)
}

func (p *persister) GetPostPersister() PostPersister {
	return NewPostPersister(p.DB)
}
