package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// This file uses TDD approach for Asset repository
// Tests are written FIRST, then implementation

// setupTestDB creates a test database connection
// TODO: This is a placeholder - we'll implement actual test DB setup
func setupTestDB(t *testing.T) *sqlx.DB {
	// For now, return nil to make tests compile
	// In real implementation, this would create a test database
	t.Skip("Test database not yet set up - backfill in progress")
	return nil
}

func TestAssetRepository_GetByID(t *testing.T) {
	t.Run("returns asset when it exists", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Test: Get by ID
		result, err := repo.GetByID(ctx, 123)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("expected asset, got nil")
		}
		if result.ID != 123 {
			t.Errorf("expected ID 123, got %d", result.ID)
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

		// Test
		result, err := repo.GetByCanonicalName(ctx, 1, "test-asset.example.com")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("expected asset, got nil")
		}
		if result.TenantID != 1 {
			t.Errorf("expected tenant ID 1, got %d", result.TenantID)
		}
		if result.CanonicalName != "test-asset.example.com" {
			t.Errorf("expected canonical name 'test-asset.example.com', got '%s'", result.CanonicalName)
		}
	})

	t.Run("returns error for wrong tenant", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Test: Try to get asset from different tenant
		result, err := repo.GetByCanonicalName(ctx, 2, "test-asset.example.com")

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

		// Setup: Asset to create
		asset := &Asset{
			TenantID:      1,
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

		// Setup: Create first asset
		asset1 := &Asset{
			TenantID:      1,
			CanonicalName: "duplicate.example.com",
			FirstSeenAt:   time.Now(),
			LastSeenAt:    time.Now(),
			IsActive:      true,
		}
		_ = repo.Create(ctx, asset1)

		// Test: Try to create duplicate
		asset2 := &Asset{
			TenantID:      1,
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

		// Setup
		lastSeen := time.Now().Add(-1 * time.Hour)

		// Test: Update last seen
		err := repo.UpdateLastSeen(ctx, 123, lastSeen)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify: Fetch asset and check timestamp
		asset, _ := repo.GetByID(ctx, 123)
		if asset.LastSeenAt.Unix() != lastSeen.Unix() {
			t.Errorf("expected LastSeenAt to be updated")
		}
	})
}

func TestAssetRepository_AddIdentifier(t *testing.T) {
	t.Run("adds identifier to existing asset", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Setup
		identifier := &AssetIdentifier{
			TenantID:    1,
			AssetID:     123,
			IDType:      "hostname_norm",
			IDValue:     "test-server.example.com",
			FirstSeenAt: time.Now(),
			LastSeenAt:  time.Now(),
			Source:      "tenable",
		}

		// Test
		err := repo.AddIdentifier(ctx, identifier)

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

		// Test
		identifiers, err := repo.GetIdentifiers(ctx, 123)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if identifiers == nil {
			t.Fatal("expected identifiers slice, got nil")
		}
		if len(identifiers) == 0 {
			t.Error("expected at least one identifier, got empty slice")
		}
	})

	t.Run("returns empty slice for asset with no identifiers", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAssetRepository(db)
		ctx := context.Background()

		// Test
		identifiers, err := repo.GetIdentifiers(ctx, 999)

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
