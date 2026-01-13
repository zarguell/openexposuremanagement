package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/openexposuremanagement/oem/internal/database"
	"github.com/openexposuremanagement/oem/internal/repository"
)

// TestUpsertAsset tests the asset upsert functionality
// Following TDD: Tests written BEFORE implementation
func TestUpsertAsset(t *testing.T) {
	// Setup test database
	db := database.SetupTestDB(t)
	if db == nil {
		return // Test was skipped
	}

	ctx := context.Background()
	tenantID := database.CreateTestTenant(t, db)
	source := "tenable"

	t.Run("creates new asset when no match exists", func(t *testing.T) {
		// Use unique hostname and external ID to avoid conflicts with other tests
		uniqueHost := "webserver01-create.example.com"
		uniqueExtID := "i-create1234567890abcdef"

		asset := VMAsset{
			Hostname:    uniqueHost,
			IPAddresses: []string{"192.168.1.10"},
			ExternalIDs: map[string]string{
				"aws:instance_id": uniqueExtID,
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

		if result.Asset.CanonicalName != uniqueHost {
			t.Errorf("expected canonical name '%s', got '%s'", uniqueHost, result.Asset.CanonicalName)
		}

		// Verify identifiers were created
		// Should have: hostname_norm, shortname_norm, ipv4, external_id
		identifiers := getAssetIdentifiers(t, db, result.Asset.ID)

		if len(identifiers) != 4 { // hostname + shortname + IP + external ID
			t.Errorf("expected 4 identifiers, got %d", len(identifiers))
		}
	})

	t.Run("updates existing asset when matched by hostname", func(t *testing.T) {
		// Setup: create existing asset with unique external ID
		uniqueExtID := "i-update1234567890abcdef"
		existingAsset := createTestAsset(t, db, tenantID, "old-hostname.example.com", time.Now().Add(-24*time.Hour))
		createTestIdentifier(t, db, tenantID, existingAsset.ID, "external_id:aws:instance_id", uniqueExtID, source, time.Now().Add(-24*time.Hour))

		asset := VMAsset{
			Hostname: "new-hostname.example.com",
			ExternalIDs: map[string]string{
				"aws:instance_id": uniqueExtID,
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
		// Setup: existing asset with hostname only, use unique external ID
		uniqueHost := "webserver01-add-id.example.com"
		uniqueExtID := "i-addid1234567890abcdef"
		existingAsset := createTestAsset(t, db, tenantID, uniqueHost, time.Now().Add(-24*time.Hour))
		createTestIdentifier(t, db, tenantID, existingAsset.ID, "hostname_norm", uniqueHost, source, time.Now().Add(-24*time.Hour))

		asset := VMAsset{
			Hostname:    uniqueHost,
			IPAddresses: []string{"192.168.1.10"},
			ExternalIDs: map[string]string{
				"aws:instance_id": uniqueExtID,
			},
		}

		// Execute
		result, err := UpsertAsset(ctx, db, tenantID, source, &asset, time.Now())

		// Assertions
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify new identifiers were added
		// Should have: original hostname + shortname + new IP + new external ID
		identifiers := getAssetIdentifiers(t, db, result.Asset.ID)

		if len(identifiers) != 4 { // hostname + shortname + IP + external ID
			t.Errorf("expected 4 identifiers, got %d", len(identifiers))
		}
	})

	t.Run("updates identifier last_seen_at for existing identifiers", func(t *testing.T) {
		// Setup: existing asset with old identifier
		uniqueHost := "webserver01-update-time.example.com"
		oldTime := time.Now().Add(-24 * time.Hour)
		existingAsset := createTestAsset(t, db, tenantID, uniqueHost, oldTime)
		createTestIdentifier(t, db, tenantID, existingAsset.ID, "hostname_norm", uniqueHost, source, oldTime)

		asset := VMAsset{
			Hostname: uniqueHost,
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
		uniqueHost := "webserver01-multi-ip.example.com"

		asset := VMAsset{
			Hostname:    uniqueHost,
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
