package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openexposuremanagement/oem/internal/services/query"
)

// TestOQLQueryEndpoint_Integration tests the OQL query endpoint with real HTTP calls
func TestOQLQueryEndpoint_Integration(t *testing.T) {
	// Create mock executor with realistic data
	executor := &mockQueryExecutor{
		result: &query.QueryResult{
			Data: []map[string]interface{}{
				{
					"id":            int64(1),
					"canonical_name": "webserver01.example.com",
					"is_active":      true,
					"last_seen_at":   "2024-01-15T10:30:00Z",
				},
				{
					"id":            int64(2),
					"canonical_name": "dbserver01.example.com",
					"is_active":      true,
					"last_seen_at":   "2024-01-15T10:31:00Z",
				},
			},
			Meta: &query.QueryMeta{
				TotalRows:       2,
				ExecutionTimeMs: 5,
				HasMore:         false,
			},
		},
	}

	handler := NewQueryHandler(executor, &mockQueryTranslator{})

	t.Run("valid OQL query returns results", func(t *testing.T) {
		body := `{"query": "is_active = true limit 10"}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-integration-001")

		w := httptest.NewRecorder()
		handler.QueryOQL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
			t.Logf("Response body: %s", w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["data"] == nil {
			t.Fatal("expected data field in response")
		}

		data, ok := resp["data"].([]interface{})
		if !ok {
			t.Fatal("data field is not an array")
		}

		if len(data) != 2 {
			t.Errorf("got %d results, want 2", len(data))
		}

		if resp["meta"] == nil {
			t.Fatal("expected meta field in response")
		}
	})

	t.Run("OQL query with dot-walking syntax", func(t *testing.T) {
		body := `{"query": "software.vendor = \"Microsoft\" limit 10"}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-integration-002")

		w := httptest.NewRecorder()
		handler.QueryOQL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
			t.Logf("Response body: %s", w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["data"] == nil {
			t.Fatal("expected data field in response")
		}
	})

	t.Run("OQL query with NOT operator (anti-join)", func(t *testing.T) {
		body := `{"query": "NOT software.vendor = \"CrowdStrike\" limit 10"}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-integration-003")

		w := httptest.NewRecorder()
		handler.QueryOQL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
			t.Logf("Response body: %s", w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["data"] == nil {
			t.Fatal("expected data field in response")
		}
	})

	t.Run("OQL query with sort clause", func(t *testing.T) {
		body := `{"query": "is_active = true sort canonical_name desc limit 10"}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-integration-004")

		w := httptest.NewRecorder()
		handler.QueryOQL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
			t.Logf("Response body: %s", w.Body.String())
		}
	})

	t.Run("OQL query with complex logic", func(t *testing.T) {
		body := `{"query": "is_active = true AND (findings.severity = \"critical\" OR findings.severity = \"high\") limit 10"}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-integration-005")

		w := httptest.NewRecorder()
		handler.QueryOQL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
			t.Logf("Response body: %s", w.Body.String())
		}
	})

	t.Run("invalid OQL syntax returns 400", func(t *testing.T) {
		body := `{"query": "is_active == true"}` // Invalid operator

		req := httptest.NewRequest("POST", "/api/v1/query/oql", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-integration-006")

		w := httptest.NewRecorder()
		handler.QueryOQL(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["error"] == nil {
			t.Fatal("expected error field in response")
		}

		errResp, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("error field is not an object")
		}

		if errResp["code"] != "OQL_PARSE_ERROR" {
			t.Errorf("got error code %v, want OQL_PARSE_ERROR", errResp["code"])
		}
	})

	t.Run("missing query field returns 400", func(t *testing.T) {
		body := `{}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-integration-007")

		w := httptest.NewRecorder()
		handler.QueryOQL(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["error"] == nil {
			t.Fatal("expected error field in response")
		}

		errResp, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("error field is not an object")
		}

		if errResp["code"] != "INVALID_REQUEST" {
			t.Errorf("got error code %v, want INVALID_REQUEST", errResp["code"])
		}
	})

	t.Run("missing user context returns 401", func(t *testing.T) {
		body := `{"query": "is_active = true"}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		// No user context set

		w := httptest.NewRecorder()
		handler.QueryOQL(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})
}

// TestOQLValidateEndpoint_Integration tests the validation endpoint
func TestOQLValidateEndpoint_Integration(t *testing.T) {
	executor := &mockQueryExecutor{}
	handler := NewQueryHandler(executor, &mockQueryTranslator{})

	t.Run("valid OQL query", func(t *testing.T) {
		body := `{"query": "is_active = true AND software.vendor = \"Microsoft\""}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql/validate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ValidateOQL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
			t.Logf("Response body: %s", w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["valid"] != true {
			t.Errorf("got valid=%v, want true", resp["valid"])
		}

		errors, ok := resp["errors"].([]interface{})
		if !ok {
			t.Fatal("errors field is not an array")
		}

		if len(errors) != 0 {
			t.Errorf("got %d errors, want 0", len(errors))
		}
	})

	t.Run("invalid OQL syntax", func(t *testing.T) {
		body := `{"query": "is_active == true"}` // Invalid operator

		req := httptest.NewRequest("POST", "/api/v1/query/oql/validate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ValidateOQL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["valid"] != false {
			t.Errorf("got valid=%v, want false", resp["valid"])
		}

		errors, ok := resp["errors"].([]interface{})
		if !ok {
			t.Fatal("errors field is not an array")
		}

		if len(errors) == 0 {
			t.Error("expected at least one error")
		}
	})

	t.Run("missing query field", func(t *testing.T) {
		body := `{}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql/validate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ValidateOQL(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", w.Code)
		}
	})

	t.Run("complex valid query", func(t *testing.T) {
		body := `{"query": "is_active = true AND (findings.severity = \"critical\" OR findings.epss_score > 0.9) sort canonical_name limit 10"}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql/validate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ValidateOQL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["valid"] != true {
			t.Errorf("got valid=%v, want true", resp["valid"])
		}
	})
}

// TestOQLExplainEndpoint_Integration tests the explain endpoint
func TestOQLExplainEndpoint_Integration(t *testing.T) {
	executor := &mockQueryExecutor{}
	handler := NewQueryHandler(executor, &mockQueryTranslator{})

	t.Run("explain simple OQL query", func(t *testing.T) {
		body := `{"query": "is_active = true limit 10"}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql/explain", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ExplainOQL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
			t.Logf("Response body: %s", w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["unified_query"] == nil {
			t.Fatal("expected unified_query field in response")
		}

		if resp["sql"] == nil {
			t.Fatal("expected sql field in response")
		}

		if resp["args"] == nil {
			t.Fatal("expected args field in response")
		}
	})

	t.Run("explain query with dot-walking", func(t *testing.T) {
		body := `{"query": "software.vendor = \"Microsoft\" AND findings.severity = \"critical\""}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql/explain", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ExplainOQL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
			t.Logf("Response body: %s", w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		unifiedQuery, ok := resp["unified_query"].(map[string]interface{})
		if !ok {
			t.Fatal("unified_query is not an object")
		}

		if unifiedQuery["filters"] == nil {
			t.Fatal("expected filters in unified_query")
		}
	})

	t.Run("explain query with NOT operator", func(t *testing.T) {
		body := `{"query": "NOT software.vendor = \"CrowdStrike\""}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql/explain", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ExplainOQL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
			t.Logf("Response body: %s", w.Body.String())
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		unifiedQuery, ok := resp["unified_query"].(map[string]interface{})
		if !ok {
			t.Fatal("unified_query is not an object")
		}

		filters, ok := unifiedQuery["filters"].([]interface{})
		if !ok {
			t.Fatal("filters is not an array")
		}

		if len(filters) == 0 {
			t.Fatal("expected at least one filter")
		}

		// Check that the filter has negate: true
		firstFilter, ok := filters[0].(map[string]interface{})
		if !ok {
			t.Fatal("filter is not an object")
		}

		if firstFilter["negate"] != true {
			t.Errorf("expected negate=true for NOT query, got %v", firstFilter["negate"])
		}
	})

	t.Run("explain query with invalid syntax", func(t *testing.T) {
		body := `{"query": "is_active == true"}`

		req := httptest.NewRequest("POST", "/api/v1/query/oql/explain", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.ExplainOQL(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["error"] == nil {
			t.Fatal("expected error field in response")
		}

		errResp, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("error field is not an object")
		}

		if errResp["code"] != "OQL_PARSE_ERROR" {
			t.Errorf("got error code %v, want OQL_PARSE_ERROR", errResp["code"])
		}
	})
}

// TestOQLRealWorldScenarios tests real-world OQL query scenarios
func TestOQLRealWorldScenarios(t *testing.T) {
	executor := &mockQueryExecutor{
		result: &query.QueryResult{
			Data: []map[string]interface{}{
				{"id": int64(1), "canonical_name": "webserver01.example.com"},
			},
			Meta: &query.QueryMeta{
				TotalRows:       1,
				ExecutionTimeMs: 3,
				HasMore:         false,
			},
		},
	}
	handler := NewQueryHandler(executor, &mockQueryTranslator{})

	scenarios := []struct {
		name   string
		query  string
		valid  bool
		desc   string
	}{
		{
			name:   "active assets with Microsoft software",
			query:  `is_active = true AND software.vendor = "Microsoft"`,
			valid:  true,
			desc:   "Find all active assets with Microsoft software installed",
		},
		{
			name:   "assets without EDR",
			query:  `is_active = true AND NOT software.vendor = "CrowdStrike"`,
			valid:  true,
			desc:   "Find active assets missing CrowdStrike EDR",
		},
		{
			name:   "critical vulnerabilities",
			query:  `findings.severity = "critical" AND findings.epss_score > 0.9`,
			valid:  true,
			desc:   "Find assets with critical exploitable vulnerabilities",
		},
		{
			name:   "known exploited vulnerabilities",
			query:  `findings.is_kev = true AND findings.effective_status = "open"`,
			valid:  true,
			desc:   "Find assets with known exploited vulnerabilities",
		},
		{
			name:   "production high-risk assets",
			query:  `environment = "production" AND (findings.severity = "critical" OR findings.epss_score > 0.9)`,
			valid:  true,
			desc:   "Find production assets with critical or exploitable vulnerabilities",
		},
		{
			name:   "complex nested logic",
			query:  `is_active = true AND (software.vendor = "Microsoft" OR software.vendor = "Oracle") AND NOT findings.severity = "low"`,
			valid:  true,
			desc:   "Complex query with nested OR and NOT",
		},
		{
			name:   "with sorting and pagination",
			query:  `is_active = true sort last_seen_at desc limit 20 offset 40`,
			valid:  true,
			desc:   "Active assets with sorting and pagination",
		},
		{
			name:   "pattern matching",
			query:  `hostname like "web%" AND is_active = true`,
			valid:  true,
			desc:   "Find assets with hostname matching pattern",
		},
		{
			name:   "in operator",
			query:  `findings.severity in ("critical", "high") AND findings.effective_status = "open"`,
			valid:  true,
			desc:   "Find assets with critical or high severity findings",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("Scenario: %s", scenario.desc)
			t.Logf("Query: %s", scenario.query)

			// First validate
			validateBody := map[string]string{"query": scenario.query}
			validateJSON, _ := json.Marshal(validateBody)

			req := httptest.NewRequest("POST", "/api/v1/query/oql/validate", bytes.NewReader(validateJSON))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.ValidateOQL(w, req)

			var validateResp map[string]interface{}
			json.NewDecoder(w.Body).Decode(&validateResp)

		 isValid, hasValid := validateResp["valid"].(bool)
		 if !hasValid {
				t.Errorf("validation response missing 'valid' field")
				return
		 }

		 if isValid != scenario.valid {
				t.Errorf("validation got valid=%v, want %v", isValid, scenario.valid)
		 }

			// If valid, also test execution
			if isValid {
				queryBody := map[string]string{"query": scenario.query}
				queryJSON, _ := json.Marshal(queryBody)

				req = httptest.NewRequest("POST", "/api/v1/query/oql", bytes.NewReader(queryJSON))
				req.Header.Set("Content-Type", "application/json")
				req = setQueryUserContextWithRequestID(req, 1, "test-scenario-"+scenario.name)

				w = httptest.NewRecorder()
				handler.QueryOQL(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("execution failed with status %d: %s", w.Code, w.Body.String())
				}
			}
		})
	}
}

