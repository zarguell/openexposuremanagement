package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/openexposuremanagement/oem/internal/database"
)

// This file uses TDD approach for Asset repository
// Tests are written FIRST, then implementation

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sqlx.DB {
	return database.SetupTestDB(t)
}

func TestAssetRepository_GetByID(t *testing.T) {
	t.Run("returns asset when it exists", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Create test tenant
		tenantID := database.CreateTestTenant(t, db)

		// Create test asset
		_, err := db.Exec(
			`INSERT INTO assets (tenant_id, canonical_name, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())`,
			tenantID, "test-asset.example.com",
		)
		if err != nil {
			t.Fatalf("Failed to create test asset: %v", err)
		}

		var assetID int64
		err = db.QueryRow("SELECT id FROM assets WHERE canonical_name = $1", "test-asset.example.com").Scan(&assetID)
		if err != nil {
			t.Fatalf("Failed to get asset ID: %v", err)
		}

		// Test: Get by ID
		result, err := repo.GetByID(ctx, assetID)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("expected asset, got nil")
		}
		if result.ID != assetID {
			t.Errorf("expected ID %d, got %d", assetID, result.ID)
		}
		if result.CanonicalName != "test-asset.example.com" {
			t.Errorf("expected canonical name 'test-asset.example.com', got %s", result.CanonicalName)
		}
	})

	t.Run("returns error when asset doesn't exist", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Test: Get non-existent asset
		result, err := repo.GetByID(ctx, 999)

		// Assert
		if err == nil {
			t.Error("expected error, got nil")
		}
		if result != nil {
			t.Error("expected nil asset, got asset")
		}
	})
}

func TestAssetRepository_GetByCanonicalName(t *testing.T) {
	t.Run("returns asset for tenant and canonical name", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Create test tenant
		tenantID := database.CreateTestTenant(t, db)

		// Create test asset
		canonicalName := "test-canonical.example.com"
		_, err := db.Exec(
			`INSERT INTO assets (tenant_id, canonical_name, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())`,
			tenantID, canonicalName,
		)
		if err != nil {
			t.Fatalf("Failed to create test asset: %v", err)
		}

		// Test
		result, err := repo.GetByCanonicalName(ctx, tenantID, canonicalName)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("expected asset, got nil")
		}
		if result.TenantID != tenantID {
			t.Errorf("expected tenant ID %d, got %d", tenantID, result.TenantID)
		}
		if result.CanonicalName != canonicalName {
			t.Errorf("expected canonical name '%s', got '%s'", canonicalName, result.CanonicalName)
		}
	})

	t.Run("returns error for wrong tenant", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Create test tenant and asset
		tenantID := database.CreateTestTenant(t, db)
		canonicalName := "test-tenant-isolation.example.com"
		_, err := db.Exec(
			`INSERT INTO assets (tenant_id, canonical_name, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())`,
			tenantID, canonicalName,
		)
		if err != nil {
			t.Fatalf("Failed to create test asset: %v", err)
		}

		// Test: Try to get asset from different tenant
		result, err := repo.GetByCanonicalName(ctx, tenantID+1, canonicalName)

		// Assert
		if err == nil {
			t.Error("expected error for wrong tenant, got nil")
		}
		if result != nil {
			t.Error("expected nil asset for wrong tenant, got asset")
		}
	})
}

func TestAssetRepository_Create(t *testing.T) {
	t.Run("creates asset and returns ID", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Create test tenant
		tenantID := database.CreateTestTenant(t, db)

		// Setup: Asset to create
		asset := &Asset{
			TenantID:      tenantID,
			CanonicalName: "new-asset.example.com",
			FirstSeenAt:   time.Now(),
			LastSeenAt:    time.Now(),
			IsActive:      true,
		}

		// Test: Create asset
		err := repo.Create(ctx, asset)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if asset.ID == 0 {
			t.Error("expected ID to be set, got 0")
		}
		if asset.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set, got zero time")
		}
		if asset.UpdatedAt.IsZero() {
			t.Error("expected UpdatedAt to be set, got zero time")
		}
	})

	t.Run("enforces unique constraint on tenant + canonical name", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Create test tenant
		tenantID := database.CreateTestTenant(t, db)

		// Setup: Create first asset
		asset1 := &Asset{
			TenantID:      tenantID,
			CanonicalName: "duplicate.example.com",
			FirstSeenAt:   time.Now(),
			LastSeenAt:    time.Now(),
			IsActive:      true,
		}
		_ = repo.Create(ctx, asset1)

		// Test: Try to create duplicate
		asset2 := &Asset{
			TenantID:      tenantID,
			CanonicalName: "duplicate.example.com",
			FirstSeenAt:   time.Now(),
			LastSeenAt:    time.Now(),
			IsActive:      true,
		}
		err := repo.Create(ctx, asset2)

		// Assert: Should fail unique constraint
		if err == nil {
			t.Error("expected error for duplicate asset, got nil")
		}
	})
}

