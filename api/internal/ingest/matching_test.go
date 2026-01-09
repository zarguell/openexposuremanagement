package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/database"
	"github.com/openexposuremanagement/oem/internal/repository"
)

// TestAssetMatcher_Integration tests the asset matching logic with a real database
func TestAssetMatcher_Integration(t *testing.T) {
	// Setup test database
	db := database.SetupTestDB(t)
	if db == nil {
		return // Test was skipped
	}

	ctx := context.Background()

	t.Run("external_id_match takes precedence", func(t *testing.T) {
		// Each subtest gets its own tenant to avoid conflicts
		tenantID := database.CreateTestTenant(t, db)
		// Create an asset with external ID
		asset := createAssetWithIdentifier(t, db, tenantID, "external_id:aws-id", "i-1234567890abcdef0")

		// Create matcher
		matcher := NewAssetMatcher(db, tenantID, "test-source")

		// Try to match with same external ID
		vmAsset := &VMAsset{
			ExternalIDs: map[string]string{
				"aws-id": "i-1234567890abcdef0",
			},
			Hostname: "different-hostname.com",
		}

		result, err := matcher.MatchAsset(ctx, vmAsset)
		if err != nil {
			t.Fatalf("MatchAsset failed: %v", err)
		}

		if result.NewAsset {
			t.Error("Expected to match existing asset, but got NewAsset=true")
		}

		if result.Reason != MatchReasonExternalID {
			t.Errorf("Expected match reason %s, got %s", MatchReasonExternalID, result.Reason)
		}

		if result.Asset.ID != asset.ID {
			t.Errorf("Expected asset ID %d, got %d", asset.ID, result.Asset.ID)
		}
	})

	t.Run("hostname_match when no external_id", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)
		// Create an asset with hostname
		asset := createAssetWithIdentifier(t, db, tenantID, "hostname_norm", "web-server-01.example.com")

		// Create matcher
		matcher := NewAssetMatcher(db, tenantID, "test-source")

		// Try to match with hostname
		vmAsset := &VMAsset{
			Hostname: "web-server-01.example.com",
		}

		result, err := matcher.MatchAsset(ctx, vmAsset)
		if err != nil {
			t.Fatalf("MatchAsset failed: %v", err)
		}

		if result.NewAsset {
			t.Error("Expected to match existing asset, but got NewAsset=true")
		}

		if result.Reason != MatchReasonHostname {
			t.Errorf("Expected match reason %s, got %s", MatchReasonHostname, result.Reason)
		}

		if result.Asset.ID != asset.ID {
			t.Errorf("Expected asset ID %d, got %d", asset.ID, result.Asset.ID)
		}
	})

	t.Run("shortname_match when enabled", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		// Create an asset with ONLY a shortname identifier (no hostname)
		// This simulates a DHCP environment where we only know the shortname
		scannedAt := time.Now().UTC()
		var assetID int64

		// Use a unique canonical name that won't match the test hostname
		canonicalName := "asset-with-only-shortname-" + scannedAt.Format("150405000")

		err := db.QueryRow(
			`INSERT INTO assets (tenant_id, canonical_name, first_seen_at, last_seen_at, is_active)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			tenantID, canonicalName, scannedAt, scannedAt, true,
		).Scan(&assetID)
		if err != nil {
			t.Fatalf("Failed to create test asset: %v", err)
		}

		// Add ONLY the shortname identifier (no hostname_norm)
		_, err = db.Exec(
			`INSERT INTO asset_identifiers (tenant_id, asset_id, id_type, id_value, first_seen_at, last_seen_at, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			tenantID, assetID, "shortname_norm", "web-server-01", scannedAt, scannedAt, "test",
		)
		if err != nil {
			t.Fatalf("Failed to create identifier: %v", err)
		}

		// Create matcher with shortnames enabled
		matcher := NewAssetMatcher(db, tenantID, "test-source")
		matcher.EnableShortnames()

		// Try to match with hostname that has same shortname
		vmAsset := &VMAsset{
			Hostname: "web-server-01.example.com",
		}

		result, err := matcher.MatchAsset(ctx, vmAsset)
		if err != nil {
			t.Fatalf("MatchAsset failed: %v", err)
		}

		if result.NewAsset {
			t.Error("Expected to match existing asset, but got NewAsset=true")
		}

		if result.Reason != MatchReasonShortname {
			t.Errorf("Expected match reason %s, got %s", MatchReasonShortname, result.Reason)
		}

		if result.Asset.ID != assetID {
			t.Errorf("Expected asset ID %d, got %d", assetID, result.Asset.ID)
		}
	})

	t.Run("shortname_not_used_when_disabled", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		// Create an asset with ONLY a shortname identifier (no hostname)
		scannedAt := time.Now().UTC()
		var assetID int64
		canonicalName := "asset-no-shortname-match-" + scannedAt.Format("150405000")

		err := db.QueryRow(
			`INSERT INTO assets (tenant_id, canonical_name, first_seen_at, last_seen_at, is_active)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			tenantID, canonicalName, scannedAt, scannedAt, true,
		).Scan(&assetID)
		if err != nil {
			t.Fatalf("Failed to create test asset: %v", err)
		}

		// Add ONLY the shortname identifier (no hostname_norm)
		_, err = db.Exec(
			`INSERT INTO asset_identifiers (tenant_id, asset_id, id_type, id_value, first_seen_at, last_seen_at, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			tenantID, assetID, "shortname_norm", "web-server-02", scannedAt, scannedAt, "test",
		)
		if err != nil {
			t.Fatalf("Failed to create identifier: %v", err)
		}

		// Create matcher WITHOUT shortnames enabled (default)
		matcher := NewAssetMatcher(db, tenantID, "test-source")

		// Try to match with hostname
		vmAsset := &VMAsset{
			Hostname: "web-server-02.example.com",
		}

		result, err := matcher.MatchAsset(ctx, vmAsset)
		if err != nil {
			t.Fatalf("MatchAsset failed: %v", err)
		}

		// Should NOT match (shortname matching disabled by default)
		if !result.NewAsset {
			t.Error("Expected no match with shortname disabled, but got existing asset")
		}

		if result.Reason != MatchReasonNoMatch {
			t.Errorf("Expected match reason %s, got %s", MatchReasonNoMatch, result.Reason)
		}
	})

	t.Run("ip_and_hostname_match", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		// Create an asset with ONLY IP and hostname identifiers (but different canonical name)
		// This tests that when we have both IP and hostname, and they match, we find the asset
		scannedAt := time.Now().UTC()
		var assetID int64

		// Use a different canonical name so hostname matching alone won't work
		canonicalName := "server-with-ip-" + scannedAt.Format("150405000")

		err := db.QueryRow(
			`INSERT INTO assets (tenant_id, canonical_name, first_seen_at, last_seen_at, is_active)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			tenantID, canonicalName, scannedAt, scannedAt, true,
		).Scan(&assetID)
		if err != nil {
			t.Fatalf("Failed to create test asset: %v", err)
		}

		// Add hostname identifier
		_, err = db.Exec(
			`INSERT INTO asset_identifiers (tenant_id, asset_id, id_type, id_value, first_seen_at, last_seen_at, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			tenantID, assetID, "hostname_norm", "web-server-03.example.com", scannedAt, scannedAt, "test",
		)
		if err != nil {
			t.Fatalf("Failed to create identifier: %v", err)
		}

		// Add IP identifier to the same asset
		_, err = db.Exec(
			`INSERT INTO asset_identifiers (tenant_id, asset_id, id_type, id_value, first_seen_at, last_seen_at, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			tenantID, assetID, "ipv4", "192.168.1.102", scannedAt, scannedAt, "test",
		)
		if err != nil {
			t.Fatalf("Failed to create identifier: %v", err)
		}

		// Create matcher
		matcher := NewAssetMatcher(db, tenantID, "test-source")

		// Try to match with same IP+hostname
		vmAsset := &VMAsset{
			Hostname:     "web-server-03.example.com",
			IPAddresses:  []string{"192.168.1.102"},
		}

		result, err := matcher.MatchAsset(ctx, vmAsset)
		if err != nil {
			t.Fatalf("MatchAsset failed: %v", err)
		}

		if result.NewAsset {
			t.Error("Expected to match existing asset, but got NewAsset=true")
		}

		// When both IP and hostname are provided in the incoming data, and both match,
		// it will use the IP+hostname match reason (which is checked after hostname in priority)
		// This verifies the IP+hostname matching logic works
		if result.Reason != MatchReasonIPAndHost {
			t.Errorf("Expected match reason %s, got %s", MatchReasonIPAndHost, result.Reason)
		}

		if result.Asset.ID != assetID {
			t.Errorf("Expected asset ID %d, got %d", assetID, result.Asset.ID)
		}
	})

	t.Run("ip_only_does_not_match", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)
		// Create an asset with IP
		_ = createAssetWithIdentifier(t, db, tenantID, "ipv4", "192.168.1.101")

		// Create matcher
		matcher := NewAssetMatcher(db, tenantID, "test-source")

		// Try to match with IP only (no hostname) - should NOT match
		vmAsset := &VMAsset{
			IPAddresses: []string{"192.168.1.101"},
		}

		result, err := matcher.MatchAsset(ctx, vmAsset)
		if err != nil {
			t.Fatalf("MatchAsset failed: %v", err)
		}

		// IP-only should NOT match
		if !result.NewAsset {
			t.Error("IP-only should not match, but got existing asset")
		}

		if result.Reason != MatchReasonNoMatch {
			t.Errorf("Expected match reason %s, got %s", MatchReasonNoMatch, result.Reason)
		}
	})

	t.Run("no_match_when_asset_doesnt_exist", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)
		// Create matcher
		matcher := NewAssetMatcher(db, tenantID, "test-source")

		// Try to match non-existent asset
		vmAsset := &VMAsset{
			Hostname:    "nonexistent.example.com",
			ExternalIDs: map[string]string{"aws-id": "i-nonexistent"},
		}

		result, err := matcher.MatchAsset(ctx, vmAsset)
		if err != nil {
			t.Fatalf("MatchAsset failed: %v", err)
		}

		if !result.NewAsset {
			t.Error("Expected NewAsset=true for non-existent asset")
		}

		if result.Asset != nil {
			t.Error("Expected nil Asset for non-existent match")
		}

		if result.Reason != MatchReasonNoMatch {
			t.Errorf("Expected match reason %s, got %s", MatchReasonNoMatch, result.Reason)
		}
	})

	t.Run("matching_priority_order", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		// Create multiple assets with different identifiers
		// Use unique names to avoid conflicts
		scannedAt := time.Now().UTC()

		// Asset 1: Only hostname
		var asset1ID int64
		canonicalName1 := "server-hostname-" + scannedAt.Format("150405000")
		err := db.QueryRow(
			`INSERT INTO assets (tenant_id, canonical_name, first_seen_at, last_seen_at, is_active)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			tenantID, canonicalName1, scannedAt, scannedAt, true,
		).Scan(&asset1ID)
		if err != nil {
			t.Fatalf("Failed to create test asset: %v", err)
		}

		_, err = db.Exec(
			`INSERT INTO asset_identifiers (tenant_id, asset_id, id_type, id_value, first_seen_at, last_seen_at, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			tenantID, asset1ID, "hostname_norm", "priority.example.com", scannedAt, scannedAt, "test",
		)
		if err != nil {
			t.Fatalf("Failed to create identifier: %v", err)
		}

		// Asset 2: External ID (and we'll add hostname to it)
		var asset2ID int64
		canonicalName2 := "server-external-" + scannedAt.Format("150405001")
		err = db.QueryRow(
			`INSERT INTO assets (tenant_id, canonical_name, first_seen_at, last_seen_at, is_active)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			tenantID, canonicalName2, scannedAt, scannedAt, true,
		).Scan(&asset2ID)
		if err != nil {
			t.Fatalf("Failed to create test asset: %v", err)
		}

		// Add external ID identifier
		_, err = db.Exec(
			`INSERT INTO asset_identifiers (tenant_id, asset_id, id_type, id_value, first_seen_at, last_seen_at, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			tenantID, asset2ID, "external_id:aws-id", "i-priority12345", scannedAt, scannedAt, "test",
		)
		if err != nil {
			t.Fatalf("Failed to create identifier: %v", err)
		}

		// Add hostname identifier to asset2 as well (same hostname as asset1)
		_, err = db.Exec(
			`INSERT INTO asset_identifiers (tenant_id, asset_id, id_type, id_value, first_seen_at, last_seen_at, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			tenantID, asset2ID, "hostname_norm", "priority.example.com", scannedAt, scannedAt, "test",
		)
		if err != nil {
			t.Fatalf("Failed to create identifier: %v", err)
		}

		// Create matcher
		matcher := NewAssetMatcher(db, tenantID, "test-source")

		// Asset with both external ID and hostname - should match by external ID first (priority)
		vmAsset := &VMAsset{
			ExternalIDs: map[string]string{
				"aws-id": "i-priority12345",
			},
			Hostname: "priority.example.com",
		}

		result, err := matcher.MatchAsset(ctx, vmAsset)
		if err != nil {
			t.Fatalf("MatchAsset failed: %v", err)
		}

		// Should match by external ID (higher priority), not hostname
		if result.Asset.ID != asset2ID {
			t.Errorf("Expected to match by external ID (asset %d), got asset %d", asset2ID, result.Asset.ID)
		}

		if result.Reason != MatchReasonExternalID {
			t.Errorf("Expected match reason %s, got %s", MatchReasonExternalID, result.Reason)
		}
	})
}

