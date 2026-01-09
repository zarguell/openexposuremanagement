package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openexposuremanagement/oem/internal/config"
)

func TestHandleHealthz(t *testing.T) {
	cfg := &config.Config{
		DatabaseURL: "postgres://test:test@localhost:5432/test",
		Port:        "8080",
	}

	s := New(cfg, nil)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expectedBody := `{"status":"healthy"}`
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, w.Body.String())
	}
}
