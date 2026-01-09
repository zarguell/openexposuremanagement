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
	ID                int64                    `db:"id"`
	TenantID          int64                    `db:"tenant_id"`
	AssetID           int64                    `db:"asset_id"`
	DefinitionUID     string                   `db:"definition_uid"`
	ScannerStatus     string                   `db:"scanner_status"`
	FirstObservedAt   time.Time                `db:"first_observed_at"`
	LastObservedAt    time.Time                `db:"last_observed_at"`
	EvidenceJSON      map[string]interface{}   `db:"evidence_json"`
	EffectiveStatus   string                   `db:"effective_status"`
	EffectiveReason   string                   `db:"effective_reason"`
	EffectiveRevision int64                    `db:"effective_revision"`
	CreatedAt         time.Time                `db:"created_at"`
	UpdatedAt         time.Time                `db:"updated_at"`
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
	if instance.EvidenceJSON != nil {
		var err error
		evidenceJSON, err = json.Marshal(instance.EvidenceJSON)
		if err != nil {
			return err
		}
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
