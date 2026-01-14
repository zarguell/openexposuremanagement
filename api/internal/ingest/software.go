package ingest

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// UpsertSoftwareForAsset processes installed software for an asset
// This performs:
// 1. Upserts each software product to the software table
// 2. Upserts asset_software relations
// 3. Deletes software not seen in the current scan
func UpsertSoftwareForAsset(ctx context.Context, db *sqlx.DB, tenantID, assetID int64, softwareList []InstalledSoftware, source string) (*SoftwareUpsertResult, error) {
	result := &SoftwareUpsertResult{
		TotalSoftware: len(softwareList),
	}

	if len(softwareList) == 0 {
		return result, nil
	}

	// Track which software IDs we see in this scan
	seenSoftwareIDs := make(map[int64]bool)

	// Process each software item
	for _, sw := range softwareList {
		// 1. Get or create software record
		softwareID, isNew, err := upsertSoftware(ctx, db, sw)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert software: %w", err)
		}

		if isNew {
			result.SoftwareCreated++
		}
		result.SoftwareUpserted++

		// 2. Create or update asset_software relation
		relationCreated, err := upsertAssetSoftware(ctx, db, tenantID, assetID, softwareID, source, sw.InstallPath)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert asset_software: %w", err)
		}

		if relationCreated {
			result.RelationsCreated++
		} else {
			result.RelationsUpdated++
		}

		// Mark this software as seen
		seenSoftwareIDs[softwareID] = true
	}

	// 3. Delete software not seen in this scan
	deleted, err := deleteUnseenSoftware(ctx, db, tenantID, assetID, seenSoftwareIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to delete unseen software: %w", err)
	}

	result.RelationsDeleted = deleted

	return result, nil
}

// upsertSoftware inserts or updates a software record
// Returns software ID and whether it was newly created
func upsertSoftware(ctx context.Context, db *sqlx.DB, sw InstalledSoftware) (int64, bool, error) {
	cpe := sw.GetCPE()
	title := sw.GetFormattedTitle()

	// Check if software already exists
	var existingID int64
	err := db.GetContext(ctx, &existingID,
		"SELECT id FROM software WHERE cpe_string = $1",
		cpe)

	if err == nil {
		// Software exists, update it if needed
		_, err = db.ExecContext(ctx,
			`UPDATE software
				SET title_formatted = $2,
				    updated_at = NOW()
				WHERE id = $1`,
			existingID, title)

		return existingID, false, err
	}

	// Software doesn't exist, insert it
	var newID int64
	err = db.QueryRowContext(ctx,
		`INSERT INTO software (cpe_string, vendor, product_name, version, edition, title_formatted)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (cpe_string) DO UPDATE
				SET title_formatted = EXCLUDED.title_formatted,
				    updated_at = NOW()
			RETURNING id`,
		cpe, sw.Vendor, sw.Product, sw.Version, sw.Edition, title).Scan(&newID)

	if err != nil {
		return 0, false, fmt.Errorf("failed to insert software: %w", err)
	}

	return newID, true, nil
}

// upsertAssetSoftware creates or updates an asset_software relation
// Returns whether the relation was newly created
func upsertAssetSoftware(ctx context.Context, db *sqlx.DB, tenantID, assetID, softwareID int64, source, installPath string) (bool, error) {
	// Check if relation exists
	var existingID int64
	err := db.GetContext(ctx, &existingID,
		`SELECT id FROM asset_software
			WHERE tenant_id = $1 AND asset_id = $2 AND software_id = $3`,
		tenantID, assetID, softwareID)

	if err == nil {
		// Relation exists, update it
		_, err = db.ExecContext(ctx,
			`UPDATE asset_software
				SET last_seen_at = NOW(),
				    source = $1,
				    install_path = $2,
				    updated_at = NOW()
				WHERE id = $3`,
			source, installPath, existingID)

		return false, err
	}

	// Relation doesn't exist, insert it
	_, err = db.ExecContext(ctx,
		`INSERT INTO asset_software (tenant_id, asset_id, software_id, source, install_path)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (tenant_id, asset_id, software_id) DO UPDATE
				SET last_seen_at = NOW(),
				    source = EXCLUDED.source,
				    install_path = EXCLUDED.install_path,
				    updated_at = NOW()`,
		tenantID, assetID, softwareID, source, installPath)

	if err != nil {
		return false, fmt.Errorf("failed to insert asset_software: %w", err)
	}

	return true, nil
}

// deleteUnseenSoftware deletes asset_software relations not seen in the current scan
func deleteUnseenSoftware(ctx context.Context, db *sqlx.DB, tenantID, assetID int64, seenSoftwareIDs map[int64]bool) (int, error) {
	if len(seenSoftwareIDs) == 0 {
		// Delete all software for this asset
		result, err := db.ExecContext(ctx,
			`DELETE FROM asset_software
			WHERE tenant_id = $1 AND asset_id = $2`,
			tenantID, assetID)

		if err != nil {
			return 0, err
		}

		count, _ := result.RowsAffected()
		return int(count), nil
	}

	// Delete only software not in the seen list
	// Build a list of unseen software IDs
	var unseenIDs []int64
	for id := range seenSoftwareIDs {
		unseenIDs = append(unseenIDs, id)
	}

	// Delete software not in the seen list
	query := `DELETE FROM asset_software
		WHERE tenant_id = $1 AND asset_id = $2 AND software_id NOT IN (`
	args := []interface{}{tenantID, assetID}

	for i, id := range unseenIDs {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("$%d", len(args)+1)
		args = append(args, id)
	}
	query += ")"

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete unseen software: %w", err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}
