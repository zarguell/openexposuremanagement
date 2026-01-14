package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/openexposuremanagement/oem/internal/services/query"
)

// mockQueryExecutor is a mock implementation of QueryExecutor for testing
type mockQueryExecutor struct {
	result *query.QueryResult
	err    error
}

func (m *mockQueryExecutor) Execute(ctx context.Context, tenantID string, entityType string, q *query.Query) (*query.QueryResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// TestPostQueriesExecute_Success tests successful query execution
func TestPostQueriesExecute_Success(t *testing.T) {
	t.Run("valid query returns results", func(t *testing.T) {
		executor := &mockQueryExecutor{
			result: &query.QueryResult{
				Data: []map[string]interface{}{
					{"id": int64(1), "canonical_name": "test-asset.local"},
				},
				Meta: &query.QueryMeta{
					TotalRows:       1,
					ExecutionTimeMs: 5,
					HasMore:         false,
				},
			},
		}

		handler := NewQueryHandler(executor)

		body := `{
			"entity_type": "assets",
			"query": {
				"filters": [{"field": "is_active", "operator": "eq", "value": true}]
			}
		}`

		req := httptest.NewRequest("POST", "/api/v1/queries/execute", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "test-req-123")

		// Add user context (simulating auth middleware)
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["data"] == nil {
			t.Fatal("expected data field in response")
		}

		data, ok := resp["data"].([]interface{})
		if !ok || len(data) != 1 {
			t.Fatalf("expected 1 data row, got %d", len(data))
		}

		meta, ok := resp["meta"].(map[string]interface{})
		if !ok {
			t.Fatal("expected meta field in response")
		}

		if meta["total_rows"] != float64(1) {
			t.Errorf("got total_rows %v, want 1", meta["total_rows"])
		}
	})
}

func TestPostQueriesExecute_Errors(t *testing.T) {
	t.Run("invalid JSON returns 400", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{invalid json`
		req := httptest.NewRequest("POST", "/api/v1/queries/execute", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "test-req-456")
		req = setQueryUserContext(req, 1) // Add user context first

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		errResp, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("expected error field in response")
		}

		if errResp["code"] != "INVALID_JSON" {
			t.Errorf("got error code %v, want INVALID_JSON", errResp["code"])
		}
	})

	t.Run("validation error returns 422", func(t *testing.T) {
		executor := &mockQueryExecutor{
			err: errors.New("validation error: field 'severity' not allowed for entity assets"),
		}
		handler := NewQueryHandler(executor)

		body := `{
			"entity_type": "assets",
			"query": {
				"filters": [{"field": "severity", "operator": "eq", "value": "critical"}]
			}
		}`

		req := httptest.NewRequest("POST", "/api/v1/queries/execute", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "test-req-789")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("got status %d, want 422", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		errResp, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("expected error field in response")
		}

		if errResp["code"] != "VALIDATION_ERROR" {
			t.Errorf("got error code %v, want VALIDATION_ERROR", errResp["code"])
		}
	})

	t.Run("missing user context returns 401", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{
			"entity_type": "assets",
			"query": {
				"filters": [{"field": "is_active", "operator": "eq", "value": true}]
			}
		}`

		req := httptest.NewRequest("POST", "/api/v1/queries/execute", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "test-req-no-user")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})

	t.Run("invalid entity_type returns 400", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{
			"entity_type": "invalid_type",
			"query": {
				"filters": [{"field": "is_active", "operator": "eq", "value": true}]
			}
		}`

		req := httptest.NewRequest("POST", "/api/v1/queries/execute", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "test-req-bad-type")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		errResp, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("expected error field in response")
		}

		if errResp["code"] != "INVALID_ENTITY_TYPE" {
			t.Errorf("got error code %v, want INVALID_ENTITY_TYPE", errResp["code"])
		}
	})

	t.Run("defaults entity_type to findings", func(t *testing.T) {
		executor := &mockQueryExecutor{
			result: &query.QueryResult{
				Data: []map[string]interface{}{},
				Meta: &query.QueryMeta{
					TotalRows:       0,
					ExecutionTimeMs: 2,
					HasMore:         false,
				},
			},
		}
		handler := NewQueryHandler(executor)

		body := `{
			"query": {
				"filters": [{"field": "severity", "operator": "eq", "value": "critical"}]
			}
		}`

		req := httptest.NewRequest("POST", "/api/v1/queries/execute", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "test-req-default")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
		}
	})
}

func TestGetQueriesSaved(t *testing.T) {
	t.Run("returns not implemented yet", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		req := httptest.NewRequest("GET", "/api/v1/queries/saved", nil)
		req.Header.Set("X-Request-ID", "test-req-saved-list")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotImplemented {
			t.Errorf("got status %d, want 501", w.Code)
		}
	})
}

func TestPostQueriesSaved(t *testing.T) {
	t.Run("returns not implemented yet", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{"name": "test", "query": {"filters": []}}`
		req := httptest.NewRequest("POST", "/api/v1/queries/saved", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "test-req-saved-create")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotImplemented {
			t.Errorf("got status %d, want 501", w.Code)
		}
	})
}

