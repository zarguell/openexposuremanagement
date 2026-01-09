package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// TenantPolicyState represents the policy state for a tenant
type TenantPolicyState struct {
	TenantID       int64     `db:"tenant_id"`
	PolicyRevision int64     `db:"policy_revision"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// TenantPolicyStateRepository handles tenant policy state operations
type TenantPolicyStateRepository struct {
	db *sqlx.DB
}

// NewTenantPolicyStateRepository creates a new tenant policy state repository
func NewTenantPolicyStateRepository(db *sqlx.DB) *TenantPolicyStateRepository {
	return &TenantPolicyStateRepository{db: db}
}

// Get retrieves the policy state for a tenant
// If no state exists, returns nil (caller should create default)
func (r *TenantPolicyStateRepository) Get(ctx context.Context, tenantID int64) (*TenantPolicyState, error) {
	query := `
		SELECT tenant_id, policy_revision, updated_at
		FROM tenant_policy_state
		WHERE tenant_id = $1
	`

	var state TenantPolicyState
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&state.TenantID,
		&state.PolicyRevision,
		&state.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &state, nil
}

// GetOrCreate retrieves the policy state for a tenant, or creates a default one
func (r *TenantPolicyStateRepository) GetOrCreate(ctx context.Context, tenantID int64) (*TenantPolicyState, error) {
	// Try to get existing state
	state, err := r.Get(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// If state exists, return it
	if state != nil {
		return state, nil
	}

	// Create default state with revision = 1
	state = &TenantPolicyState{
		TenantID:       tenantID,
		PolicyRevision: 1,
	}

	err = r.Create(ctx, state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// Create creates a new policy state for a tenant
func (r *TenantPolicyStateRepository) Create(ctx context.Context, state *TenantPolicyState) error {
	query := `
		INSERT INTO tenant_policy_state (tenant_id, policy_revision, updated_at)
		VALUES ($1, $2, NOW())
		RETURNING tenant_id, policy_revision, updated_at
	`

	return r.db.QueryRowContext(ctx, query,
		state.TenantID, state.PolicyRevision,
	).Scan(
		&state.TenantID,
		&state.PolicyRevision,
		&state.UpdatedAt,
	)
}

// IncrementRevision increments the policy revision for a tenant
// This is called when a suppression is approved or revoked
func (r *TenantPolicyStateRepository) IncrementRevision(ctx context.Context, tenantID int64) (int64, error) {
	query := `
		UPDATE tenant_policy_state
		SET policy_revision = policy_revision + 1,
		    updated_at = NOW()
		WHERE tenant_id = $1
		RETURNING policy_revision
	`

	var newRevision int64
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&newRevision)
	if err != nil {
		return 0, err
	}

	return newRevision, nil
}
