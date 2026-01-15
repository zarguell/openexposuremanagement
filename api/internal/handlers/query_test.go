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

// mockQueryExecutor is a mock implementation of ExecutorInterface for testing
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

// TestQueryAssets_Success tests successful asset query execution
func TestQueryAssets_Success(t *testing.T) {
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
			"filters": [{"field": "is_active", "operator": "eq", "value": true}]
		}`

		req := httptest.NewRequest("POST", "/api/v1/query/assets", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		// Add user context (simulating auth middleware)
		req = setQueryUserContextWithRequestID(req, 1, "test-req-123")

		w := httptest.NewRecorder()
		handler.QueryAssets(w, req)

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

// TestQueryFindings_Success tests successful findings query execution
func TestQueryFindings_Success(t *testing.T) {
	t.Run("valid query returns results", func(t *testing.T) {
		executor := &mockQueryExecutor{
			result: &query.QueryResult{
				Data: []map[string]interface{}{
					{
						"id":        int64(1),
						"severity":  "critical",
						"cve":       "CVE-2024-1234",
						"epss_score": 0.95,
					},
				},
				Meta: &query.QueryMeta{
					TotalRows:       142,
					ExecutionTimeMs: 45,
					HasMore:         true,
				},
			},
		}

		handler := NewQueryHandler(executor)

		body := `{
			"filters": [
				{"field": "severity", "operator": "in", "value": ["critical", "high"]}
			],
			"limit": 50
		}`

		req := httptest.NewRequest("POST", "/api/v1/query/findings", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-findings")

		w := httptest.NewRecorder()
		handler.QueryFindings(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		meta, ok := resp["meta"].(map[string]interface{})
		if !ok {
			t.Fatal("expected meta field in response")
		}

		if meta["total_rows"] != float64(142) {
			t.Errorf("got total_rows %v, want 142", meta["total_rows"])
		}

		if meta["has_more"] != true {
			t.Errorf("got has_more %v, want true", meta["has_more"])
		}
	})
}

func TestQueryAssets_Errors(t *testing.T) {
	t.Run("invalid JSON returns 400", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{invalid json`
		req := httptest.NewRequest("POST", "/api/v1/query/assets", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-456")

		w := httptest.NewRecorder()
		handler.QueryAssets(w, req)

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
			"filters": [{"field": "severity", "operator": "eq", "value": "critical"}]
		}`

		req := httptest.NewRequest("POST", "/api/v1/query/assets", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-789")

		w := httptest.NewRecorder()
		handler.QueryAssets(w, req)

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
			"filters": [{"field": "is_active", "operator": "eq", "value": true}]
		}`

		req := httptest.NewRequest("POST", "/api/v1/query/assets", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.QueryAssets(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})

	t.Run("missing filters returns 400", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{"limit": 50}`

		req := httptest.NewRequest("POST", "/api/v1/query/assets", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-no-filters")

		w := httptest.NewRecorder()
		handler.QueryAssets(w, req)

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

		if errResp["code"] != "MISSING_FILTERS" {
			t.Errorf("got error code %v, want MISSING_FILTERS", errResp["code"])
		}
	})

	t.Run("returns 500 for non-validation errors", func(t *testing.T) {
		executor := &mockQueryExecutor{
			err: errors.New("database connection failed"),
		}
		handler := NewQueryHandler(executor)

		body := `{
			"filters": [{"field": "is_active", "operator": "eq", "value": true}]
		}`

		req := httptest.NewRequest("POST", "/api/v1/query/assets", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-db-error")

		w := httptest.NewRecorder()
		handler.QueryAssets(w, req)

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

func TestQueryFindings_Errors(t *testing.T) {
	t.Run("invalid JSON returns 400", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{invalid json`
		req := httptest.NewRequest("POST", "/api/v1/query/findings", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-findings-400")

		w := httptest.NewRecorder()
		handler.QueryFindings(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", w.Code)
		}
	})

	t.Run("missing filters returns 400", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{"sort": [{"field": "severity", "order": "desc"}]}`

		req := httptest.NewRequest("POST", "/api/v1/query/findings", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-findings-no-filter")

		w := httptest.NewRecorder()
		handler.QueryFindings(w, req)

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

		if errResp["code"] != "MISSING_FILTERS" {
			t.Errorf("got error code %v, want MISSING_FILTERS", errResp["code"])
		}
	})
}

func TestSavedQueries(t *testing.T) {
	t.Run("ListSavedQueries returns not implemented", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		req := httptest.NewRequest("GET", "/api/v1/query/saved", nil)
		req = setQueryUserContextWithRequestID(req, 1, "test-req-saved-list")

		w := httptest.NewRecorder()
		handler.ListSavedQueries(w, req)

		if w.Code != http.StatusNotImplemented {
			t.Errorf("got status %d, want 501", w.Code)
		}
	})

	t.Run("CreateSavedQuery returns not implemented", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{"name": "test", "filters": []}`
		req := httptest.NewRequest("POST", "/api/v1/query/saved", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-saved-create")

		w := httptest.NewRecorder()
		handler.CreateSavedQuery(w, req)

		if w.Code != http.StatusNotImplemented {
			t.Errorf("got status %d, want 501", w.Code)
		}
	})

	t.Run("GetSavedQuery returns not implemented", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		req := httptest.NewRequest("GET", "/api/v1/query/saved/my-query", nil)
		req = setQueryUserContextWithRequestID(req, 1, "test-req-saved-get")

		w := httptest.NewRecorder()
		handler.GetSavedQuery(w, req, "my-query")

		if w.Code != http.StatusNotImplemented {
			t.Errorf("got status %d, want 501", w.Code)
		}
	})

	t.Run("DeleteSavedQuery returns not implemented", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		req := httptest.NewRequest("DELETE", "/api/v1/query/saved/my-query", nil)
		req = setQueryUserContextWithRequestID(req, 1, "test-req-saved-delete")

		w := httptest.NewRecorder()
		handler.DeleteSavedQuery(w, req, "my-query")

		if w.Code != http.StatusNotImplemented {
			t.Errorf("got status %d, want 501", w.Code)
		}
	})
}

