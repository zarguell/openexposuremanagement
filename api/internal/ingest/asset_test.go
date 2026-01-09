package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/openexposuremanagement/oem/internal/repository"
)

// TestUpsertAsset tests the asset upsert functionality
// Following TDD: Tests written BEFORE implementation
func TestUpsertAsset(t *testing.T) {
	// TODO: Set up test database
	t.Skip("Test database setup needed - writing tests first")

	ctx := context.Background()
	db := setupTestDB(t)
	tenantID := int64(1)
	source := "tenable"

	t.Run("creates new asset when no match exists", func(t *testing.T) {
		// Setup: clean state
		cleanTestDB(t, db)

		asset := VMAsset{
			Hostname:    "webserver01.example.com",
			IPAddresses: []string{"192.168.1.10"},
			ExternalIDs: map[string]string{
				"aws:instance_id": "i-1234567890abcdef0",
			},
		}

		// Execute
		result, err := UpsertAsset(ctx, db, tenantID, source, &asset, time.Now())

		// Assertions
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !result.NewAsset {
			t.Error("expected NewAsset to be true for new asset")
		}

		if result.Asset == nil {
			t.Fatal("expected asset to be returned, got nil")
		}

		if result.Asset.CanonicalName != "webserver01.example.com" {
			t.Errorf("expected canonical name 'webserver01.example.com', got '%s'", result.Asset.CanonicalName)
		}

		// Verify identifiers were created
		identifiers := getAssetIdentifiers(t, db, result.Asset.ID)

		if len(identifiers) != 2 { // hostname + external ID
			t.Errorf("expected 2 identifiers, got %d", len(identifiers))
		}
	})

	t.Run("updates existing asset when matched by hostname", func(t *testing.T) {
		// Setup: create existing asset
		cleanTestDB(t, db)
		existingAsset := createTestAsset(t, db, tenantID, "webserver01.example.com", time.Now().Add(-24*time.Hour))
		createTestIdentifier(t, db, tenantID, existingAsset.ID, "hostname_norm", "webserver01.example.com", source, time.Now().Add(-24*time.Hour))

		asset := VMAsset{
			Hostname:    "WebServer01.Example.COM", // Different case
			IPAddresses: []string{"192.168.1.10"},
		}

		// Execute
		scannedAt := time.Now()
		result, err := UpsertAsset(ctx, db, tenantID, source, &asset, scannedAt)

		// Assertions
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.NewAsset {
			t.Error("expected NewAsset to be false for existing asset")
		}

		if result.Asset.ID != existingAsset.ID {
			t.Errorf("expected existing asset ID %d, got %d", existingAsset.ID, result.Asset.ID)
		}

		// Verify last_seen_at was updated
		if result.Asset.LastSeenAt.Before(scannedAt) || result.Asset.LastSeenAt.After(time.Now().Add(1*time.Minute)) {
			t.Errorf("expected last_seen_at to be close to scanned time, got %v", result.Asset.LastSeenAt)
		}

		// Verify first_seen_at was NOT updated
		if result.Asset.FirstSeenAt.After(existingAsset.FirstSeenAt.Add(1 * time.Minute)) {
			t.Errorf("expected first_seen_at to remain unchanged, got %v (old: %v)", result.Asset.FirstSeenAt, existingAsset.FirstSeenAt)
		}
	})

	t.Run("updates existing asset when matched by external ID", func(t *testing.T) {
		// Setup: create existing asset with external ID
		cleanTestDB(t, db)
		existingAsset := createTestAsset(t, db, tenantID, "old-hostname.example.com", time.Now().Add(-24*time.Hour))
		createTestIdentifier(t, db, tenantID, existingAsset.ID, "external_id:aws:instance_id", "i-1234567890abcdef0", source, time.Now().Add(-24*time.Hour))

		asset := VMAsset{
			Hostname: "new-hostname.example.com",
			ExternalIDs: map[string]string{
				"aws:instance_id": "i-1234567890abcdef0",
			},
		}

		// Execute
		result, err := UpsertAsset(ctx, db, tenantID, source, &asset, time.Now())

		// Assertions
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.NewAsset {
			t.Error("expected NewAsset to be false for existing asset")
		}

		if result.Asset.ID != existingAsset.ID {
			t.Errorf("expected existing asset ID %d, got %d", existingAsset.ID, result.Asset.ID)
		}

		// Verify canonical_name was updated to new hostname
		if result.Asset.CanonicalName != "new-hostname.example.com" {
			t.Errorf("expected canonical name to update to 'new-hostname.example.com', got '%s'", result.Asset.CanonicalName)
		}
	})

	t.Run("adds new identifier to existing asset", func(t *testing.T) {
		// Setup: existing asset with hostname only
		cleanTestDB(t, db)
		existingAsset := createTestAsset(t, db, tenantID, "webserver01.example.com", time.Now().Add(-24*time.Hour))
		createTestIdentifier(t, db, tenantID, existingAsset.ID, "hostname_norm", "webserver01.example.com", source, time.Now().Add(-24*time.Hour))

		asset := VMAsset{
			Hostname:    "webserver01.example.com",
			IPAddresses: []string{"192.168.1.10"},
			ExternalIDs: map[string]string{
				"aws:instance_id": "i-1234567890abcdef0",
			},
		}

		// Execute
		result, err := UpsertAsset(ctx, db, tenantID, source, &asset, time.Now())

		// Assertions
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify new identifiers were added
		identifiers := getAssetIdentifiers(t, db, result.Asset.ID)

		// Should have original hostname + new IP + new external ID
		if len(identifiers) != 3 {
			t.Errorf("expected 3 identifiers, got %d", len(identifiers))
		}
	})

	t.Run("updates identifier last_seen_at for existing identifiers", func(t *testing.T) {
		// Setup: existing asset with old identifier
		cleanTestDB(t, db)
		oldTime := time.Now().Add(-24 * time.Hour)
		existingAsset := createTestAsset(t, db, tenantID, "webserver01.example.com", oldTime)
		createTestIdentifier(t, db, tenantID, existingAsset.ID, "hostname_norm", "webserver01.example.com", source, oldTime)

		asset := VMAsset{
			Hostname: "webserver01.example.com",
		}

		// Execute
		scannedAt := time.Now()
		result, err := UpsertAsset(ctx, db, tenantID, source, &asset, scannedAt)

		// Assertions
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify identifier last_seen_at was updated
		identifiers := getAssetIdentifiers(t, db, result.Asset.ID)

		var hostnameID *AssetIdentifier
		for _, id := range identifiers {
			if id.IDType == "hostname_norm" {
				hostnameID = &id
				break
			}
		}

		if hostnameID == nil {
			t.Fatal("expected to find hostname identifier")
		}

		if hostnameID.LastSeenAt.Before(scannedAt) || hostnameID.LastSeenAt.After(time.Now().Add(1*time.Minute)) {
			t.Errorf("expected identifier last_seen_at to be updated to scanned time, got %v", hostnameID.LastSeenAt)
		}

		if hostnameID.FirstSeenAt.After(oldTime.Add(1 * time.Minute)) {
			t.Errorf("expected identifier first_seen_at to remain unchanged, got %v", hostnameID.FirstSeenAt)
		}
	})

	t.Run("handles multiple IP addresses", func(t *testing.T) {
		// Setup: clean state
		cleanTestDB(t, db)

		asset := VMAsset{
			Hostname:    "webserver01.example.com",
			IPAddresses: []string{"192.168.1.10", "192.168.1.11", "10.0.0.5"},
		}

		// Execute
		result, err := UpsertAsset(ctx, db, tenantID, source, &asset, time.Now())

		// Assertions
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify all IPs were added as identifiers
		identifiers := getAssetIdentifiers(t, db, result.Asset.ID)

		ipCount := 0
		for _, id := range identifiers {
			if id.IDType == "ipv4" {
				ipCount++
			}
		}

		if ipCount != 3 {
			t.Errorf("expected 3 IP identifiers, got %d", ipCount)
		}
	})

	t.Run("normalizes hostname and creates normalized identifiers", func(t *testing.T) {
		// Setup: clean state
		cleanTestDB(t, db)

		asset := VMAsset{
			Hostname: "  WebServer01.Example.COM.  ", // Has spaces, dots, mixed case
		}

		// Execute
		result, err := UpsertAsset(ctx, db, tenantID, source, &asset, time.Now())

		// Assertions
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify canonical name is normalized
		if result.Asset.CanonicalName != "webserver01.example.com" {
			t.Errorf("expected normalized canonical name 'webserver01.example.com', got '%s'", result.Asset.CanonicalName)
		}

		// Verify normalized identifier
		identifiers := getAssetIdentifiers(t, db, result.Asset.ID)

		var hostnameID *AssetIdentifier
		for _, id := range identifiers {
			if id.IDType == "hostname_norm" {
				hostnameID = &id
				break
			}
		}

		if hostnameID == nil {
			t.Fatal("expected to find hostname identifier")
		}

		if hostnameID.IDValue != "webserver01.example.com" {
			t.Errorf("expected normalized hostname value 'webserver01.example.com', got '%s'", hostnameID.IDValue)
		}
	})
}

