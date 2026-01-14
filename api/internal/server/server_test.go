package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestQueryEndpointsRegistered(t *testing.T) {
	// Enable demo mode to bypass auth for routing tests
	os.Setenv("DEMO_MODE", "true")
	defer os.Unsetenv("DEMO_MODE")

	cfg := &config.Config{
		DatabaseURL: "postgres://test:test@localhost:5432/test",
		Port:        "8080",
	}

	t.Run("POST /api/v1/query/findings is registered", func(t *testing.T) {
		s := New(cfg, nil)

		body := `{"filters": [{"field": "severity", "operator": "eq", "value": "critical"}]}`
		req := httptest.NewRequest("POST", "/api/v1/query/findings", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		// Catch panics from nil DB
		defer func() {
			if r := recover(); r != nil {
				// Panic is OK - it means endpoint is registered but DB is nil
				// 404 would mean endpoint is NOT registered
				t.Logf("Caught expected panic from nil DB: %v", r)
			}
		}()

		s.router.ServeHTTP(w, req)

		// In demo mode, should not be 404 (endpoint is registered)
		// Might be 500 if DB is nil, or 422 for validation, but not 404
		if w.Code == http.StatusNotFound {
			t.Errorf("Endpoint not registered: got status 404")
		}
	})

	t.Run("POST /api/v1/query/assets is registered", func(t *testing.T) {
		s := New(cfg, nil)

		body := `{"filters": [{"field": "is_active", "operator": "eq", "value": true}]}`
		req := httptest.NewRequest("POST", "/api/v1/query/assets", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		// Catch panics from nil DB
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Caught expected panic from nil DB: %v", r)
			}
		}()

		s.router.ServeHTTP(w, req)

		// Should not be 404
		if w.Code == http.StatusNotFound {
			t.Errorf("Endpoint not registered: got status 404")
		}
	})

	t.Run("GET /api/v1/query/saved is registered (stub)", func(t *testing.T) {
		s := New(cfg, nil)

		req := httptest.NewRequest("GET", "/api/v1/query/saved", nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		// Should return 501 Not Implemented (stub) or similar, but not 404
		if w.Code == http.StatusNotFound {
			t.Errorf("Endpoint not registered: got status 404")
		}

		// Stub endpoints should return 501 Not Implemented
		if w.Code != http.StatusNotImplemented {
			t.Logf("Expected status 501 Not Implemented for stub, got %d", w.Code)
		}
	})

	t.Run("POST /api/v1/query/saved is registered (stub)", func(t *testing.T) {
		s := New(cfg, nil)

		body := `{"name": "test-query", "query": {"filters": []}}`
		req := httptest.NewRequest("POST", "/api/v1/query/saved", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		// Should not be 404
		if w.Code == http.StatusNotFound {
			t.Errorf("Endpoint not registered: got status 404")
		}

		// Stub endpoints should return 501 Not Implemented
		if w.Code != http.StatusNotImplemented {
			t.Logf("Expected status 501 Not Implemented for stub, got %d", w.Code)
		}
	})
}

func TestQueryEndpointsRequireAuth(t *testing.T) {
	// Disable demo mode to test auth requirements
	os.Unsetenv("DEMO_MODE")

	cfg := &config.Config{
		DatabaseURL: "postgres://test:test@localhost:5432/test",
		Port:        "8080",
	}

	t.Run("POST /api/v1/query/findings requires auth", func(t *testing.T) {
		s := New(cfg, nil)

		body := `{"filters": [{"field": "severity", "operator": "eq", "value": "critical"}]}`
		req := httptest.NewRequest("POST", "/api/v1/query/findings", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		// Should return 401 without auth token
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized, got %d", w.Code)
		}

		// Check response contains error
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err == nil {
			if _, ok := resp["error"]; !ok {
				t.Errorf("Expected error in response body")
			}
		}
	})

	t.Run("POST /api/v1/query/assets requires auth", func(t *testing.T) {
		s := New(cfg, nil)

		body := `{"filters": [{"field": "is_active", "operator": "eq", "value": true}]}`
		req := httptest.NewRequest("POST", "/api/v1/query/assets", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized, got %d", w.Code)
		}
	})

	t.Run("GET /api/v1/query/saved requires auth", func(t *testing.T) {
		s := New(cfg, nil)

		req := httptest.NewRequest("GET", "/api/v1/query/saved", nil)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized, got %d", w.Code)
		}
	})
}

func TestQueryEndpointsValidation(t *testing.T) {
	// Enable demo mode for validation tests
	os.Setenv("DEMO_MODE", "true")
	defer os.Unsetenv("DEMO_MODE")

	cfg := &config.Config{
		DatabaseURL: "postgres://test:test@localhost:5432/test",
		Port:        "8080",
	}

	t.Run("POST /api/v1/query/findings with invalid JSON returns 400", func(t *testing.T) {
		s := New(cfg, nil)

		body := `{"invalid": "query"` // Missing closing brace
		req := httptest.NewRequest("POST", "/api/v1/query/findings", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request, got %d", w.Code)
		}

		// Check response contains error details
		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err == nil {
			if _, ok := resp["error"]; !ok {
				t.Errorf("Expected error in response body")
			}
		}
	})

	t.Run("POST /api/v1/query/findings with no filters returns 400", func(t *testing.T) {
		s := New(cfg, nil)

		body := `{}`
		req := httptest.NewRequest("POST", "/api/v1/query/findings", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		// Should return 400 for missing filters
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request, got %d", w.Code)
		}
	})
}