func TestGetQueriesSavedByID(t *testing.T) {
	t.Run("returns not implemented yet", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		req := httptest.NewRequest("GET", "/api/v1/queries/saved/123", nil)
		req.Header.Set("X-Request-ID", "test-req-saved-get")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotImplemented {
			t.Errorf("got status %d, want 501", w.Code)
		}
	})
}

func TestDeleteQueriesSaved(t *testing.T) {
	t.Run("returns not implemented yet", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		req := httptest.NewRequest("DELETE", "/api/v1/queries/saved/123", nil)
		req.Header.Set("X-Request-ID", "test-req-saved-delete")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotImplemented {
			t.Errorf("got status %d, want 501", w.Code)
		}
	})
}

func TestQueryHandler_MethodNotAllowed(t *testing.T) {
	t.Run("rejects unsupported methods", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		req := httptest.NewRequest("PUT", "/api/v1/queries/saved", nil)
		req.Header.Set("X-Request-ID", "test-req-method")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("got status %d, want 405", w.Code)
		}
	})

	t.Run("rejects unsupported method on saved query detail", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		req := httptest.NewRequest("POST", "/api/v1/queries/saved/123", nil)
		req.Header.Set("X-Request-ID", "test-req-method-detail")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("got status %d, want 405", w.Code)
		}
	})
}

func TestQueryHandler_Routing(t *testing.T) {
	t.Run("returns 404 for unknown path", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		req := httptest.NewRequest("GET", "/api/v1/queries/unknown", nil)
		req.Header.Set("X-Request-ID", "test-req-404")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want 404", w.Code)
		}
	})

	t.Run("returns 400 for invalid saved query path", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		req := httptest.NewRequest("GET", "/api/v1/queries/saved/", nil)
		req.Header.Set("X-Request-ID", "test-req-bad-path")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		errResp, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("expected error field in response")
		}

		if errResp["code"] != "INVALID_PATH" {
			t.Errorf("got error code %v, want INVALID_PATH", errResp["code"])
		}
	})
}

func TestPostQueriesExecute_AdditionalErrors(t *testing.T) {
	t.Run("returns 400 when query is missing", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{"entity_type": "assets"}`
		req := httptest.NewRequest("POST", "/api/v1/queries/execute", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "test-req-no-query")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		errResp, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("expected error field in response")
		}

		if errResp["code"] != "MISSING_QUERY" {
			t.Errorf("got error code %v, want MISSING_QUERY", errResp["code"])
		}
	})

	t.Run("returns 500 for non-validation errors", func(t *testing.T) {
		executor := &mockQueryExecutor{
			err: errors.New("database connection failed"),
		}
		handler := NewQueryHandler(executor)

		body := `{
			"entity_type": "assets",
			"query": {
				"filters": [{"field": "is_active", "operator": "eq", "value": true}]
			}
		}`

		req := httptest.NewRequest("POST", "/api/v1/queries/execute", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "test-req-db-error")
		req = setQueryUserContext(req, 1)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("got status %d, want 500", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		errResp, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("expected error field in response")
		}

		if errResp["code"] != "QUERY_FAILED" {
			t.Errorf("got error code %v, want QUERY_FAILED", errResp["code"])
		}
	})
}

// Helper function to set user context for query tests
func setQueryUserContext(req *http.Request, tenantID int64) *http.Request {
	userCtx := &auth.UserContext{
		UserID:   "1",
		TenantID: tenantID,
		Email:    "test@example.com",
		Roles:    []string{"analyst"},
	}
	ctx := context.WithValue(req.Context(), auth.UserContextKey, userCtx)
	return req.WithContext(ctx)
}
