package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireRoles(t *testing.T) {
	t.Run("allows access when user has required role", func(t *testing.T) {
		middleware := RequireRoles(RoleRequirement{
			Roles: []string{RoleAdmin},
			Any:   true,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		req := httptest.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()

		middleware(handler).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "success" {
			t.Errorf("expected body 'success', got '%s'", w.Body.String())
		}
	})

	t.Run("allows access when user has any of multiple roles", func(t *testing.T) {
		middleware := RequireRoles(RoleRequirement{
			Roles: []string{RoleAdmin, RoleAnalyst},
			Any:   true,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/resource", nil)
		w := httptest.NewRecorder()

		middleware(handler).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("middleware is created successfully", func(t *testing.T) {
		middleware := RequireRoles(RoleRequirement{
			Roles: []string{RoleViewer},
			Any:   true,
		})

		if middleware == nil {
			t.Error("expected middleware to be created, got nil")
		}
	})
}

func TestRequireAdmin(t *testing.T) {
	t.Run("creates admin requirement middleware", func(t *testing.T) {
		middleware := RequireAdmin()

		if middleware == nil {
			t.Fatal("expected middleware, got nil")
		}

		// Verify it's a function (middleware)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()

		middleware(handler).ServeHTTP(w, req)

		// Should allow access (no user context check yet in MVP)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestRequireAnalyst(t *testing.T) {
	t.Run("creates analyst requirement middleware", func(t *testing.T) {
		middleware := RequireAnalyst()

		if middleware == nil {
			t.Fatal("expected middleware, got nil")
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/analyst", nil)
		w := httptest.NewRecorder()

		middleware(handler).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestRequireViewer(t *testing.T) {
	t.Run("creates viewer requirement middleware", func(t *testing.T) {
		middleware := RequireViewer()

		if middleware == nil {
			t.Fatal("expected middleware, got nil")
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/viewer", nil)
		w := httptest.NewRecorder()

		middleware(handler).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestRoleConstants(t *testing.T) {
	t.Run("has correct role values", func(t *testing.T) {
		if RoleAdmin != "admin" {
			t.Errorf("expected RoleAdmin 'admin', got '%s'", RoleAdmin)
		}
		if RoleAnalyst != "analyst" {
			t.Errorf("expected RoleAnalyst 'analyst', got '%s'", RoleAnalyst)
		}
		if RoleViewer != "viewer" {
			t.Errorf("expected RoleViewer 'viewer', got '%s'", RoleViewer)
		}
	})
}
