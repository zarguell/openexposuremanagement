package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/openexposuremanagement/oem/internal/handlers"
	"github.com/openexposuremanagement/oem/internal/services/query"
)

// Mock executor for testing
type mockExecutor struct{}

func (m *mockExecutor) Execute(ctx context.Context, tenantID string, entityType string, q *query.Query) (*query.QueryResult, error) {
	return &query.QueryResult{
		Data: []map[string]interface{}{
			{
				"id":            int64(1),
				"canonical_name": "test-asset-01.example.com",
				"is_active":      true,
				"last_seen_at":   "2024-01-15T10:00:00Z",
			},
			{
				"id":            int64(2),
				"canonical_name": "test-asset-02.example.com",
				"is_active":      true,
				"last_seen_at":   "2024-01-15T11:00:00Z",
			},
		},
		Meta: &query.QueryMeta{
			TotalRows:       2,
			ExecutionTimeMs: 3,
			HasMore:         false,
		},
	}, nil
}

// Mock translator for testing
type mockTranslator struct{}

func (m *mockTranslator) Translate(entityType string, q *query.Query) (string, []interface{}, error) {
	return "SELECT * FROM assets WHERE tenant_id = $1", []interface{}{}, nil
}

func main() {
	fmt.Println("🧪 OQL Backend Test Suite")
	fmt.Println("========================")
	fmt.Println()

	// Create handler
	executor := &mockExecutor{}
	translator := &mockTranslator{}
	handler := handlers.NewQueryHandler(executor, translator)

	// Track test results
	totalTests := 0
	passedTests := 0
	failedTests := 0

	// Test 1: Validate Endpoint - Valid Query
	fmt.Println("Test 1: Validate Endpoint - Valid Query")
	totalTests++
	if testValidateValidQuery(handler) {
		fmt.Println("✅ PASS")
		passedTests++
	} else {
		fmt.Println("❌ FAIL")
		failedTests++
	}
	fmt.Println()

	// Test 2: Validate Endpoint - Invalid Query
	fmt.Println("Test 2: Validate Endpoint - Invalid Query")
	totalTests++
	if testValidateInvalidQuery(handler) {
		fmt.Println("✅ PASS")
		passedTests++
	} else {
		fmt.Println("❌ FAIL")
		failedTests++
	}
	fmt.Println()

	// Test 3: Explain Endpoint - Simple Query
	fmt.Println("Test 3: Explain Endpoint - Simple Query")
	totalTests++
	if testExplainSimpleQuery(handler) {
		fmt.Println("✅ PASS")
		passedTests++
	} else {
		fmt.Println("❌ FAIL")
		failedTests++
	}
	fmt.Println()

	// Test 4: Query Endpoint - Simple Query
	fmt.Println("Test 4: Query Endpoint - Simple Query")
	totalTests++
	if testQuerySimple(handler) {
		fmt.Println("✅ PASS")
		passedTests++
	} else {
		fmt.Println("❌ FAIL")
		failedTests++
	}
	fmt.Println()

	// Test 5: Query Endpoint - Complex Query with Dot-Walking
	fmt.Println("Test 5: Query Endpoint - Complex Query with Dot-Walking")
	totalTests++
	if testQueryComplex(handler) {
		fmt.Println("✅ PASS")
		passedTests++
	} else {
		fmt.Println("❌ FAIL")
		failedTests++
	}
	fmt.Println()

	// Test 6: Query Endpoint - NOT Operator (Anti-Join)
	fmt.Println("Test 6: Query Endpoint - NOT Operator (Anti-Join)")
	totalTests++
	if testQueryNotOperator(handler) {
		fmt.Println("✅ PASS")
		passedTests++
	} else {
		fmt.Println("❌ FAIL")
		failedTests++
	}
	fmt.Println()

	// Test 7: Query Endpoint - Sort Clause
	fmt.Println("Test 7: Query Endpoint - Sort Clause")
	totalTests++
	if testQuerySort(handler) {
		fmt.Println("✅ PASS")
		passedTests++
	} else {
		fmt.Println("❌ FAIL")
		failedTests++
	}
	fmt.Println()

	// Test 8: Query Endpoint - Invalid Syntax
	fmt.Println("Test 8: Query Endpoint - Invalid Syntax")
	totalTests++
	if testQueryInvalidSyntax(handler) {
		fmt.Println("✅ PASS")
		passedTests++
	} else {
		fmt.Println("❌ FAIL")
		failedTests++
	}
	fmt.Println()

	// Test 9: Real-World Scenario - Assets without EDR
	fmt.Println("Test 9: Real-World Scenario - Assets without EDR")
	totalTests++
	if testScenarioAssetsWithoutEDR(handler) {
		fmt.Println("✅ PASS")
		passedTests++
	} else {
		fmt.Println("❌ FAIL")
		failedTests++
	}
	fmt.Println()

	// Test 10: Real-World Scenario - Critical Vulnerabilities
	fmt.Println("Test 10: Real-World Scenario - Critical Vulnerabilities")
	totalTests++
	if testScenarioCriticalVulns(handler) {
		fmt.Println("✅ PASS")
		passedTests++
	} else {
		fmt.Println("❌ FAIL")
		failedTests++
	}
	fmt.Println()

	// Print summary
	fmt.Println("========================")
	fmt.Println("Test Summary")
	fmt.Println("========================")
	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed: %d ✅\n", passedTests)
	fmt.Printf("Failed: %d ❌\n", failedTests)
	fmt.Printf("Success Rate: %.1f%%\n", float64(passedTests)/float64(totalTests)*100)
	fmt.Println()

	if failedTests == 0 {
		fmt.Println("🎉 All tests passed! OQL backend is working correctly.")
		os.Exit(0)
	} else {
		fmt.Println("⚠️  Some tests failed. Please review the errors above.")
		os.Exit(1)
	}
}

