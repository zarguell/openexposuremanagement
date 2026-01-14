package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Software represents a software product in the catalog
type Software struct {
	ID             int64  `db:"id" json:"id"`
	CPEString      string `db:"cpe_string" json:"cpe_string"`
	Vendor         string `db:"vendor" json:"vendor"`
	ProductName    string `db:"product_name" json:"product_name"`
	Version        string `db:"version" json:"version,omitempty"`
	Edition        string `db:"edition" json:"edition,omitempty"`
	TargetHW       string `db:"target_hw" json:"target_hw,omitempty"`
	Lang           string `db:"lang" json:"lang,omitempty"`
	TitleFormatted string `db:"title_formatted" json:"title_formatted"`
	CreatedAt      string `db:"created_at" json:"created_at"`
	UpdatedAt      string `db:"updated_at" json:"updated_at"`
}

// AssetSoftware represents a software installation on an asset
type AssetSoftware struct {
	ID          int64  `db:"id" json:"id"`
	TenantID    int64  `db:"tenant_id" json:"tenant_id"`
	AssetID     int64  `db:"asset_id" json:"asset_id"`
	SoftwareID  int64  `db:"software_id" json:"software_id"`
	Source      string `db:"source" json:"source"`
	InstallPath string `db:"install_path" json:"install_path,omitempty"`
	FirstSeenAt string `db:"first_seen_at" json:"first_seen_at"`
	LastSeenAt  string `db:"last_seen_at" json:"last_seen_at"`
	CreatedAt   string `db:"created_at" json:"created_at"`
	UpdatedAt   string `db:"updated_at" json:"updated_at"`
}

// SoftwareSummary represents a software product with installation count
type SoftwareSummary struct {
	SoftwareID     int64  `db:"software_id" json:"software_id"`
	CPEString      string `db:"cpe_string" json:"cpe_string"`
	Vendor         string `db:"vendor" json:"vendor"`
	ProductName    string `db:"product_name" json:"product_name"`
	Version        string `db:"version" json:"version,omitempty"`
	TitleFormatted string `db:"title_formatted" json:"title_formatted"`
	InstallCount   int64  `db:"install_count" json:"install_count"`
}

// SoftwareListParams represents parameters for listing software
type SoftwareListParams struct {
	TenantID  int64  `json:"-"` // Tenant ID (required)
	Vendor    string `json:"vendor,omitempty"`
	Product   string `json:"product,omitempty"`
	Version   string `json:"version,omitempty"`
	CPE       string `json:"cpe,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// SoftwareRepository handles software data access
type SoftwareRepository struct {
	db *sqlx.DB
}

// NewSoftwareRepository creates a new software repository
func NewSoftwareRepository(db *sqlx.DB) *SoftwareRepository {
	return &SoftwareRepository{db: db}
}

// GetByID retrieves software by ID
func (r *SoftwareRepository) GetByID(ctx context.Context, id int64) (*Software, error) {
	var sw Software
	err := r.db.GetContext(ctx, &sw, "SELECT * FROM software WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &sw, nil
}

// List retrieves a paginated list of software with optional filters
func (r *SoftwareRepository) List(ctx context.Context, params SoftwareListParams) ([]SoftwareSummary, int, error) {
	// Build base query
	baseQuery := `
		SELECT s.id as software_id,
		       s.cpe_string,
		       s.vendor,
		       s.product_name,
		       s.version,
		       s.title_formatted,
		       COUNT(DISTINCT asw.asset_id) as install_count
		FROM software s
		LEFT JOIN asset_software asw ON asw.software_id = s.id
		WHERE 1=1
	`

	// Build WHERE clause
	whereClause := ""
	args := []interface{}{}
	argCount := 1

	if params.Vendor != "" {
		whereClause += fmt.Sprintf(" AND s.vendor ILIKE $%d", argCount)
		args = append(args, "%"+params.Vendor+"%")
		argCount++
	}

	if params.Product != "" {
		whereClause += fmt.Sprintf(" AND s.product_name ILIKE $%d", argCount)
		args = append(args, "%"+params.Product+"%")
		argCount++
	}

	if params.CPE != "" {
		whereClause += fmt.Sprintf(" AND s.cpe_string LIKE $%d", argCount)
		args = append(args, params.CPE+"%")
		argCount++
	}

	// Add WHERE clause to base query
	query := baseQuery + whereClause + " GROUP BY s.id, s.cpe_string, s.vendor, s.product_name, s.version, s.title_formatted"

	// Get total count
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as count"
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count software: %w", err)
	}

	// Add ordering and pagination
	query += " ORDER BY s.vendor, s.product_name, s.version"

	if params.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, params.Limit)
		argCount++
	}

	if params.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, params.Offset)
	}

	// Execute query
	var software []SoftwareSummary
	err = r.db.SelectContext(ctx, &software, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list software: %w", err)
	}

	return software, total, nil
}

// GetSoftwareForAsset retrieves all software installed on a specific asset
func (r *SoftwareRepository) GetSoftwareForAsset(ctx context.Context, tenantID, assetID int64) ([]AssetSoftware, error) {
	query := `
		SELECT asw.id,
		       asw.tenant_id,
		       asw.asset_id,
		       asw.software_id,
		       asw.source,
		       asw.install_path,
		       asw.first_seen_at,
		       asw.last_seen_at,
		       asw.created_at,
		       asw.updated_at,
		       s.cpe_string,
		       s.vendor,
		       s.product_name,
		       s.version,
		       s.edition,
		       s.title_formatted
		FROM asset_software asw
		JOIN software s ON s.id = asw.software_id
		WHERE asw.tenant_id = $1 AND asw.asset_id = $2
		ORDER BY s.vendor, s.product_name
	`

	var software []AssetSoftware
	err := r.db.SelectContext(ctx, &software, query, tenantID, assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get software for asset: %w", err)
	}

	return software, nil
}

// GetAffectedAssets retrieves all assets that have a specific software installed
func (r *SoftwareRepository) GetAffectedAssets(ctx context.Context, tenantID, softwareID int64, limit, offset int) ([]map[string]interface{}, int, error) {
	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(DISTINCT a.id)
			FROM assets a
			JOIN asset_software asw ON asw.asset_id = a.id
			WHERE a.tenant_id = $1 AND asw.software_id = $2`,
		tenantID, softwareID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count affected assets: %w", err)
	}

	// Get paginated results
	query := `
		SELECT a.id,
		       a.canonical_name,
		       a.is_active,
		       asw.first_seen_at,
		       asw.last_seen_at,
		       asw.install_path
		FROM assets a
		JOIN asset_software asw ON asw.asset_id = a.id
		WHERE a.tenant_id = $1 AND asw.software_id = $2
		ORDER BY a.canonical_name
	`

	args := []interface{}{tenantID, softwareID}

	if limit > 0 {
		query += " LIMIT $3"
		args = append(args, limit)
	}

	if offset > 0 {
		query += " OFFSET $4"
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get affected assets: %w", err)
	}
	defer rows.Close()

	var assets []map[string]interface{}
	for rows.Next() {
		var asset struct {
			ID            int64  `db:"id"`
			CanonicalName string `db:"canonical_name"`
			IsActive      bool   `db:"is_active"`
			FirstSeenAt   string `db:"first_seen_at"`
			LastSeenAt    string `db:"last_seen_at"`
			InstallPath   string `db:"install_path"`
		}
		if err := rows.Scan(&asset.ID, &asset.CanonicalName, &asset.IsActive, &asset.FirstSeenAt, &asset.LastSeenAt, &asset.InstallPath); err != nil {
			return nil, 0, fmt.Errorf("failed to scan asset: %w", err)
		}

		assets = append(assets, map[string]interface{}{
			"asset_id":       asset.ID,
			"canonical_name": asset.CanonicalName,
			"is_active":      asset.IsActive,
			"first_seen_at":  asset.FirstSeenAt,
			"last_seen_at":   asset.LastSeenAt,
			"install_path":   asset.InstallPath,
		})
	}

	return assets, total, nil
}

