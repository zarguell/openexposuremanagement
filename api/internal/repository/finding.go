package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

// FindingInstance represents a single finding instance
type FindingInstance struct {
	ID                int64                  `db:"id"`
	TenantID          int64                  `db:"tenant_id"`
	AssetID           int64                  `db:"asset_id"`
	DefinitionUID     string                 `db:"definition_uid"`
	ScannerStatus     string                 `db:"scanner_status"`
	FirstObservedAt   time.Time              `db:"first_observed_at"`
	LastObservedAt    time.Time              `db:"last_observed_at"`
	EvidenceJSON      map[string]interface{} `db:"evidence_json"`
	EffectiveStatus   string                 `db:"effective_status"`
	EffectiveReason   string                 `db:"effective_reason"`
	EffectiveRevision int64                  `db:"effective_revision"`
	CreatedAt         time.Time              `db:"created_at"`
	UpdatedAt         time.Time              `db:"updated_at"`
}

// EnrichedFinding represents a finding with additional context (definition, asset, intel)
type EnrichedFinding struct {
	FindingInstance
	AssetCanonicalName string               `json:"asset_canonical_name"`
	DefinitionTitle    string               `json:"definition_title"`
	DefinitionSeverity string               `json:"definition_severity"`
	CVEAliases         []string             `json:"cve_aliases,omitempty"`
	Intel              map[string]*IntelCVE `json:"intel,omitempty"` // Keyed by CVE
}

// FindingInstanceRepository handles finding instance operations
type FindingInstanceRepository struct {
	db *sqlx.DB
}

// NewFindingInstanceRepository creates a new finding instance repository
func NewFindingInstanceRepository(db *sqlx.DB) *FindingInstanceRepository {
	return &FindingInstanceRepository{db: db}
}

