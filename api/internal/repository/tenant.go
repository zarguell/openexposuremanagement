package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
)

// Tenant represents a tenant in the system
type Tenant struct {
	ID        int64  `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	CreatedAt string `db:"created_at" json:"created_at"`
	UpdatedAt string `db:"updated_at" json:"updated_at"`
}

// TenantRepository handles tenant data access
type TenantRepository struct {
	db *sqlx.DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *sqlx.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// GetByName retrieves a tenant by name
func (r *TenantRepository) GetByName(ctx context.Context, name string) (*Tenant, error) {
	var tenant Tenant
	err := r.db.GetContext(ctx, &tenant, "SELECT * FROM tenants WHERE name = $1", name)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(ctx context.Context, id int64) (*Tenant, error) {
	var tenant Tenant
	err := r.db.GetContext(ctx, &tenant, "SELECT * FROM tenants WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// Create creates a new tenant
func (r *TenantRepository) Create(ctx context.Context, tenant *Tenant) error {
	query := `INSERT INTO tenants (name) VALUES ($1) RETURNING id, created_at, updated_at`
	rows, err := r.db.QueryContext(ctx, query, tenant.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&tenant.ID, &tenant.CreatedAt, &tenant.UpdatedAt)
	}

	return nil
}