// Helper function to create request with user context
func createRequest(method, url string, body []byte) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Add user context
	ctx := req.Context()
	ctx = context.WithValue(ctx, auth.UserContextKey, &auth.UserContext{
		UserID:    "1",
		TenantID:  1,
		Email:     "test@example.com",
		Roles:     []string{"analyst"},
	})
	ctx = context.WithValue(ctx, "request_id", "test-oql-backend-"+time.Now().Format("20060102150405"))
	return req.WithContext(ctx)
}

// Test: Validate valid query
func testValidateValidQuery(handler *handlers.QueryHandler) bool {
	body := []byte(`{"query": "is_active = true AND software.vendor = \"Microsoft\""}`)
	req := createRequest("POST", "/api/v1/query/oql/validate", body)
	w := httptest.NewRecorder()

	handler.ValidateOQL(w, req)

	if w.Code != http.StatusOK {
		fmt.Printf("  Expected status 200, got %d\n", w.Code)
		fmt.Printf("  Response: %s\n", w.Body.String())
		return false
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		fmt.Printf("  Failed to decode response: %v\n", err)
		return false
	}

	if resp["valid"] != true {
		fmt.Printf("  Expected valid=true, got %v\n", resp["valid"])
		return false
	}

	fmt.Println("  Query validated successfully")
	return true
}

// Test: Validate invalid query
func testValidateInvalidQuery(handler *handlers.QueryHandler) bool {
	body := []byte(`{"query": "is_active == true"}`) // Invalid operator
	req := createRequest("POST", "/api/v1/query/oql/validate", body)
	w := httptest.NewRecorder()

	handler.ValidateOQL(w, req)

	if w.Code != http.StatusOK {
		fmt.Printf("  Expected status 200, got %d\n", w.Code)
		return false
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		fmt.Printf("  Failed to decode response: %v\n", err)
		return false
	}

	if resp["valid"] != false {
		fmt.Printf("  Expected valid=false, got %v\n", resp["valid"])
		return false
	}

	errors, ok := resp["errors"].([]interface{})
	if !ok || len(errors) == 0 {
		fmt.Println("  Expected error messages in response")
		return false
	}

	fmt.Printf("  Correctly identified invalid query with %d error(s)\n", len(errors))
	return true
}

// Test: Explain simple query
func testExplainSimpleQuery(handler *handlers.QueryHandler) bool {
	body := []byte(`{"query": "is_active = true limit 10"}`)
	req := createRequest("POST", "/api/v1/query/oql/explain", body)
	w := httptest.NewRecorder()

	handler.ExplainOQL(w, req)

	if w.Code != http.StatusOK {
		fmt.Printf("  Expected status 200, got %d\n", w.Code)
		fmt.Printf("  Response: %s\n", w.Body.String())
		return false
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		fmt.Printf("  Failed to decode response: %v\n", err)
		return false
	}

	if resp["unified_query"] == nil {
		fmt.Println("  Expected unified_query in response")
		return false
	}

	if resp["sql"] == nil {
		fmt.Println("  Expected sql in response")
		return false
	}

	fmt.Println("  Query explanation generated successfully")
	return true
}

// Test: Query simple
func testQuerySimple(handler *handlers.QueryHandler) bool {
	body := []byte(`{"query": "is_active = true limit 10"}`)
	req := createRequest("POST", "/api/v1/query/oql", body)
	w := httptest.NewRecorder()

	handler.QueryOQL(w, req)

	if w.Code != http.StatusOK {
		fmt.Printf("  Expected status 200, got %d\n", w.Code)
		fmt.Printf("  Response: %s\n", w.Body.String())
		return false
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		fmt.Printf("  Failed to decode response: %v\n", err)
		return false
	}

	if resp["data"] == nil {
		fmt.Println("  Expected data field in response")
		return false
	}

	if resp["meta"] == nil {
		fmt.Println("  Expected meta field in response")
		return false
	}

	data, ok := resp["data"].([]interface{})
	if !ok {
		fmt.Println("  Data is not an array")
		return false
	}

	fmt.Printf("  Returned %d results\n", len(data))
	return true
}