// Helper functions for tests

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sqlx.DB {
	// TODO: Implement actual test DB setup
	// For now, this is a placeholder to make tests compile
	t.Helper()
	return nil
}

// cleanTestDB cleans the test database
func cleanTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.Exec("DELETE FROM asset_identifiers")
	if err != nil {
		t.Fatalf("failed to clean asset_identifiers: %v", err)
	}
	_, err = db.Exec("DELETE FROM assets")
	if err != nil {
		t.Fatalf("failed to clean assets: %v", err)
	}
}

// createTestAsset creates a test asset in the database
func createTestAsset(t *testing.T, db *sqlx.DB, tenantID int64, canonicalName string, seenAt time.Time) *repository.Asset {
	t.Helper()
	asset := &repository.Asset{
		TenantID:      tenantID,
		CanonicalName: canonicalName,
		FirstSeenAt:   seenAt,
		LastSeenAt:    seenAt,
		IsActive:      true,
	}

	err := db.QueryRow(
		`INSERT INTO assets (tenant_id, canonical_name, first_seen_at, last_seen_at, is_active)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		asset.TenantID, asset.CanonicalName, asset.FirstSeenAt, asset.LastSeenAt, asset.IsActive,
	).Scan(&asset.ID, &asset.CreatedAt, &asset.UpdatedAt)

	if err != nil {
		t.Fatalf("failed to create test asset: %v", err)
	}

	return asset
}

// createTestIdentifier creates a test asset identifier in the database
func createTestIdentifier(t *testing.T, db *sqlx.DB, tenantID, assetID int64, idType, idValue, source string, seenAt time.Time) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO asset_identifiers (tenant_id, asset_id, id_type, id_value, first_seen_at, last_seen_at, source)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tenantID, assetID, idType, idValue, seenAt, seenAt, source,
	)

	if err != nil {
		t.Fatalf("failed to create test identifier: %v", err)
	}
}

// getAssetIdentifiers retrieves all identifiers for an asset
func getAssetIdentifiers(t *testing.T, db *sqlx.DB, assetID int64) []AssetIdentifier {
	t.Helper()

	var identifiers []AssetIdentifier
	err := db.Select(&identifiers, "SELECT * FROM asset_identifiers WHERE asset_id = $1", assetID)
	if err != nil {
		t.Fatalf("failed to get identifiers: %v", err)
	}

	return identifiers
}

// AssetIdentifier is a test-only struct for querying identifiers
type AssetIdentifier struct {
	ID          int64     `db:"id"`
	TenantID    int64     `db:"tenant_id"`
	AssetID     int64     `db:"asset_id"`
	IDType      string    `db:"id_type"`
	IDValue     string    `db:"id_value"`
	FirstSeenAt time.Time `db:"first_seen_at"`
	LastSeenAt  time.Time `db:"last_seen_at"`
	Source      string    `db:"source"`
}
