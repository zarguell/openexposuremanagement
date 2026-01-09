package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
)

// User represents a user in the system
type User struct {
	ID          int64  `db:"id" json:"id"`
	TenantID    int64  `db:"tenant_id" json:"tenant_id"`
	Email       string `db:"email" json:"email"`
	DisplayName string `db:"display_name" json:"display_name"`
	Status      string `db:"status" json:"status"`
	CreatedAt   string `db:"created_at" json:"created_at"`
	UpdatedAt   string `db:"updated_at" json:"updated_at"`
}

// UserRepository handles user data access
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByEmail retrieves a user by email within a tenant
func (r *UserRepository) GetByEmail(ctx context.Context, tenantID int64, email string) (*User, error) {
	var user User
	err := r.db.GetContext(ctx, &user,
		"SELECT * FROM users WHERE tenant_id = $1 AND email = $2", tenantID, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *User) error {
	query := `INSERT INTO users (tenant_id, email, display_name, status)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	rows, err := r.db.QueryContext(ctx, query, user.TenantID, user.Email, user.DisplayName, user.Status)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	}

	return nil
}