// TestQueryUnified tests unified query endpoint with JOIN support
func TestQueryUnified(t *testing.T) {
	t.Run("valid join query returns results", func(t *testing.T) {
		executor := &mockQueryExecutor{
			result: &query.QueryResult{
				Data: []map[string]interface{}{
					{
						"id":             int64(1),
						"canonical_name": "server1.local",
						"vendor":         "Microsoft",
						"product_name":   "Windows Server",
					},
				},
				Meta: &query.QueryMeta{
					TotalRows:       1,
					ExecutionTimeMs: 25,
					HasMore:         false,
				},
			},
		}

		handler := NewQueryHandler(executor)

		body := `{
			"primary_entity": "assets",
			"join": {
				"entity": "software_inventory",
				"type": "left",
				"on": {"primary": "id", "joined": "asset_id"}
			},
			"filters": [
				{"field": "vendor", "operator": "eq", "value": "Microsoft"}
			],
			"limit": 100
		}`

		req := httptest.NewRequest("POST", "/api/v1/query/unified", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-unified")

		w := httptest.NewRecorder()
		handler.QueryUnified(w, req)

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
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{invalid json`
		req := httptest.NewRequest("POST", "/api/v1/query/unified", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-unified-400")

		w := httptest.NewRecorder()
		handler.QueryUnified(w, req)

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

	t.Run("missing user context returns 401", func(t *testing.T) {
		executor := &mockQueryExecutor{}
		handler := NewQueryHandler(executor)

		body := `{
			"primary_entity": "assets",
			"filters": [{"field": "is_active", "operator": "eq", "value": true}]
		}`

		req := httptest.NewRequest("POST", "/api/v1/query/unified", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.QueryUnified(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})

	t.Run("validation error returns 422", func(t *testing.T) {
		executor := &mockQueryExecutor{
			err: errors.New("validation error: unsupported join entity 'unknown_entity'"),
		}
		handler := NewQueryHandler(executor)

		body := `{
			"primary_entity": "assets",
			"join": {
				"entity": "unknown_entity",
				"type": "left"
			}
		}`

		req := httptest.NewRequest("POST", "/api/v1/query/unified", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-unified-422")

		w := httptest.NewRecorder()
		handler.QueryUnified(w, req)

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

	t.Run("query execution error returns 500", func(t *testing.T) {
		executor := &mockQueryExecutor{
			err: errors.New("database connection failed"),
		}
		handler := NewQueryHandler(executor)

		body := `{
			"primary_entity": "assets",
			"filters": [{"field": "is_active", "operator": "eq", "value": true}]
		}`

		req := httptest.NewRequest("POST", "/api/v1/query/unified", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-unified-500")

		w := httptest.NewRecorder()
		handler.QueryUnified(w, req)

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
func setQueryUserContextWithRequestID(req *http.Request, tenantID int64, requestID string) *http.Request {
	userCtx := &auth.UserContext{
		UserID:   "1",
		TenantID: tenantID,
		Email:    "test@example.com",
		Roles:    []string{"analyst"},
	}
	ctx := context.WithValue(req.Context(), auth.UserContextKey, userCtx)
	ctx = context.WithValue(ctx, "request_id", requestID)
	return req.WithContext(ctx)
}
