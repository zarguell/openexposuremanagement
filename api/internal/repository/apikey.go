package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// APIKey represents an API key in the system
type APIKey struct {
	ID          int64      `db:"id" json:"id"`
	TenantID    int64      `db:"tenant_id" json:"tenant_id"`
	Name        string     `db:"name" json:"name"`
	KeyHash     string     `db:"key_hash" json:"key_hash"`
	ScopesJSON  string     `db:"scopes_json" json:"scopes_json"`
	BoundSource *string    `db:"bound_source" json:"bound_source"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	RevokedAt   *time.Time `db:"revoked_at" json:"revoked_at"`
}

// APIKeyRepository handles API key data access
type APIKeyRepository struct {
	db *sqlx.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *sqlx.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// GetByKeyHash retrieves an API key by its hash
func (r *APIKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*APIKey, error) {
	var key APIKey
	err := r.db.GetContext(ctx, &key,
		"SELECT * FROM api_keys WHERE key_hash = $1 AND revoked_at IS NULL", keyHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Key not found
		}
		return nil, err
	}
	return &key, nil
}

// GetByID retrieves an API key by ID
func (r *APIKeyRepository) GetByID(ctx context.Context, id int64) (*APIKey, error) {
	var key APIKey
	err := r.db.GetContext(ctx, &key, "SELECT * FROM api_keys WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// Create creates a new API key
func (r *APIKeyRepository) Create(ctx context.Context, key *APIKey) error {
	query := `INSERT INTO api_keys (tenant_id, name, key_hash, scopes_json, bound_source)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		key.TenantID, key.Name, key.KeyHash, key.ScopesJSON, key.BoundSource).
		Scan(&key.ID, &key.CreatedAt)
}

// Revoke revokes an API key
func (r *APIKeyRepository) Revoke(ctx context.Context, id int64) error {
	query := `UPDATE api_keys SET revoked_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