func TestAssetRepository_UpdateLastSeen(t *testing.T) {
	t.Run("updates last_seen_at timestamp", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Create test tenant and asset
		tenantID := database.CreateTestTenant(t, db)
		_, err := db.Exec(
			`INSERT INTO assets (tenant_id, canonical_name, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())`,
			tenantID, "update-test.example.com",
		)
		if err != nil {
			t.Fatalf("Failed to create test asset: %v", err)
		}

		var assetID int64
		err = db.QueryRow("SELECT id FROM assets WHERE canonical_name = $1", "update-test.example.com").Scan(&assetID)
		if err != nil {
			t.Fatalf("Failed to get asset ID: %v", err)
		}

		// Setup
		lastSeen := time.Now().Add(-1 * time.Hour)

		// Test: Update last seen
		err = repo.UpdateLastSeen(ctx, assetID, lastSeen)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify: Fetch asset and check timestamp
		asset, _ := repo.GetByID(ctx, assetID)
		if asset.LastSeenAt.Unix() != lastSeen.Unix() {
			t.Errorf("expected LastSeenAt to be updated to %v, got %v", lastSeen, asset.LastSeenAt)
		}
	})
}

func TestAssetRepository_AddIdentifier(t *testing.T) {
	t.Run("adds identifier to existing asset", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Create test tenant and asset
		tenantID := database.CreateTestTenant(t, db)
		_, err := db.Exec(
			`INSERT INTO assets (tenant_id, canonical_name, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())`,
			tenantID, "identifier-test.example.com",
		)
		if err != nil {
			t.Fatalf("Failed to create test asset: %v", err)
		}

		var assetID int64
		err = db.QueryRow("SELECT id FROM assets WHERE canonical_name = $1", "identifier-test.example.com").Scan(&assetID)
		if err != nil {
			t.Fatalf("Failed to get asset ID: %v", err)
		}

		// Setup
		identifier := &AssetIdentifier{
			TenantID:    tenantID,
			AssetID:     assetID,
			IDType:      "hostname_norm",
			IDValue:     "test-server.example.com",
			FirstSeenAt: time.Now(),
			LastSeenAt:  time.Now(),
			Source:      "tenable",
		}

		// Test
		err = repo.AddIdentifier(ctx, identifier)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if identifier.ID == 0 {
			t.Error("expected ID to be set, got 0")
		}
	})
}

func TestAssetRepository_GetIdentifiers(t *testing.T) {
	t.Run("returns all identifiers for asset", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Create test tenant and asset
		tenantID := database.CreateTestTenant(t, db)
		_, err := db.Exec(
			`INSERT INTO assets (tenant_id, canonical_name, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())`,
			tenantID, "identifiers-test.example.com",
		)
		if err != nil {
			t.Fatalf("Failed to create test asset: %v", err)
		}

		var assetID int64
		err = db.QueryRow("SELECT id FROM assets WHERE canonical_name = $1", "identifiers-test.example.com").Scan(&assetID)
		if err != nil {
			t.Fatalf("Failed to get asset ID: %v", err)
		}

		// Add some identifiers
		_, err = db.Exec(
			`INSERT INTO asset_identifiers (tenant_id, asset_id, id_type, id_value, source, first_seen_at, last_seen_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW()), ($1, $2, $6, $7, $8, NOW(), NOW())`,
			tenantID, assetID, "hostname_norm", "test-server.example.com", "tenable",
			"ip", "192.168.1.1", "tenable",
		)
		if err != nil {
			t.Fatalf("Failed to create test identifiers: %v", err)
		}

		// Test
		identifiers, err := repo.GetIdentifiers(ctx, assetID)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if identifiers == nil {
			t.Fatal("expected identifiers slice, got nil")
		}
		if len(identifiers) != 2 {
			t.Errorf("expected 2 identifiers, got %d", len(identifiers))
		}
	})

	t.Run("returns empty slice for asset with no identifiers", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Test with non-existent asset ID
		identifiers, err := repo.GetIdentifiers(ctx, 999999)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if identifiers == nil {
			t.Fatal("expected empty slice, got nil")
		}
		if len(identifiers) != 0 {
			t.Errorf("expected empty slice, got %d identifiers", len(identifiers))
		}
	})
}
