package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openexposuremanagement/oem/internal/auth"
)

// TestGetDashboard tests the GET /dashboard endpoint
func TestGetDashboard_Authentication(t *testing.T) {
	t.Run("returns 401 without user context", func(t *testing.T) {
		db := getTestDB(t)
		handler := GetDashboard(db)

		req := httptest.NewRequest("GET", "/api/v1/dashboard", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})
}

func TestGetDashboard_Response(t *testing.T) {
	t.Run("returns asset counts", func(t *testing.T) {
		t.Skip("Test database not yet set up - backfill in progress")
	})

	t.Run("returns finding counts by severity", func(t *testing.T) {
		t.Skip("Test database not yet set up - backfill in progress")
	})

	t.Run("returns intel sync status", func(t *testing.T) {
		t.Skip("Test database not yet set up - backfill in progress")
	})

	t.Run("includes 'Intel last updated at' timestamp", func(t *testing.T) {
		t.Skip("Test database not yet set up - backfill in progress")
	})
}

func TestGetDashboard_TenantScoping(t *testing.T) {
	t.Run("only returns data for user's tenant", func(t *testing.T) {
		t.Skip("Test database not yet set up - backfill in progress")
	})
}

// Helper function for setting user context in tests
func setDashboardUserContext(req *http.Request, tenantID int64) *http.Request {
	userCtx := &auth.UserContext{
		UserID:   "1",
		TenantID: tenantID,
		Email:    "test@example.com",
		Roles:    []string{"analyst"},
	}
	ctx := context.WithValue(req.Context(), auth.UserContextKey, userCtx)
	return req.WithContext(ctx)
}
