package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/auth"
)

func TestRefreshIntel(t *testing.T) {
	t.Run("requires admin role", func(t *testing.T) {
		db := setupTestDB(t)
		defer teardownTestDB(t, db)

		handler := RefreshIntel(db)

		// Test without admin role - missing user context
		req := httptest.NewRequest("POST", "/intel/refresh", nil)
		ctx := context.Background()
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		// Missing user context - should return unauthorized
		handler(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("rejects non-admin users", func(t *testing.T) {
		db := setupTestDB(t)
		defer teardownTestDB(t, db)

		handler := RefreshIntel(db)

		// Create user context without admin role
		userCtx := &auth.UserContext{
			UserID:   "user123",
			TenantID: 1,
			Roles:    []string{"viewer"},
		}
		ctx := context.WithValue(context.Background(), auth.UserContextKey, userCtx)

		req := httptest.NewRequest("POST", "/intel/refresh", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403 for non-admin, got %d", w.Code)
		}
	})

	t.Run("allows admin users", func(t *testing.T) {
		db := setupTestDB(t)
		defer teardownTestDB(t, db)

		handler := RefreshIntel(db)

		// Create admin user context
		userCtx := &auth.UserContext{
			UserID:   "admin123",
			TenantID: 1,
			Roles:    []string{"admin"},
		}
		ctx := context.WithValue(context.Background(), auth.UserContextKey, userCtx)

		req := httptest.NewRequest("POST", "/intel/refresh", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("expected status 202, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response["status"] != "accepted" {
			t.Errorf("expected status 'accepted', got %v", response["status"])
		}

		if response["message"] != "Threat intelligence sync started" {
			t.Errorf("unexpected message: %v", response["message"])
		}
	})

	t.Run("rejects non-POST methods", func(t *testing.T) {
		db := setupTestDB(t)
		defer teardownTestDB(t, db)

		handler := RefreshIntel(db)

		userCtx := &auth.UserContext{
			UserID:   "admin123",
			TenantID: 1,
			Roles:    []string{"admin"},
		}
		ctx := context.WithValue(context.Background(), auth.UserContextKey, userCtx)

		req := httptest.NewRequest("GET", "/intel/refresh", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405 for GET, got %d", w.Code)
		}
	})
}

func TestGetIntelStatus(t *testing.T) {
	t.Run("returns sync status", func(t *testing.T) {
		db := setupTestDB(t)
		defer teardownTestDB(t, db)

		handler := GetIntelStatus(db)

		req := httptest.NewRequest("GET", "/intel/status", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if _, ok := response["sources"]; !ok {
			t.Error("expected 'sources' field in response")
		}
	})

	t.Run("rejects non-GET methods", func(t *testing.T) {
		db := setupTestDB(t)
		defer teardownTestDB(t, db)

		handler := GetIntelStatus(db)

		req := httptest.NewRequest("POST", "/intel/status", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405 for POST, got %d", w.Code)
		}
	})
}

// Helper functions for tests

func setupTestDB(t *testing.T) *sqlx.DB {
	// For unit tests, we'd use a mock or test database
	// For now, skip to indicate integration test
	t.Skip("integration test - requires database")
	return nil
}

func teardownTestDB(t *testing.T, db *sqlx.DB) {
	// Cleanup
}
