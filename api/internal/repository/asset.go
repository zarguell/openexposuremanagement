package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// Asset represents an asset in the system
type Asset struct {
	ID            int64     `db:"id" json:"id"`
	TenantID      int64     `db:"tenant_id" json:"tenant_id"`
	CanonicalName string    `db:"canonical_name" json:"canonical_name"`
	FirstSeenAt   time.Time `db:"first_seen_at" json:"first_seen_at"`
	LastSeenAt    time.Time `db:"last_seen_at" json:"last_seen_at"`
	OwnerTeamID   *int64    `db:"owner_team_id" json:"owner_team_id,omitempty"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// AssetIdentifier represents an asset identifier
type AssetIdentifier struct {
	ID          int64     `db:"id" json:"id"`
	TenantID    int64     `db:"tenant_id" json:"tenant_id"`
	AssetID     int64     `db:"asset_id" json:"asset_id"`
	IDType      string    `db:"id_type" json:"id_type"`
	IDValue     string    `db:"id_value" json:"id_value"`
	FirstSeenAt time.Time `db:"first_seen_at" json:"first_seen_at"`
	LastSeenAt  time.Time `db:"last_seen_at" json:"last_seen_at"`
	Source      string    `db:"source" json:"source"`
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
	if identifiers == nil {
		identifiers = []AssetIdentifier{}
	}
	return identifiers, nil
}

// UpdateCanonicalName updates the canonical name for an asset
func (r *AssetRepository) UpdateCanonicalName(ctx context.Context, id int64, canonicalName string) error {
	query := `UPDATE assets SET canonical_name = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, canonicalName, id)
	return err
}

// UpdateIdentifierLastSeen updates the last_seen_at timestamp for an identifier
func (r *AssetRepository) UpdateIdentifierLastSeen(ctx context.Context, tenantID, assetID int64, idType, idValue string, lastSeen time.Time) error {
	query := `UPDATE asset_identifiers SET last_seen_at = $1 WHERE tenant_id = $2 AND asset_id = $3 AND id_type = $4 AND id_value = $5`
	_, err := r.db.ExecContext(ctx, query, lastSeen, tenantID, assetID, idType, idValue)
	return err
}

