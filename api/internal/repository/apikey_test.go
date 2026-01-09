package repository

import (
	"context"
	"testing"
)

func TestAPIKeyRepository_GetByKeyHash(t *testing.T) {
	t.Run("returns API key for valid hash", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAPIKeyRepository(db)
		ctx := context.Background()

		// Test
		result, err := repo.GetByKeyHash(ctx, "valid-hash")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("expected API key, got nil")
		}
		if result.KeyHash != "valid-hash" {
			t.Errorf("expected hash 'valid-hash', got '%s'", result.KeyHash)
		}
		if result.RevokedAt != nil {
			t.Error("expected nil RevokedAt for active key")
		}
	})

	t.Run("returns nil for revoked key", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAPIKeyRepository(db)
		ctx := context.Background()

		// Test
		result, err := repo.GetByKeyHash(ctx, "revoked-hash")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != nil {
			t.Error("expected nil for revoked key, got result")
		}
	})

	t.Run("returns nil for non-existent hash", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAPIKeyRepository(db)
		ctx := context.Background()

		// Test
		result, err := repo.GetByKeyHash(ctx, "nonexistent")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != nil {
			t.Error("expected nil for non-existent key, got result")
		}
	})
}

func TestAPIKeyRepository_GetByID(t *testing.T) {
	t.Run("returns API key when ID exists", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAPIKeyRepository(db)
		ctx := context.Background()

		// Test
		result, err := repo.GetByID(ctx, 1)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("expected API key, got nil")
		}
		if result.ID != 1 {
			t.Errorf("expected ID 1, got %d", result.ID)
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAPIKeyRepository(db)
		ctx := context.Background()

		// Test
		result, err := repo.GetByID(ctx, 999)

		// Assert
		if err == nil {
			t.Error("expected error, got nil")
		}
		if result != nil {
			t.Error("expected nil API key, got result")
		}
	})
}

func TestAPIKeyRepository_Create(t *testing.T) {
	t.Run("creates API key with generated ID and timestamp", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAPIKeyRepository(db)
		ctx := context.Background()

		// Setup
		key := &APIKey{
			TenantID:   1,
			Name:       "Test Key",
			KeyHash:    "test-hash",
			ScopesJSON: `["ingest:vm"]`,
		}

		// Test
		err := repo.Create(ctx, key)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if key.ID == 0 {
			t.Error("expected ID to be set, got 0")
		}
		if key.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
	})
}

func TestAPIKeyRepository_Revoke(t *testing.T) {
	t.Run("revokes API key by setting revoked_at", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewAPIKeyRepository(db)
		ctx := context.Background()

		// Test
		err := repo.Revoke(ctx, 1)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify: Fetch key and check revoked_at
		key, _ := repo.GetByID(ctx, 1)
		if key.RevokedAt == nil {
			t.Error("expected RevokedAt to be set")
		}
	})
}
