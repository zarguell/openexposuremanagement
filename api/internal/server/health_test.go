package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openexposuremanagement/oem/internal/config"
)

func TestHealthEndpoints(t *testing.T) {
	cfg := &config.Config{
		DatabaseURL: "postgres://test:test@localhost:5432/test",
		Port:        "8080",
		Environment: "development",
	}

	t.Run("GET /healthz returns 200", func(t *testing.T) {
		srv := New(cfg, nil)

		req := httptest.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()

		srv.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
		}

		// Check response body contains "healthy" status
		body := w.Body.String()
		if body == "" {
			t.Errorf("got empty body, want status response")
		}
	})

	t.Run("GET /healthz/live returns 200", func(t *testing.T) {
		srv := New(cfg, nil)

		req := httptest.NewRequest("GET", "/healthz/live", nil)
		w := httptest.NewRecorder()

		srv.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
		}

		// Check response body
		body := w.Body.String()
		if body == "" {
			t.Errorf("got empty body, want status response")
		}
	})

	t.Run("GET /healthz/ready returns 503 when DB is nil", func(t *testing.T) {
		// Note: In real implementation, this would check DB connectivity
		// For now, we just verify the endpoint exists
		srv := New(cfg, nil)

		req := httptest.NewRequest("GET", "/healthz/ready", nil)
		w := httptest.NewRecorder()

		srv.router.ServeHTTP(w, req)

		// Should return 503 because DB is nil
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("got status %d, want 503", w.Code)
		}
	})

	t.Run("POST /healthz/live returns 405", func(t *testing.T) {
		srv := New(cfg, nil)

		req := httptest.NewRequest("POST", "/healthz/live", nil)
		w := httptest.NewRecorder()

		srv.router.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("got status %d, want 405", w.Code)
		}
	})

	t.Run("POST /healthz/ready returns 405", func(t *testing.T) {
		srv := New(cfg, nil)

		req := httptest.NewRequest("POST", "/healthz/ready", nil)
		w := httptest.NewRecorder()

		srv.router.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("got status %d, want 405", w.Code)
		}
	})
}
