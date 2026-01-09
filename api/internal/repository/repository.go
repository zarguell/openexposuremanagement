package repository

import (
	"github.com/jmoiron/sqlx"
)

// Repositories holds all repository instances
type Repositories struct {
	Tenant *TenantRepository
	User   *UserRepository
	APIKey *APIKeyRepository
}

// NewRepositories creates all repository instances
func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		Tenant: NewTenantRepository(db),
		User:   NewUserRepository(db),
		APIKey: NewAPIKeyRepository(db),
	}
}
