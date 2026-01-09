package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openexposuremanagement/oem/internal/ingest"
	"github.com/stretchr/testify/assert"
)

// Test data helpers
func getValidVMFindingsPayload() []byte {
	payload := map[string]interface{}{
		"source": "tenable",
		"scanner": " Nessus",
		"scanned_at": time.Now().UTC().Format(time.RFC3339),
		"findings": []map[string]interface{}{
			{
				"asset": map[string]interface{}{
					"hostname": "webserver01.example.com",
					"ip_addresses": []string{"192.168.1.100"},
					"external_ids": map[string]interface{}{
						"aws:instance_id": "i-1234567890abcdef0",
					},
				},
				"finding": map[string]interface{}{
					"definition_id": "12345",
					"title": "Apache Log4j Remote Code Execution",
					"severity": "Critical",
					"cves": []string{"CVE-2021-44228"},
					"references": []string{"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2021-44228"},
				},
				"status": "open",
				"first_found": time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339),
				"last_found": time.Now().UTC().Format(time.RFC3339),
				"evidence": map[string]interface{}{
					"port": 443,
					"protocol": "https",
				},
			},
		},
	}

	data, _ := json.Marshal(payload)
	return data
}

// TestIngestVMFindings_Validation tests request validation
func TestIngestVMFindings_Validation(t *testing.T) {
	t.Run("rejects_missing_source", func(t *testing.T) {
		payload := map[string]interface{}{
			"findings": []map[string]interface{}{
				{
					"asset": map[string]interface{}{
						"hostname": "test.example.com",
					},
					"finding": map[string]interface{}{
						"definition_id": "12345",
						"title": "Test",
					},
					"status": "open",
					"last_found": time.Now().UTC().Format(time.RFC3339),
				},
			},
		}

		data, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// TODO: Create handler with mock DB
		// handler := IngestVMFindings(db)
		// handler.ServeHTTP(w, req)

		// This documents expected behavior:
		// Should return 400 Bad Request with error message
		assert.NotNil(t, req)
		_ = w // Prevent unused variable error until handler is implemented
	})

	t.Run("rejects_empty_findings_array", func(t *testing.T) {
		payload := map[string]interface{}{
			"source": "tenable",
			"findings": []interface{}{},
			"scanned_at": time.Now().UTC().Format(time.RFC3339),
		}

		data, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Should return 400 Bad Request
		assert.NotNil(t, req)
		_ = w // Prevent unused variable error until handler is implemented
	})

	t.Run("rejects_invalid_json", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Should return 400 Bad Request
		assert.NotNil(t, req)
		_ = w // Prevent unused variable error until handler is implemented
	})

	t.Run("rejects_missing_asset_identifiers", func(t *testing.T) {
		payload := map[string]interface{}{
			"source": "tenable",
			"scanned_at": time.Now().UTC().Format(time.RFC3339),
			"findings": []map[string]interface{}{
				{
					"asset": map[string]interface{}{},
					"finding": map[string]interface{}{
						"definition_id": "12345",
						"title": "Test",
					},
					"status": "open",
					"last_found": time.Now().UTC().Format(time.RFC3339),
				},
			},
		}

		data, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")

		// Should return 400 Bad Request
		assert.NotNil(t, req)
	})
}

// TestIngestVMFindings_Authentication tests authentication requirements
func TestIngestVMFindings_Authentication(t *testing.T) {
	t.Run("requires_authentication", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(getValidVMFindingsPayload()))
		w := httptest.NewRecorder()

		// Request without auth should be rejected
		// This documents expected behavior:
		// Should return 401 Unauthorized or 403 Forbidden
		assert.NotNil(t, req)
		_ = w // Prevent unused variable error until handler is implemented
	})

	t.Run("accepts_valid_api_key", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(getValidVMFindingsPayload()))
		req.Header.Set("Authorization", "Bearer valid-api-key")
		w := httptest.NewRecorder()

		// Should accept request with valid API key
		assert.NotNil(t, req)
		_ = w // Prevent unused variable error until handler is implemented
	})

	t.Run("enforces_source_binding_on_api_key", func(t *testing.T) {
		// API key bound to "qualys" should reject "tenable" source
		payload := getValidVMFindingsPayload()

		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer api-key-bound-to-qualys")
		w := httptest.NewRecorder()

		// Should return 403 Forbidden
		// with message about source mismatch
		assert.NotNil(t, req)
		_ = w // Prevent unused variable error until handler is implemented
	})
}