// Test: Query complex with dot-walking
func testQueryComplex(handler *handlers.QueryHandler) bool {
	body := []byte(`{"query": "software.vendor = \"Microsoft\" AND findings.severity = \"critical\""}`)
	req := createRequest("POST", "/api/v1/query/oql", body)
	w := httptest.NewRecorder()

	handler.QueryOQL(w, req)

	if w.Code != http.StatusOK {
		fmt.Printf("  Expected status 200, got %d\n", w.Code)
		fmt.Printf("  Response: %s\n", w.Body.String())
		return false
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		fmt.Printf("  Failed to decode response: %v\n", err)
		return false
	}

	if resp["data"] == nil {
		fmt.Println("  Expected data field in response")
		return false
	}

	fmt.Println("  Dot-walking query executed successfully")
	return true
}

// Test: Query with NOT operator
func testQueryNotOperator(handler *handlers.QueryHandler) bool {
	body := []byte(`{"query": "NOT software.vendor = \"CrowdStrike\""}`)
	req := createRequest("POST", "/api/v1/query/oql", body)
	w := httptest.NewRecorder()

	handler.QueryOQL(w, req)

	if w.Code != http.StatusOK {
		fmt.Printf("  Expected status 200, got %d\n", w.Code)
		fmt.Printf("  Response: %s\n", w.Body.String())
		return false
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		fmt.Printf("  Failed to decode response: %v\n", err)
		return false
	}

	if resp["data"] == nil {
		fmt.Println("  Expected data field in response")
		return false
	}

	fmt.Println("  NOT operator (anti-join) query executed successfully")
	return true
}

// Test: Query with sort clause
func testQuerySort(handler *handlers.QueryHandler) bool {
	body := []byte(`{"query": "is_active = true sort canonical_name desc limit 10"}`)
	req := createRequest("POST", "/api/v1/query/oql", body)
	w := httptest.NewRecorder()

	handler.QueryOQL(w, req)

	if w.Code != http.StatusOK {
		fmt.Printf("  Expected status 200, got %d\n", w.Code)
		fmt.Printf("  Response: %s\n", w.Body.String())
		return false
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		fmt.Printf("  Failed to decode response: %v\n", err)
		return false
	}

	if resp["data"] == nil {
		fmt.Println("  Expected data field in response")
		return false
	}

	fmt.Println("  Sort clause query executed successfully")
	return true
}

// Test: Query with invalid syntax
func testQueryInvalidSyntax(handler *handlers.QueryHandler) bool {
	body := []byte(`{"query": "is_active == true"}`) // Invalid operator
	req := createRequest("POST", "/api/v1/query/oql", body)
	w := httptest.NewRecorder()

	handler.QueryOQL(w, req)

	if w.Code != http.StatusBadRequest {
		fmt.Printf("  Expected status 400, got %d\n", w.Code)
		return false
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		fmt.Printf("  Failed to decode response: %v\n", err)
		return false
	}

	if resp["error"] == nil {
		fmt.Println("  Expected error field in response")
		return false
	}

	errResp, ok := resp["error"].(map[string]interface{})
	if !ok {
		fmt.Println("  Error field is not an object")
		return false
	}

	if errResp["code"] != "OQL_PARSE_ERROR" {
		fmt.Printf("  Expected error code OQL_PARSE_ERROR, got %v\n", errResp["code"])
		return false
	}

	fmt.Println("  Invalid syntax correctly rejected with error")
	return true
}

// Test: Assets without EDR scenario
func testScenarioAssetsWithoutEDR(handler *handlers.QueryHandler) bool {
	body := []byte(`{"query": "is_active = true AND NOT software.vendor = \"CrowdStrike\""}`)
	req := createRequest("POST", "/api/v1/query/oql", body)
	w := httptest.NewRecorder()

	handler.QueryOQL(w, req)

	if w.Code != http.StatusOK {
		fmt.Printf("  Expected status 200, got %d\n", w.Code)
		fmt.Printf("  Response: %s\n", w.Body.String())
		return false
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		fmt.Printf("  Failed to decode response: %v\n", err)
		return false
	}

	if resp["data"] == nil {
		fmt.Println("  Expected data field in response")
		return false
	}

	fmt.Println("  'Assets without EDR' scenario executed successfully")
	return true
}

// Test: Critical vulnerabilities scenario
func testScenarioCriticalVulns(handler *handlers.QueryHandler) bool {
	body := []byte(`{"query": "findings.severity = \"critical\" AND findings.epss_score > 0.9"}`)
	req := createRequest("POST", "/api/v1/query/oql", body)
	w := httptest.NewRecorder()

	handler.QueryOQL(w, req)

	if w.Code != http.StatusOK {
		fmt.Printf("  Expected status 200, got %d\n", w.Code)
		fmt.Printf("  Response: %s\n", w.Body.String())
		return false
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		fmt.Printf("  Failed to decode response: %v\n", err)
		return false
	}

	if resp["data"] == nil {
		fmt.Println("  Expected data field in response")
		return false
	}

	fmt.Println("  'Critical vulnerabilities' scenario executed successfully")
	return true
}
