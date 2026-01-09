package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// Asset represents an asset in the system
type Asset struct {
	ID            int64      `db:"id" json:"id"`
	TenantID      int64      `db:"tenant_id" json:"tenant_id"`
	CanonicalName string     `db:"canonical_name" json:"canonical_name"`
	FirstSeenAt   time.Time  `db:"first_seen_at" json:"first_seen_at"`
	LastSeenAt    time.Time  `db:"last_seen_at" json:"last_seen_at"`
	OwnerTeamID   *int64     `db:"owner_team_id" json:"owner_team_id,omitempty"`
	IsActive      bool       `db:"is_active" json:"is_active"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// AssetIdentifier represents an asset identifier
type AssetIdentifier struct {
	ID           int64      `db:"id" json:"id"`
	TenantID     int64      `db:"tenant_id" json:"tenant_id"`
	AssetID      int64      `db:"asset_id" json:"asset_id"`
	IDType       string     `db:"id_type" json:"id_type"`
	IDValue      string     `db:"id_value" json:"id_value"`
	FirstSeenAt  time.Time  `db:"first_seen_at" json:"first_seen_at"`
	LastSeenAt   time.Time  `db:"last_seen_at" json:"last_seen_at"`
	Source       string     `db:"source" json:"source"`
}

// AssetRepository handles asset data access
type AssetRepository struct {
	db *sqlx.DB
}

// NewAssetRepository creates a new asset repository
func NewAssetRepository(db *sqlx.DB) *AssetRepository {
	return &AssetRepository{db: db}
}

// GetByID retrieves an asset by ID
func (r *AssetRepository) GetByID(ctx context.Context, id int64) (*Asset, error) {
	var asset Asset
	err := r.db.GetContext(ctx, &asset, "SELECT * FROM assets WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// GetByCanonicalName retrieves an asset by canonical name within a tenant
func (r *AssetRepository) GetByCanonicalName(ctx context.Context, tenantID int64, canonicalName string) (*Asset, error) {
	var asset Asset
	err := r.db.GetContext(ctx, &asset,
		"SELECT * FROM assets WHERE tenant_id = $1 AND canonical_name = $2",
		tenantID, canonicalName)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// Create creates a new asset
func (r *AssetRepository) Create(ctx context.Context, asset *Asset) error {
	query := `INSERT INTO assets (tenant_id, canonical_name, first_seen_at, last_seen_at, owner_team_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		asset.TenantID, asset.CanonicalName, asset.FirstSeenAt, asset.LastSeenAt,
		asset.OwnerTeamID, asset.IsActive).
		Scan(&asset.ID, &asset.CreatedAt, &asset.UpdatedAt)
}

// UpdateLastSeen updates the last_seen_at timestamp for an asset
func (r *AssetRepository) UpdateLastSeen(ctx context.Context, id int64, lastSeen time.Time) error {
	query := `UPDATE assets SET last_seen_at = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, lastSeen, id)
	return err
}

// AddIdentifier adds an identifier to an asset
func (r *AssetRepository) AddIdentifier(ctx context.Context, identifier *AssetIdentifier) error {
	query := `INSERT INTO asset_identifiers (tenant_id, asset_id, id_type, id_value, first_seen_at, last_seen_at, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		identifier.TenantID, identifier.AssetID, identifier.IDType, identifier.IDValue,
		identifier.FirstSeenAt, identifier.LastSeenAt, identifier.Source).
		Scan(&identifier.ID)
}

// GetIdentifiers retrieves all identifiers for an asset
func (r *AssetRepository) GetIdentifiers(ctx context.Context, assetID int64) ([]AssetIdentifier, error) {
	var identifiers []AssetIdentifier
	err := r.db.SelectContext(ctx, &identifiers,
		"SELECT * FROM asset_identifiers WHERE asset_id = $1", assetID)
	if err != nil {
		return nil, err
	}
	return identifiers, nil
}