// UpsertFindingInstance creates or updates a finding instance with observation window tracking
func (r *FindingInstanceRepository) UpsertFindingInstance(ctx context.Context, instance *FindingInstance) error {
	query := `
		INSERT INTO finding_instances (
			tenant_id, asset_id, definition_uid, scanner_status,
			first_observed_at, last_observed_at, evidence_json,
			effective_status, effective_reason, effective_revision,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		ON CONFLICT (tenant_id, asset_id, definition_uid)
		DO UPDATE SET
			scanner_status = EXCLUDED.scanner_status,
			-- First observed: only move earlier
			first_observed_at = CASE
				WHEN EXCLUDED.first_observed_at < finding_instances.first_observed_at
				THEN EXCLUDED.first_observed_at
				ELSE finding_instances.first_observed_at
			END,
			-- Last observed: only move later
			last_observed_at = CASE
				WHEN EXCLUDED.last_observed_at > finding_instances.last_observed_at
				THEN EXCLUDED.last_observed_at
				ELSE finding_instances.last_observed_at
			END,
			evidence_json = EXCLUDED.evidence_json,
			effective_status = EXCLUDED.effective_status,
			effective_reason = EXCLUDED.effective_reason,
			effective_revision = EXCLUDED.effective_revision,
			updated_at = NOW()
		RETURNING id, tenant_id, asset_id, definition_uid, scanner_status,
				  first_observed_at, last_observed_at, evidence_json,
				  effective_status, effective_reason, effective_revision,
				  created_at, updated_at
	`

	var evidenceJSON []byte
	if instance.EvidenceJSON != nil && len(instance.EvidenceJSON) > 0 {
		var err error
		evidenceJSON, err = json.Marshal(instance.EvidenceJSON)
		if err != nil {
			return err
		}
	} else {
		// Use empty JSON object for nil/empty evidence
		evidenceJSON = []byte("{}")
	}

	err := r.db.QueryRowContext(ctx, query,
		instance.TenantID, instance.AssetID, instance.DefinitionUID,
		instance.ScannerStatus, instance.FirstObservedAt, instance.LastObservedAt,
		evidenceJSON, instance.EffectiveStatus, instance.EffectiveReason,
		instance.EffectiveRevision,
	).Scan(
		&instance.ID, &instance.TenantID, &instance.AssetID,
		&instance.DefinitionUID, &instance.ScannerStatus,
		&instance.FirstObservedAt, &instance.LastObservedAt,
		&evidenceJSON, &instance.EffectiveStatus, &instance.EffectiveReason,
		&instance.EffectiveRevision, &instance.CreatedAt, &instance.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// Unmarshal evidence back
	if evidenceJSON != nil {
		json.Unmarshal(evidenceJSON, &instance.EvidenceJSON)
	}

	return nil
}

// GetByTenantAndAsset retrieves all finding instances for a tenant and asset
func (r *FindingInstanceRepository) GetByTenantAndAsset(ctx context.Context, tenantID, assetID int64) ([]FindingInstance, error) {
	query := `
		SELECT id, tenant_id, asset_id, definition_uid, scanner_status,
		       first_observed_at, last_observed_at, evidence_json,
		       effective_status, effective_reason, effective_revision,
		       created_at, updated_at
		FROM finding_instances
		WHERE tenant_id = $1 AND asset_id = $2
		ORDER BY last_observed_at DESC
	`

	return r.queryInstances(ctx, query, tenantID, assetID)
}

// GetByTenant retrieves all finding instances for a tenant
func (r *FindingInstanceRepository) GetByTenant(ctx context.Context, tenantID int64) ([]FindingInstance, error) {
	query := `
		SELECT id, tenant_id, asset_id, definition_uid, scanner_status,
		       first_observed_at, last_observed_at, evidence_json,
		       effective_status, effective_reason, effective_revision,
		       created_at, updated_at
		FROM finding_instances
		WHERE tenant_id = $1
		ORDER BY last_observed_at DESC
	`

	return r.queryInstances(ctx, query, tenantID)
}

// GetByTenantAndDefinition retrieves all finding instances for a tenant and definition
func (r *FindingInstanceRepository) GetByTenantAndDefinition(ctx context.Context, tenantID int64, definitionUID string) ([]FindingInstance, error) {
	query := `
		SELECT id, tenant_id, asset_id, definition_uid, scanner_status,
		       first_observed_at, last_observed_at, evidence_json,
		       effective_status, effective_reason, effective_revision,
		       created_at, updated_at
		FROM finding_instances
		WHERE tenant_id = $1 AND definition_uid = $2
		ORDER BY last_observed_at DESC
	`

	return r.queryInstances(ctx, query, tenantID, definitionUID)
}

// GetByTenantAssetAndDefinition retrieves a specific finding instance
func (r *FindingInstanceRepository) GetByTenantAssetAndDefinition(ctx context.Context, tenantID, assetID int64, definitionUID string) (*FindingInstance, error) {
	query := `
		SELECT id, tenant_id, asset_id, definition_uid, scanner_status,
		       first_observed_at, last_observed_at, evidence_json,
		       effective_status, effective_reason, effective_revision,
		       created_at, updated_at
		FROM finding_instances
		WHERE tenant_id = $1 AND asset_id = $2 AND definition_uid = $3
	`

	var instance FindingInstance
	var evidenceJSON []byte

	err := r.db.QueryRowContext(ctx, query, tenantID, assetID, definitionUID).Scan(
		&instance.ID, &instance.TenantID, &instance.AssetID,
		&instance.DefinitionUID, &instance.ScannerStatus,
		&instance.FirstObservedAt, &instance.LastObservedAt,
		&evidenceJSON, &instance.EffectiveStatus, &instance.EffectiveReason,
		&instance.EffectiveRevision, &instance.CreatedAt, &instance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal evidence
	if evidenceJSON != nil {
		json.Unmarshal(evidenceJSON, &instance.EvidenceJSON)
	}

	return &instance, nil
}

// queryInstances is a helper for querying multiple instances
func (r *FindingInstanceRepository) queryInstances(ctx context.Context, query string, args ...interface{}) ([]FindingInstance, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []FindingInstance
	for rows.Next() {
		var instance FindingInstance
		var evidenceJSON []byte

		if err := rows.Scan(
			&instance.ID, &instance.TenantID, &instance.AssetID,
			&instance.DefinitionUID, &instance.ScannerStatus,
			&instance.FirstObservedAt, &instance.LastObservedAt,
			&evidenceJSON, &instance.EffectiveStatus, &instance.EffectiveReason,
			&instance.EffectiveRevision, &instance.CreatedAt, &instance.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Unmarshal evidence
		if evidenceJSON != nil {
			json.Unmarshal(evidenceJSON, &instance.EvidenceJSON)
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// DeleteByAsset removes all finding instances for an asset
func (r *FindingInstanceRepository) DeleteByAsset(ctx context.Context, tenantID, assetID int64) error {
	query := `DELETE FROM finding_instances WHERE tenant_id = $1 AND asset_id = $2`
	_, err := r.db.ExecContext(ctx, query, tenantID, assetID)
	return err
}

// FindingListParams represents query parameters for listing findings
type FindingListParams struct {
	TenantID        int64
	Source          string // Filter by source
	Severity        string // Filter by severity (from definition)
	EffectiveStatus string // Filter by effective_status
	CVE             string // Filter by CVE alias
	AssetName       string // Filter by asset canonical name
	Limit           int    // Max results to return
	Offset          int    // Number of results to skip
}

// FindingListResult represents the result of listing findings
type FindingListResult struct {
	Findings []FindingInstance
	Total    int
	Limit    int
	Offset   int
}

// List retrieves findings with optional filtering and pagination
func (r *FindingInstanceRepository) List(ctx context.Context, params FindingListParams) (*FindingListResult, error) {
	// Build base query with joins for filtering
	baseQuery := `
		FROM finding_instances fi
		INNER JOIN assets a ON fi.asset_id = a.id
		INNER JOIN finding_definitions fd ON fi.definition_uid = fd.definition_uid
		WHERE fi.tenant_id = $1
	`
	args := []interface{}{params.TenantID}
	argCount := 1

	// Add filters
	if params.Source != "" {
		argCount++
		baseQuery += " AND fd.source = $" + string(rune('0'+argCount))
		args = append(args, params.Source)
	}

	if params.Severity != "" {
		argCount++
		baseQuery += " AND fd.severity_default = $" + string(rune('0'+argCount))
		args = append(args, params.Severity)
	}

	if params.EffectiveStatus != "" {
		argCount++
		baseQuery += " AND fi.effective_status = $" + string(rune('0'+argCount))
		args = append(args, params.EffectiveStatus)
	}

	if params.CVE != "" {
		argCount++
		baseQuery += " AND EXISTS (SELECT 1 FROM finding_definition_aliases fda WHERE fda.definition_uid = fi.definition_uid AND fda.alias_type = 'CVE' AND fda.alias_value = $" + string(rune('0'+argCount)) + ")"
		args = append(args, params.CVE)
	}

	if params.AssetName != "" {
		argCount++
		baseQuery += " AND a.canonical_name ILIKE $" + string(rune('0'+argCount))
		args = append(args, "%"+params.AssetName+"%")
	}

	// Get total count
	countQuery := "SELECT COUNT(*)" + baseQuery
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Build select query
	selectQuery := `
		SELECT fi.id, fi.tenant_id, fi.asset_id, fi.definition_uid,
		       fi.scanner_status, fi.first_observed_at, fi.last_observed_at,
		       fi.evidence_json, fi.effective_status, fi.effective_reason,
		       fi.effective_revision, fi.created_at, fi.updated_at
	` + baseQuery + " ORDER BY fi.last_observed_at DESC"

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

	// Query findings
	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []FindingInstance
	for rows.Next() {
		var finding FindingInstance
		var evidenceJSON []byte

		if err := rows.Scan(
			&finding.ID, &finding.TenantID, &finding.AssetID,
			&finding.DefinitionUID, &finding.ScannerStatus,
			&finding.FirstObservedAt, &finding.LastObservedAt,
			&evidenceJSON, &finding.EffectiveStatus, &finding.EffectiveReason,
			&finding.EffectiveRevision, &finding.CreatedAt, &finding.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Unmarshal evidence
		if evidenceJSON != nil {
			json.Unmarshal(evidenceJSON, &finding.EvidenceJSON)
		}

		findings = append(findings, finding)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &FindingListResult{
		Findings: findings,
		Total:    total,
		Limit:    params.Limit,
		Offset:   params.Offset,
	}, nil
}

// EnrichedFindingListResult represents the result of listing enriched findings
type EnrichedFindingListResult struct {
	Findings []EnrichedFinding
	Total    int
	Limit    int
	Offset   int
}

// ListEnriched retrieves findings with optional filtering and pagination, enriched with definition, asset, and intel data
func (r *FindingInstanceRepository) ListEnriched(ctx context.Context, params FindingListParams, includeIntel bool) (*EnrichedFindingListResult, error) {
	// Build base query with joins for filtering
	baseQuery := `
		FROM finding_instances fi
		INNER JOIN assets a ON fi.asset_id = a.id
		INNER JOIN finding_definitions fd ON fi.definition_uid = fd.definition_uid
		WHERE fi.tenant_id = $1
	`
	args := []interface{}{params.TenantID}
	argCount := 1

	// Add filters
	if params.Source != "" {
		argCount++
		baseQuery += " AND fd.source = $" + string(rune('0'+argCount))
		args = append(args, params.Source)
	}

	if params.Severity != "" {
		argCount++
		baseQuery += " AND fd.severity_default = $" + string(rune('0'+argCount))
		args = append(args, params.Severity)
	}

	if params.EffectiveStatus != "" {
		argCount++
		baseQuery += " AND fi.effective_status = $" + string(rune('0'+argCount))
		args = append(args, params.EffectiveStatus)
	}

	if params.CVE != "" {
		argCount++
		baseQuery += " AND EXISTS (SELECT 1 FROM finding_definition_aliases fda WHERE fda.definition_uid = fi.definition_uid AND fda.alias_type = 'CVE' AND fda.alias_value = $" + string(rune('0'+argCount)) + ")"
		args = append(args, params.CVE)
	}

	if params.AssetName != "" {
		argCount++
		baseQuery += " AND a.canonical_name ILIKE $" + string(rune('0'+argCount))
		args = append(args, "%"+params.AssetName+"%")
	}

	// Get total count
	countQuery := "SELECT COUNT(*)" + baseQuery
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Build select query with definition and asset info
	selectQuery := `
		SELECT fi.id, fi.tenant_id, fi.asset_id, fi.definition_uid,
		       fi.scanner_status, fi.first_observed_at, fi.last_observed_at,
		       fi.evidence_json, fi.effective_status, fi.effective_reason,
		       fi.effective_revision, fi.created_at, fi.updated_at,
		       a.canonical_name,
		       fd.title,
		       fd.severity_default
	` + baseQuery + " ORDER BY fi.last_observed_at DESC"

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

	// Query findings
	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []EnrichedFinding
	var allCVEs []string

	for rows.Next() {
		var finding EnrichedFinding
		var evidenceJSON []byte

		if err := rows.Scan(
			&finding.ID, &finding.TenantID, &finding.AssetID,
			&finding.DefinitionUID, &finding.ScannerStatus,
			&finding.FirstObservedAt, &finding.LastObservedAt,
			&evidenceJSON, &finding.EffectiveStatus, &finding.EffectiveReason,
			&finding.EffectiveRevision, &finding.CreatedAt, &finding.UpdatedAt,
			&finding.AssetCanonicalName,
			&finding.DefinitionTitle,
			&finding.DefinitionSeverity,
		); err != nil {
			return nil, err
		}

		// Unmarshal evidence
		if evidenceJSON != nil {
			json.Unmarshal(evidenceJSON, &finding.EvidenceJSON)
		}

		findings = append(findings, finding)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Fetch CVE aliases for all findings
	if len(findings) > 0 {
		defUIDs := make([]string, len(findings))
		for i, f := range findings {
			defUIDs[i] = f.DefinitionUID
		}

		cveMap, err := r.getCVEAliasesByDefinitions(ctx, defUIDs)
		if err != nil {
			return nil, err
		}

		// Attach CVE aliases to findings
		for i := range findings {
			if cves, ok := cveMap[findings[i].DefinitionUID]; ok {
				findings[i].CVEAliases = cves
				allCVEs = append(allCVEs, cves...)
			}
		}
	}

	// Fetch intel data if requested
	if includeIntel && len(allCVEs) > 0 {
		intelRepo := NewIntelRepository(r.db)
		intelMap, err := intelRepo.GetByCVEs(ctx, allCVEs)
		if err != nil {
			return nil, err
		}

		// Attach intel to findings
		for i := range findings {
			if len(findings[i].CVEAliases) > 0 {
				findings[i].Intel = make(map[string]*IntelCVE)
				for _, cve := range findings[i].CVEAliases {
					if intel, ok := intelMap[cve]; ok {
						findings[i].Intel[cve] = intel
					}
				}
			}
		}
	}

	return &EnrichedFindingListResult{
		Findings: findings,
		Total:    total,
		Limit:    params.Limit,
		Offset:   params.Offset,
	}, nil
}

// getCVEAliasesByDefinitions fetches all CVE aliases for given definition UIDs
func (r *FindingInstanceRepository) getCVEAliasesByDefinitions(ctx context.Context, definitionUIDs []string) (map[string][]string, error) {
	if len(definitionUIDs) == 0 {
		return make(map[string][]string), nil
	}

	query, args, err := sqlx.In(`
		SELECT definition_uid, alias_value
		FROM finding_definition_aliases
		WHERE definition_uid IN (?) AND alias_type = 'CVE'
	`, definitionUIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var defUID string
		var cve string
		if err := rows.Scan(&defUID, &cve); err != nil {
			return nil, err
		}
		result[defUID] = append(result[defUID], cve)
	}

	return result, rows.Err()
}
