package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openexposuremanagement/oem/internal/auth"
	"github.com/openexposuremanagement/oem/internal/database"
	"github.com/openexposuremanagement/oem/internal/ingest"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/stretchr/testify/assert"
)

// TestIngestVMFindings_EndToEnd_Integration tests the full ingestion flow with a real database
func TestIngestVMFindings_EndToEnd_Integration(t *testing.T) {
	// Setup test database
	db := database.SetupTestDB(t)
	if db == nil {
		return // Test was skipped
	}

	ctx := context.Background()

	t.Run("successful_ingestion_creates_assets_findings", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		// Create request payload
		scannedAt := time.Now().UTC()
		payload := map[string]interface{}{
			"source":     "tenable",
			"scanner":    "Nessus",
			"scanned_at": scannedAt.Format(time.RFC3339),
			"findings": []map[string]interface{}{
				{
					"asset": map[string]interface{}{
						"hostname":     "webserver01.example.com",
						"ip_addresses": []string{"192.168.1.100"},
						"external_ids": map[string]interface{}{
							"aws:instance_id": "i-1234567890abcdef0",
						},
					},
					"finding": map[string]interface{}{
						"definition_id": "12345",
						"title":         "Apache Log4j Remote Code Execution",
						"severity":      "Critical",
						"cves":          []string{"CVE-2021-44228"},
						"references":    []string{"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2021-44228"},
					},
					"status":      "open",
					"first_found": scannedAt.Add(-24 * time.Hour).Format(time.RFC3339),
					"last_found":  scannedAt.Format(time.RFC3339),
					"evidence": map[string]interface{}{
						"port":     443,
						"protocol": "https",
					},
				},
			},
		}

		data, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")

		// Add user context to request
		userCtx := &auth.UserContext{
			UserID:   "test-user-id",
			TenantID: tenantID,
			Email:    "ingest-test@example.com",
		}
		ctx = context.WithValue(context.Background(), auth.UserContextKey, userCtx)
		req = req.WithContext(ctx)

		// Create handler
		handler := IngestVMFindings(db)
		w := httptest.NewRecorder()

		// Execute request
		handler.ServeHTTP(w, req)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "success", response["status"])

		summary := response["summary"].(map[string]interface{})
		assert.Equal(t, float64(1), summary["total_findings"])
		assert.Equal(t, float64(1), summary["assets_created"])
		assert.Equal(t, float64(0), summary["assets_updated"])
		assert.Equal(t, float64(1), summary["definitions_processed"])
		assert.Equal(t, float64(1), summary["findings_upserted"])

		// Verify asset was created in database
		assetRepo := ingest.NewAssetMatcher(db, tenantID, "tenable")
		vmAsset := &ingest.VMAsset{
			Hostname:     "webserver01.example.com",
			IPAddresses:  []string{"192.168.1.100"},
			ExternalIDs: map[string]string{"aws:instance_id": "i-1234567890abcdef0"},
		}

		result, err := assetRepo.MatchAsset(ctx, vmAsset)
		assert.NoError(t, err)
		assert.False(t, result.NewAsset, "Asset should already exist from ingestion")
		assert.NotNil(t, result.Asset)

		// Verify definition was created
		definitionUID := ingest.GenerateDefinitionUID("tenable", "12345")
		defRepo := repository.NewDefinitionRepository(db)
		def, err := defRepo.GetDefinition(ctx, definitionUID)
		assert.NoError(t, err)
		assert.Equal(t, "Apache Log4j Remote Code Execution", def.Title)
		assert.Equal(t, "Critical", def.SeverityDefault)

		// Verify CVE aliases were created
		aliases, err := defRepo.GetAliasesForDefinition(ctx, definitionUID)
		assert.NoError(t, err)
		assert.Len(t, aliases, 1)
		assert.Equal(t, "CVE-2021-44228", aliases[0].AliasValue)

		// Verify finding instance was created
		findingRepo := repository.NewFindingInstanceRepository(db)
		instance, err := findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, result.Asset.ID, definitionUID)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
		assert.Equal(t, "open", instance.ScannerStatus)
		assert.Equal(t, "open", instance.EffectiveStatus)
		assert.Equal(t, "scanner", instance.EffectiveReason)
		assert.Equal(t, int64(1), instance.EffectiveRevision)
	})

	t.Run("idempotent_on_repeated_ingestion", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		scannedAt := time.Now().UTC()
		payload := map[string]interface{}{
			"source":     "tenable",
			"scanner":    "Nessus",
			"scanned_at": scannedAt.Format(time.RFC3339),
			"findings": []map[string]interface{}{
				{
					"asset": map[string]interface{}{
						"hostname": "idempotent-server.example.com",
					},
					"finding": map[string]interface{}{
						"definition_id": "54321",
						"title":         "Test Vulnerability",
						"severity":      "High",
					},
					"status":     "open",
					"last_found": scannedAt.Format(time.RFC3339),
				},
			},
		}

		data, _ := json.Marshal(payload)

		// Helper function to create and execute request
		userCtx := &auth.UserContext{
			UserID:   "test-user-id",
			TenantID: tenantID,
			Email:    "idempotent-test@example.com",
		}

		executeRequest := func() *httptest.ResponseRecorder {
			req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(data))
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(context.Background(), auth.UserContextKey, userCtx)
			req = req.WithContext(ctx)

			handler := IngestVMFindings(db)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			return w
		}

		// First ingestion
		w1 := executeRequest()
		assert.Equal(t, http.StatusOK, w1.Code)

		var response1 map[string]interface{}
		json.Unmarshal(w1.Body.Bytes(), &response1)
		summary1 := response1["summary"].(map[string]interface{})
		assert.Equal(t, float64(1), summary1["assets_created"])
		assert.Equal(t, float64(0), summary1["assets_updated"])

		// Second ingestion (same payload)
		w2 := executeRequest()
		assert.Equal(t, http.StatusOK, w2.Code)

		var response2 map[string]interface{}
		json.Unmarshal(w2.Body.Bytes(), &response2)
		summary2 := response2["summary"].(map[string]interface{})
		assert.Equal(t, float64(0), summary2["assets_created"], "Should not create duplicate asset")
		assert.Equal(t, float64(1), summary2["assets_updated"], "Should update existing asset")

		// Verify only one asset exists
		assetRepo := repository.NewAssetRepository(db)
		params := repository.AssetListParams{
			TenantID: tenantID,
		}
		result, err := assetRepo.List(ctx, params)
		assert.NoError(t, err)
		assert.Len(t, result.Assets, 1, "Should have exactly one asset")
	})

	t.Run("handles_multiple_findings_in_single_payload", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		scannedAt := time.Now().UTC()
		payload := map[string]interface{}{
			"source":     "tenable",
			"scanner":    "Nessus",
			"scanned_at": scannedAt.Format(time.RFC3339),
			"findings": []map[string]interface{}{
				{
					"asset": map[string]interface{}{
						"hostname": "server-01.example.com",
					},
					"finding": map[string]interface{}{
						"definition_id": "def-001",
						"title":         "Vulnerability 1",
						"severity":      "Critical",
					},
					"status":     "open",
					"last_found": scannedAt.Format(time.RFC3339),
				},
				{
					"asset": map[string]interface{}{
						"hostname": "server-02.example.com",
					},
					"finding": map[string]interface{}{
						"definition_id": "def-002",
						"title":         "Vulnerability 2",
						"severity":      "High",
					},
					"status":     "open",
					"last_found": scannedAt.Format(time.RFC3339),
				},
				{
					"asset": map[string]interface{}{
						"hostname": "server-03.example.com",
					},
					"finding": map[string]interface{}{
						"definition_id": "def-003",
						"title":         "Vulnerability 3",
						"severity":      "Medium",
					},
					"status":     "open",
					"last_found": scannedAt.Format(time.RFC3339),
				},
			},
		}

		data, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")

		userCtx := &auth.UserContext{
			UserID:   "test-user-id",
			TenantID: tenantID,
			Email:    "multi-test@example.com",
		}
		ctx := context.WithValue(context.Background(), auth.UserContextKey, userCtx)
		req = req.WithContext(ctx)

		handler := IngestVMFindings(db)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		summary := response["summary"].(map[string]interface{})
		assert.Equal(t, float64(3), summary["total_findings"])
		assert.Equal(t, float64(3), summary["assets_created"])
		assert.Equal(t, float64(3), summary["findings_upserted"])
	})

	t.Run("existing_asset_finding_now_fixed", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		initialTime := time.Now().UTC().Add(-24 * time.Hour)

		// First ingestion: finding is open
		payloadOpen := map[string]interface{}{
			"source":     "tenable",
			"scanner":    "Nessus",
			"scanned_at": initialTime.Format(time.RFC3339),
			"findings": []map[string]interface{}{
				{
					"asset": map[string]interface{}{
						"hostname": "fixed-server.example.com",
					},
					"finding": map[string]interface{}{
						"definition_id": "99999",
						"title":         "Fixed Vulnerability",
						"severity":      "High",
					},
					"status":     "open",
					"last_found": initialTime.Format(time.RFC3339),
				},
			},
		}

		userCtx := &auth.UserContext{
			UserID:   "test-user-id",
			TenantID: tenantID,
			Email:    "fixed-test@example.com",
		}

		data, _ := json.Marshal(payloadOpen)
		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")

		ctx := context.WithValue(context.Background(), auth.UserContextKey, userCtx)
		req = req.WithContext(ctx)

		handler := IngestVMFindings(db)
		w1 := httptest.NewRecorder()
		handler.ServeHTTP(w1, req)

		// Verify initial state
		definitionUID := ingest.GenerateDefinitionUID("tenable", "99999")
		findingRepo := repository.NewFindingInstanceRepository(db)

		// Get the asset ID first
		assetRepo := ingest.NewAssetMatcher(db, tenantID, "tenable")
		vmAsset := &ingest.VMAsset{Hostname: "fixed-server.example.com"}
		assetResult, _ := assetRepo.MatchAsset(ctx, vmAsset)

		instance, _ := findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, assetResult.Asset.ID, definitionUID)
		assert.Equal(t, "open", instance.ScannerStatus)
		assert.Equal(t, "open", instance.EffectiveStatus)

		// Second ingestion: finding is now fixed
		scannedAt := time.Now().UTC()
		payloadFixed := map[string]interface{}{
			"source":     "tenable",
			"scanner":    "Nessus",
			"scanned_at": scannedAt.Format(time.RFC3339),
			"findings": []map[string]interface{}{
				{
					"asset": map[string]interface{}{
						"hostname": "fixed-server.example.com",
					},
					"finding": map[string]interface{}{
						"definition_id": "99999",
						"title":         "Fixed Vulnerability",
						"severity":      "High",
					},
					"status":     "fixed",
					"last_found": scannedAt.Format(time.RFC3339),
				},
			},
		}

		dataFixed, _ := json.Marshal(payloadFixed)
		reqFixed := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(dataFixed))
		reqFixed.Header.Set("Content-Type", "application/json")
		reqFixed = reqFixed.WithContext(ctx)

		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, reqFixed)

		assert.Equal(t, http.StatusOK, w2.Code)

		// Verify status was updated
		instance, _ = findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, assetResult.Asset.ID, definitionUID)
		assert.Equal(t, "fixed", instance.ScannerStatus)
		assert.Equal(t, "fixed", instance.EffectiveStatus, "Effective status should be 'fixed'")
		assert.Equal(t, "scanner", instance.EffectiveReason)
	})

	t.Run("returns_validation_errors", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		// Missing source
		payload := map[string]interface{}{
			"findings": []map[string]interface{}{
				{
					"asset": map[string]interface{}{
						"hostname": "test.example.com",
					},
					"finding": map[string]interface{}{
						"definition_id": "12345",
						"title":         "Test",
					},
					"status":     "open",
					"last_found": time.Now().Format(time.RFC3339),
				},
			},
		}

		data, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")

		userCtx := &auth.UserContext{
			UserID:   "test-user-id",
			TenantID: tenantID,
			Email:    "validation-test@example.com",
		}
		ctx := context.WithValue(context.Background(), auth.UserContextKey, userCtx)
		req = req.WithContext(ctx)

		handler := IngestVMFindings(db)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "error")
	})

	t.Run("returns_error_for_invalid_json", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		req := httptest.NewRequest("POST", "/ingest/vm/findings", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		userCtx := &auth.UserContext{
			UserID:   "test-user-id",
			TenantID: tenantID,
			Email:    "json-test@example.com",
		}
		ctx := context.WithValue(context.Background(), auth.UserContextKey, userCtx)
		req = req.WithContext(ctx)

		handler := IngestVMFindings(db)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