// GetSoftwareDetails retrieves detailed information about software including affected assets and related findings
func (r *SoftwareRepository) GetSoftwareDetails(ctx context.Context, tenantID, softwareID int64) (map[string]interface{}, error) {
	// Get software info
	var sw Software
	err := r.db.GetContext(ctx, &sw, "SELECT * FROM software WHERE id = $1", softwareID)
	if err != nil {
		return nil, fmt.Errorf("failed to get software: %w", err)
	}

	// Get affected assets (first 10 as preview)
	assets, totalAssets, err := r.GetAffectedAssets(ctx, tenantID, softwareID, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get affected assets: %w", err)
	}

	// Get related findings summary
	var findingSummary struct {
		TotalFindings    int64 `db:"total_findings"`
		CriticalCount    int64 `db:"critical_count"`
		HighCount        int64 `db:"high_count"`
		MediumCount      int64 `db:"medium_count"`
		LowCount         int64 `db:"low_count"`
	}

	err = r.db.GetContext(ctx, &findingSummary, `
		SELECT COALESCE(SUM(CASE
			WHEN fi.effective_status = 'open' THEN
				CASE
					WHEN fd.severity_default = 'Critical' THEN 1
					WHEN fd.severity_default = 'High' THEN 1
					WHEN fd.severity_default = 'Medium' THEN 1
					WHEN fd.severity_default = 'Low' THEN 1
					ELSE 0
				END
			ELSE 0
		END), 0) as total_findings,
		COALESCE(SUM(CASE WHEN fi.effective_status = 'open' AND fd.severity_default = 'Critical' THEN 1 ELSE 0 END), 0) as critical_count,
		COALESCE(SUM(CASE WHEN fi.effective_status = 'open' AND fd.severity_default = 'High' THEN 1 ELSE 0 END), 0) as high_count,
		COALESCE(SUM(CASE WHEN fi.effective_status = 'open' AND fd.severity_default = 'Medium' THEN 1 ELSE 0 END), 0) as medium_count,
		COALESCE(SUM(CASE WHEN fi.effective_status = 'open' AND fd.severity_default = 'Low' THEN 1 ELSE 0 END), 0) as low_count
		FROM finding_instances fi
		JOIN finding_definitions fd ON fi.definition_uid = fd.definition_uid
		JOIN asset_software asw ON asw.asset_id = fi.asset_id
		WHERE asw.software_id = $1 AND fi.effective_status = 'open'
	`, softwareID)
	if err != nil {
		return nil, fmt.Errorf("failed to get finding summary: %w", err)
	}

	return map[string]interface{}{
		"software":          sw,
		"affected_assets":   assets,
		"total_assets":      totalAssets,
		"affected_findings": findingSummary,
	}, nil
}