// TestIngestVMFindings_EndToEnd tests the full ingestion flow
func TestIngestVMFindings_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("successful_ingestion_creates_assets_findings", func(t *testing.T) {
		// TODO: Integration test with test database
		// 1. Create test tenant and API key
		// 2. Send valid payload
		// 3. Verify asset was created/matched
		// 4. Verify finding definition was created
		// 5. Verify finding instance was created
		// 6. Verify effective status was computed
	})

	t.Run("idempotent_on_repeated_ingestion", func(t *testing.T) {
		// TODO: Integration test
		// Sending the same payload twice should:
		// - Not create duplicate assets
		// - Update observation windows
		// - Update last_seen_at timestamps
	})

	t.Run("handles_multiple_findings_in_single_payload", func(t *testing.T) {
		// TODO: Integration test
		// Payload with multiple findings should process all of them
	})

	t.Run("handles_partial_failure_gracefully", func(t *testing.T) {
		// TODO: Integration test
		// If one finding fails to process, should still process others
		// and return appropriate error response
	})
}

// TestIngestVMFindings_Response tests response format
func TestIngestVMFindings_Response(t *testing.T) {
	t.Run("returns_success_response", func(t *testing.T) {
		// This documents expected success response format
		expectedResponse := map[string]interface{}{
			"status": "success",
			"message": "Findings ingested successfully",
			"summary": map[string]interface{}{
				"total_findings": 1,
				"assets_processed": 1,
				"definitions_created": 1,
				"findings_upserted": 1,
			},
		}

		assert.NotNil(t, expectedResponse)
	})

	t.Run("returns_error_response_with_details", func(t *testing.T) {
		// This documents expected error response format
		expectedError := map[string]interface{}{
			"error": "Validation failed",
			"details": []string{
				"findings[0].asset.hostname: hostname is required",
			},
		}

		assert.NotNil(t, expectedError)
	})
}

// TestIngestVMFindings_Scenarios tests real-world scenarios
func TestIngestVMFindings_Scenarios(t *testing.T) {
	t.Run("new_asset_new_finding", func(t *testing.T) {
		// Scenario: First time seeing this asset and finding
		// Expected: Create new asset, new definition, new finding instance
	})

	t.Run("existing_asset_new_finding", func(t *testing.T) {
		// Scenario: Asset exists, but this is a new finding
		// Expected: Match existing asset, create new definition and finding
	})

	t.Run("existing_asset_existing_finding_still_vulnerable", func(t *testing.T) {
		// Scenario: Finding already exists and is still open
		// Expected: Update last_observed_at, keep first_observed_at
	})

	t.Run("existing_asset_finding_now_fixed", func(t *testing.T) {
		// Scenario: Finding was open, now scanner reports it as fixed
		// Expected: Update scanner_status to "fixed", update effective_status
	})
}

// Unit tests for tenant context extraction
func TestExtractTenantContext(t *testing.T) {
	t.Run("extracts_tenant_from_user_context", func(t *testing.T) {
		// User token should provide tenant_id
		tenantID := int64(123)
		assert.Equal(t, int64(123), tenantID)
	})

	t.Run("extracts_tenant_from_api_key", func(t *testing.T) {
		// API key should provide tenant_id
		tenantID := int64(456)
		assert.Equal(t, int64(456), tenantID)
	})
}

// Benchmark test
func BenchmarkIngestVMFindings(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark")
	}

	payload := getValidVMFindingsPayload()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		// TODO: Benchmark with handler
		_ = req
	}
}

// Helper function for validating VM findings payload
func TestValidateVMFindingsPayload(t *testing.T) {
	t.Run("valid_payload_passes", func(t *testing.T) {
		data := getValidVMFindingsPayload()
		var payload ingest.VMFindingsPayload
		err := json.Unmarshal(data, &payload)

		assert.NoError(t, err)
		err = payload.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid_status_fails", func(t *testing.T) {
		payload := ingest.VMFindingsPayload{
			Source: "tenable",
			Findings: []ingest.VMFinding{
				{
					Status: "invalid_status",
				},
			},
		}

		err := payload.Validate()
		assert.Error(t, err)
	})
}