// FindByIdentifier finds an asset by identifier type and value
func (r *AssetRepository) FindByIdentifier(ctx context.Context, tenantID int64, idType, idValue string) (*Asset, error) {
	var asset Asset
	query := `SELECT a.* FROM assets a
		INNER JOIN asset_identifiers ai ON a.id = ai.asset_id
		WHERE ai.tenant_id = $1 AND ai.id_type = $2 AND ai.id_value = $3`
	err := r.db.GetContext(ctx, &asset, query, tenantID, idType, idValue)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// FindByIPAndHostname finds an asset that has both the given IP and hostname identifiers
func (r *AssetRepository) FindByIPAndHostname(ctx context.Context, tenantID int64, ip, hostname string) (*Asset, error) {
	var asset Asset
	query := `SELECT DISTINCT a.* FROM assets a
		INNER JOIN asset_identifiers ai1 ON a.id = ai1.asset_id
		INNER JOIN asset_identifiers ai2 ON a.id = ai2.asset_id
		WHERE a.tenant_id = $1
		AND ai1.id_type = 'ipv4' AND ai1.id_value = $2
		AND ai2.id_type = 'hostname_norm' AND ai2.id_value = $3`
	err := r.db.GetContext(ctx, &asset, query, tenantID, ip, hostname)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// GetExistingIdentifier gets an existing identifier for an asset
func (r *AssetRepository) GetExistingIdentifier(ctx context.Context, tenantID, assetID int64, idType, idValue string) (*AssetIdentifier, error) {
	var identifier AssetIdentifier
	query := `SELECT * FROM asset_identifiers
		WHERE tenant_id = $1 AND asset_id = $2 AND id_type = $3 AND id_value = $4`
	err := r.db.GetContext(ctx, &identifier, query, tenantID, assetID, idType, idValue)
	if err != nil {
		return nil, err
	}
	return &identifier, nil
}

// UpsertIdentifier creates or updates an identifier for an asset
func (r *AssetRepository) UpsertIdentifier(ctx context.Context, tenantID, assetID int64, idType, idValue, source string, firstSeen, lastSeen time.Time) error {
	// First try to find existing identifier
	existing, err := r.GetExistingIdentifier(ctx, tenantID, assetID, idType, idValue)
	if err != nil {
		// Identifier doesn't exist, create it
		identifier := &AssetIdentifier{
			TenantID:    tenantID,
			AssetID:     assetID,
			IDType:      idType,
			IDValue:     idValue,
			FirstSeenAt: firstSeen,
			LastSeenAt:  lastSeen,
			Source:      source,
		}
		return r.AddIdentifier(ctx, identifier)
	}

	// Identifier exists, update last_seen_at if it's newer
	if lastSeen.After(existing.LastSeenAt) {
		return r.UpdateIdentifierLastSeen(ctx, tenantID, assetID, idType, idValue, lastSeen)
	}

	return nil
}

// AssetListParams represents query parameters for listing assets
type AssetListParams struct {
	TenantID int64
	Query    string // Optional search query for canonical name
	Limit    int    // Max results to return (0 for no limit)
	Offset   int    // Number of results to skip
}

// AssetListResult represents the result of listing assets
type AssetListResult struct {
	Assets []Asset
	Total  int // Total count matching query (for pagination)
	Limit  int
	Offset int
}

// List retrieves assets with optional filtering and pagination
func (r *AssetRepository) List(ctx context.Context, params AssetListParams) (*AssetListResult, error) {
	// Build base query
	baseQuery := "WHERE tenant_id = $1"
	args := []interface{}{params.TenantID}
	argCount := 1

	// Add search filter if provided
	if params.Query != "" {
		argCount++
		baseQuery += " AND canonical_name ILIKE $" + string(rune('0'+argCount))
		args = append(args, "%"+params.Query+"%")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM assets " + baseQuery
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Build select query
	selectQuery := "SELECT * FROM assets " + baseQuery + " ORDER BY canonical_name"
	if params.Limit > 0 {
		argCount++
		selectQuery += " LIMIT $" + string(rune('0'+argCount))
		args = append(args, params.Limit)
	}
	if params.Offset > 0 {
		argCount++
		selectQuery += " OFFSET $" + string(rune('0'+argCount))
		args = append(args, params.Offset)
	}

	var assets []Asset
	err = r.db.SelectContext(ctx, &assets, selectQuery, args...)
	if err != nil {
		return nil, err
	}

	return &AssetListResult{
		Assets: assets,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// AssetDetail represents an asset with its identifiers and finding counts
type AssetDetail struct {
	Asset
	Identifiers   []AssetIdentifier `json:"identifiers"`
	OpenFindings  int               `json:"open_findings"`
	FixedFindings int               `json:"fixed_findings"`
	TotalFindings int               `json:"total_findings"`
}

// GetWithDetails retrieves an asset with its identifiers and finding counts
func (r *AssetRepository) GetWithDetails(ctx context.Context, tenantID, assetID int64) (*AssetDetail, error) {
	// Get the asset
	asset, err := r.GetByID(ctx, assetID)
	if err != nil {
		return nil, err
	}

	// Verify tenant ownership
	if asset.TenantID != tenantID {
		return nil, nil // Will be treated as 404
	}

	// Get identifiers
	identifiers, err := r.GetIdentifiers(ctx, assetID)
	if err != nil {
		return nil, err
	}

	// Get finding counts
	openCount, fixedCount, err := r.getFindingCounts(ctx, assetID)
	if err != nil {
		return nil, err
	}

	return &AssetDetail{
		Asset:         *asset,
		Identifiers:   identifiers,
		OpenFindings:  openCount,
		FixedFindings: fixedCount,
		TotalFindings: openCount + fixedCount,
	}, nil
}

// getFindingCounts gets the count of open and fixed findings for an asset
func (r *AssetRepository) getFindingCounts(ctx context.Context, assetID int64) (open, fixed int, err error) {
	query := `
		SELECT effective_status,
			COUNT(*) as count
		FROM finding_instances
		WHERE asset_id = $1
		GROUP BY effective_status
	`

	rows, err := r.db.QueryContext(ctx, query, assetID)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return 0, 0, err
		}
		switch status {
		case "open":
			open = count
		case "fixed":
			fixed = count
		}
	}

	return open, fixed, rows.Err()
}
