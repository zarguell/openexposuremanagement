package ingest

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// ConvertToFindingInstance converts a VMFinding to a FindingInstance
func ConvertToFindingInstance(tenantID, assetID int64, definitionUID string, vmFinding *VMFinding) *repository.FindingInstance {
	// Determine observation window timestamps
	firstFound := vmFinding.FirstFound
	lastFound := vmFinding.LastFound

	// If first_found is not set, use last_found
	if firstFound.IsZero() {
		firstFound = lastFound
	}

	// Map scanner status to normalized value
	scannerStatus := NormalizeScannerStatus(vmFinding.Status)

	// Compute effective status (baseline - no suppressions yet)
	effectiveStatus := ComputeEffectiveStatus(scannerStatus, nil)
	effectiveReason := "scanner" // Default reason

	return &repository.FindingInstance{
		TenantID:          tenantID,
		AssetID:           assetID,
		DefinitionUID:     definitionUID,
		ScannerStatus:     scannerStatus,
		FirstObservedAt:   firstFound,
		LastObservedAt:    lastFound,
		EvidenceJSON:      vmFinding.Evidence,
		EffectiveStatus:   effectiveStatus,
		EffectiveReason:   effectiveReason,
		EffectiveRevision: 0, // TODO: Get from tenant_policy_state
	}
}

// NormalizeScannerStatus normalizes a scanner status value
func NormalizeScannerStatus(status string) string {
	// For MVP, we accept the status as-is from valid values
	// Valid values: open, fixed, fixed_by_verification
	return status
}

// ComputeEffectiveStatus computes the effective status based on scanner status and suppressions
// For now (baseline), effective status = scanner status (no suppressions yet)
func ComputeEffectiveStatus(scannerStatus string, suppression interface{}) string {
	// If there's an active suppression, it would override the scanner status
	// TODO: Implement suppression lookup in future task

	// Map scanner status to effective status
	switch scannerStatus {
	case "open":
		return "open"
	case "fixed", "fixed_by_verification":
		return "fixed"
	default:
		// Unknown statuses default to "open"
		return "open"
	}
}

// UpsertFindingInstance upserts a finding instance with observation window tracking
func UpsertFindingInstance(ctx context.Context, db *sqlx.DB, tenantID, assetID int64, definitionUID string, vmFinding *VMFinding, policyRevision int64) error {
	// Convert to finding instance
	instance := ConvertToFindingInstance(tenantID, assetID, definitionUID, vmFinding)
	instance.EffectiveRevision = policyRevision

	// Upsert finding instance
	repo := repository.NewFindingInstanceRepository(db)
	err := repo.UpsertFindingInstance(ctx, instance)
	if err != nil {
		log.Error().
			Int64("tenant_id", tenantID).
			Int64("asset_id", assetID).
			Str("definition_uid", definitionUID).
			Err(err).
			Msg("Failed to upsert finding instance")
		return err
	}

	log.Debug().
		Int64("tenant_id", tenantID).
		Int64("asset_id", assetID).
		Str("definition_uid", definitionUID).
		Str("scanner_status", instance.ScannerStatus).
		Str("effective_status", instance.EffectiveStatus).
		Msg("Upserted finding instance")

	return nil
}
