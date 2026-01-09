package repository

import (
	"context"
	"testing"
)

func TestUserRepository_GetByEmail(t *testing.T) {
	t.Run("returns user for tenant and email", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		// Test
		result, err := repo.GetByEmail(ctx, 1, "user@example.com")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("expected user, got nil")
		}
		if result.Email != "user@example.com" {
			t.Errorf("expected email 'user@example.com', got '%s'", result.Email)
		}
		if result.TenantID != 1 {
			t.Errorf("expected tenant ID 1, got %d", result.TenantID)
		}
	})

	t.Run("returns error for wrong tenant", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		// Test: Try different tenant
		result, err := repo.GetByEmail(ctx, 999, "user@example.com")

		// Assert
		if err == nil {
			t.Error("expected error for different tenant, got nil")
		}
		if result != nil {
			t.Error("expected nil user, got result")
		}
	})

	t.Run("returns error for non-existent email", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		// Test
		result, err := repo.GetByEmail(ctx, 1, "nonexistent@example.com")

		// Assert
		if err == nil {
			t.Error("expected error, got nil")
		}
		if result != nil {
			t.Error("expected nil user, got result")
		}
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	t.Run("returns user when ID exists", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		// Test
		result, err := repo.GetByID(ctx, 1)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("expected user, got nil")
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		// Test
		result, err := repo.GetByID(ctx, 999)

		// Assert
		if err == nil {
			t.Error("expected error, got nil")
		}
		if result != nil {
			t.Error("expected nil user, got result")
		}
	})
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("creates user with generated ID and timestamps", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		// Setup
		user := &User{
			TenantID:    1,
			Email:       "newuser@example.com",
			DisplayName: "New User",
			Status:      "active",
		}

		// Test
		err := repo.Create(ctx, user)

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.ID == 0 {
			t.Error("expected ID to be set, got 0")
		}
		if user.CreatedAt == "" {
			t.Error("expected CreatedAt to be set")
		}
		if user.UpdatedAt == "" {
			t.Error("expected UpdatedAt to be set")
		}
	})

	t.Run("enforces unique constraint on tenant + email", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		// Setup: Create first user
		user1 := &User{
			TenantID:    1,
			Email:       "duplicate@example.com",
			DisplayName: "User One",
			Status:      "active",
		}
		_ = repo.Create(ctx, user1)

		// Test: Try to create duplicate
		user2 := &User{
			TenantID:    1,
			Email:       "duplicate@example.com",
			DisplayName: "User Two",
			Status:      "active",
		}
		err := repo.Create(ctx, user2)

		// Assert: Should fail
		if err == nil {
			t.Error("expected error for duplicate email, got nil")
		}
	})

	t.Run("allows same email in different tenants", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewUserRepository(db)
		ctx := context.Background()

		// Setup: Create user in tenant 1
		user1 := &User{
			TenantID:    1,
			Email:       "shared@example.com",
			DisplayName: "User One",
			Status:      "active",
		}
		_ = repo.Create(ctx, user1)

		// Test: Create user with same email in tenant 2
		user2 := &User{
			TenantID:    2,
			Email:       "shared@example.com",
			DisplayName: "User Two",
			Status:      "active",
		}
		err := repo.Create(ctx, user2)

		// Assert: Should succeed
		if err != nil {
			t.Errorf("expected no error for same email in different tenant, got %v", err)
		}
	})
}