// createAssetWithIdentifier creates a test asset with a specific identifier
func createAssetWithIdentifier(t *testing.T, db *sqlx.DB, tenantID int64, idType, idValue string) *repository.Asset {
	t.Helper()
	return createAssetWithCanonicalName(t, db, tenantID, idType, idValue, "")
}

// createAssetWithCanonicalName creates a test asset with a specific identifier and canonical name
func createAssetWithCanonicalName(t *testing.T, db *sqlx.DB, tenantID int64, idType, idValue, canonicalName string) *repository.Asset {
	t.Helper()

	// First create the asset
	scannedAt := time.Now().UTC()
	var assetID int64

	// Determine canonical name based on identifier type if not provided
	if canonicalName == "" {
		// For hostname identifiers, use the actual value as canonical name
		if idType == "hostname_norm" {
			canonicalName = idValue
		} else if idType == "ipv4" {
			canonicalName = idValue
		} else if idType == "shortname_norm" {
			canonicalName = idValue + ".example.com"
		} else {
			// For external IDs, generate a unique name
			randomSuffix := scannedAt.Format("150405000") // Add milliseconds for uniqueness
			canonicalName = "asset-" + idType + "-" + randomSuffix + ".example.com"
		}
	}

	err := db.QueryRow(
		`INSERT INTO assets (tenant_id, canonical_name, first_seen_at, last_seen_at, is_active)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		tenantID, canonicalName, scannedAt, scannedAt, true,
	).Scan(&assetID)

	if err != nil {
		t.Fatalf("Failed to create test asset: %v", err)
	}

	// Add the identifier
	_, err = db.Exec(
		`INSERT INTO asset_identifiers (tenant_id, asset_id, id_type, id_value, first_seen_at, last_seen_at, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tenantID, assetID, idType, idValue, scannedAt, scannedAt, "test",
	)

	if err != nil {
		t.Fatalf("Failed to create identifier: %v", err)
	}

	return &repository.Asset{
		ID:            assetID,
		TenantID:      tenantID,
		CanonicalName: canonicalName,
		FirstSeenAt:   scannedAt,
		LastSeenAt:    scannedAt,
		IsActive:      true,
	}
}
