package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/database"
	"github.com/openexposuremanagement/oem/internal/repository"
)

// TestUpsertFindingInstance_Integration tests the finding instance upsert functionality with a real database
func TestUpsertFindingInstance_Integration(t *testing.T) {
	// Setup test database
	db := database.SetupTestDB(t)
	if db == nil {
		return // Test was skipped
	}

	ctx := context.Background()
	source := "tenable"

	t.Run("creates new finding instance", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		// First create an asset and definition
		asset := createTestFindingAsset(t, db, tenantID, source, "test-asset-1.example.com")
		definitionUID := createTestDefinition(t, db, source, "def-001")

		scannedAt := time.Now().UTC()
		vmFinding := &VMFinding{
			Asset: VMAsset{
				Hostname: "test-asset-1.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "def-001",
				Title:        "Test Vulnerability",
				Severity:     "High",
			},
			Status:     "open",
			LastFound:  scannedAt,
			FirstFound: scannedAt,
			Evidence: map[string]interface{}{
				"port":      8080,
				"protocol":  "tcp",
				"component": "lib",
			},
		}

		err := UpsertFindingInstance(ctx, db, tenantID, asset.ID, definitionUID, vmFinding)
		if err != nil {
			t.Fatalf("UpsertFindingInstance failed: %v", err)
		}

		// Verify finding instance was created
		findingRepo := repository.NewFindingInstanceRepository(db)
		instance, err := findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, asset.ID, definitionUID)
		if err != nil {
			t.Fatalf("Failed to get finding instance: %v", err)
		}

		if instance.ScannerStatus != "open" {
			t.Errorf("Expected scanner status 'open', got '%s'", instance.ScannerStatus)
		}

		if instance.EffectiveStatus != "open" {
			t.Errorf("Expected effective status 'open', got '%s'", instance.EffectiveStatus)
		}

		if instance.EffectiveReason != "scanner" {
			t.Errorf("Expected effective reason 'scanner', got '%s'", instance.EffectiveReason)
		}

		if instance.EffectiveRevision != 1 {
			t.Errorf("Expected effective revision 1, got %d", instance.EffectiveRevision)
		}

		// Verify timestamps (truncate to match DB precision)
		if !instance.FirstObservedAt.Truncate(time.Second).Equal(scannedAt.Truncate(time.Second)) {
			t.Errorf("Expected FirstObservedAt %v, got %v", scannedAt.Truncate(time.Second), instance.FirstObservedAt.Truncate(time.Second))
		}

		if !instance.LastObservedAt.Truncate(time.Second).Equal(scannedAt.Truncate(time.Second)) {
			t.Errorf("Expected LastObservedAt %v, got %v", scannedAt.Truncate(time.Second), instance.LastObservedAt.Truncate(time.Second))
		}

		// Verify evidence was stored
		if instance.EvidenceJSON == nil {
			t.Error("Expected evidence to be stored")
		}

		if instance.EvidenceJSON["port"] != float64(8080) { // JSON numbers become float64
			t.Errorf("Expected evidence port 8080, got %v", instance.EvidenceJSON["port"])
		}
	})

	t.Run("updates existing finding instance timestamps", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		// Create asset and definition
		asset := createTestFindingAsset(t, db, tenantID, source, "test-asset-2.example.com")
		definitionUID := createTestDefinition(t, db, source, "def-002")

		// Create initial finding instance
		initialTime := time.Now().UTC().Add(-24 * time.Hour)
		vmFinding := &VMFinding{
			Asset: VMAsset{
				Hostname: "test-asset-2.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "def-002",
				Title:        "Update Test",
				Severity:     "Medium",
			},
			Status:     "open",
			LastFound:  initialTime,
			FirstFound: initialTime,
			Evidence: map[string]interface{}{
				"scan_id": "initial",
			},
		}

		err := UpsertFindingInstance(ctx, db, tenantID, asset.ID, definitionUID, vmFinding)
		if err != nil {
			t.Fatalf("First UpsertFindingInstance failed: %v", err)
		}

		// Update with later timestamp
		laterTime := time.Now().UTC()
		vmFinding.LastFound = laterTime
		vmFinding.Evidence = map[string]interface{}{
			"scan_id": "updated",
		}

		err = UpsertFindingInstance(ctx, db, tenantID, asset.ID, definitionUID, vmFinding)
		if err != nil {
			t.Fatalf("Second UpsertFindingInstance failed: %v", err)
		}

		// Verify timestamps were updated correctly
		findingRepo := repository.NewFindingInstanceRepository(db)
		instance, err := findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, asset.ID, definitionUID)
		if err != nil {
			t.Fatalf("Failed to get finding instance: %v", err)
		}

		// First observed should stay at the earlier time
		expectedFirst := initialTime.Truncate(time.Second)
		actualFirst := instance.FirstObservedAt.Truncate(time.Second)
		if !expectedFirst.Equal(actualFirst) {
			t.Errorf("Expected FirstObservedAt to remain at %v, got %v", expectedFirst, actualFirst)
		}

		// Last observed should be at the later time
		expectedLast := laterTime.Truncate(time.Second)
		actualLast := instance.LastObservedAt.Truncate(time.Second)
		if !expectedLast.Equal(actualLast) {
			t.Errorf("Expected LastObservedAt %v, got %v", expectedLast, actualLast)
		}

		// Evidence should be updated
		if instance.EvidenceJSON["scan_id"] != "updated" {
			t.Errorf("Expected evidence scan_id 'updated', got %v", instance.EvidenceJSON["scan_id"])
		}
	})

	t.Run("does not move first_observed_at earlier if later time provided", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		asset := createTestFindingAsset(t, db, tenantID, source, "test-asset-3.example.com")
		definitionUID := createTestDefinition(t, db, source, "def-003")

		// Create with earlier time
		earlierTime := time.Now().UTC().Add(-1 * time.Hour)
		vmFinding := &VMFinding{
			Asset: VMAsset{
				Hostname: "test-asset-3.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "def-003",
				Title:        "First Observed Test",
				Severity:     "Critical",
			},
			Status:     "open",
			LastFound:  earlierTime,
			FirstFound: earlierTime,
		}

		err := UpsertFindingInstance(ctx, db, tenantID, asset.ID, definitionUID, vmFinding)
		if err != nil {
			t.Fatalf("First UpsertFindingInstance failed: %v", err)
		}

		// Try to update with later first_found time (should not update)
		laterTime := time.Now().UTC()
		vmFinding.FirstFound = laterTime
		vmFinding.LastFound = laterTime

		err = UpsertFindingInstance(ctx, db, tenantID, asset.ID, definitionUID, vmFinding)
		if err != nil {
			t.Fatalf("Second UpsertFindingInstance failed: %v", err)
		}

		// Verify first_observed_at stayed at earlier time
		findingRepo := repository.NewFindingInstanceRepository(db)
		instance, err := findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, asset.ID, definitionUID)
		if err != nil {
			t.Fatalf("Failed to get finding instance: %v", err)
		}

		expectedFirst := earlierTime.Truncate(time.Second)
		actualFirst := instance.FirstObservedAt.Truncate(time.Second)
		if !expectedFirst.Equal(actualFirst) {
			t.Errorf("Expected FirstObservedAt to remain at %v, got %v", expectedFirst, actualFirst)
		}
	})

	t.Run("does not move last_observed_at earlier if earlier time provided", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		asset := createTestFindingAsset(t, db, tenantID, source, "test-asset-4.example.com")
		definitionUID := createTestDefinition(t, db, source, "def-004")

		// Create with later time
		laterTime := time.Now().UTC()
		vmFinding := &VMFinding{
			Asset: VMAsset{
				Hostname: "test-asset-4.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "def-004",
				Title:        "Last Observed Test",
				Severity:     "Low",
			},
			Status:     "open",
			LastFound:  laterTime,
			FirstFound: laterTime,
		}

		err := UpsertFindingInstance(ctx, db, tenantID, asset.ID, definitionUID, vmFinding)
		if err != nil {
			t.Fatalf("First UpsertFindingInstance failed: %v", err)
		}

		// Try to update with earlier time (should not update)
		earlierTime := time.Now().UTC().Add(-1 * time.Hour)
		vmFinding.FirstFound = earlierTime
		vmFinding.LastFound = earlierTime

		err = UpsertFindingInstance(ctx, db, tenantID, asset.ID, definitionUID, vmFinding)
		if err != nil {
			t.Fatalf("Second UpsertFindingInstance failed: %v", err)
		}

		// Verify last_observed_at stayed at later time
		findingRepo := repository.NewFindingInstanceRepository(db)
		instance, err := findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, asset.ID, definitionUID)
		if err != nil {
			t.Fatalf("Failed to get finding instance: %v", err)
		}

		expectedLast := laterTime.Truncate(time.Second)
		actualLast := instance.LastObservedAt.Truncate(time.Second)
		if !expectedLast.Equal(actualLast) {
			t.Errorf("Expected LastObservedAt to remain at %v, got %v", expectedLast, actualLast)
		}
	})

	t.Run("computes effective status from scanner status", func(t *testing.T) {
		tests := []struct {
			name           string
			scannerStatus  string
			expectedStatus string
			expectedReason string
		}{
			{"open stays open", "open", "open", "scanner"},
			{"fixed becomes fixed", "fixed", "fixed", "scanner"},
			{"fixed_by_verification becomes fixed", "fixed_by_verification", "fixed", "scanner"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tenantID := database.CreateTestTenant(t, db)

				asset := createTestFindingAsset(t, db, tenantID, source, "test-asset-status-"+tt.name+".example.com")
				definitionUID := createTestDefinition(t, db, source, "def-status-"+tt.name)

				scannedAt := time.Now().UTC()
				vmFinding := &VMFinding{
					Asset: VMAsset{
						Hostname: "test-asset-status-" + tt.name + ".example.com",
					},
					Finding: VMFindingDetails{
						DefinitionID: "def-status-" + tt.name,
						Title:        "Status Test",
						Severity:     "High",
					},
					Status:     tt.scannerStatus,
					LastFound:  scannedAt,
					FirstFound: scannedAt,
				}

				err := UpsertFindingInstance(ctx, db, tenantID, asset.ID, definitionUID, vmFinding)
				if err != nil {
					t.Fatalf("UpsertFindingInstance failed: %v", err)
				}

				findingRepo := repository.NewFindingInstanceRepository(db)
				instance, err := findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, asset.ID, definitionUID)
				if err != nil {
					t.Fatalf("Failed to get finding instance: %v", err)
				}

				if instance.EffectiveStatus != tt.expectedStatus {
					t.Errorf("Expected effective status '%s', got '%s'", tt.expectedStatus, instance.EffectiveStatus)
				}

				if instance.EffectiveReason != tt.expectedReason {
					t.Errorf("Expected effective reason '%s', got '%s'", tt.expectedReason, instance.EffectiveReason)
				}
			})
		}
	})

	t.Run("uses tenant policy revision for effective revision", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		asset := createTestFindingAsset(t, db, tenantID, source, "test-asset-revision.example.com")
		definitionUID := createTestDefinition(t, db, source, "def-revision")

		scannedAt := time.Now().UTC()
		vmFinding := &VMFinding{
			Asset: VMAsset{
				Hostname: "test-asset-revision.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "def-revision",
				Title:        "Revision Test",
				Severity:     "Medium",
			},
			Status:     "open",
			LastFound:  scannedAt,
			FirstFound: scannedAt,
		}

		err := UpsertFindingInstance(ctx, db, tenantID, asset.ID, definitionUID, vmFinding)
		if err != nil {
			t.Fatalf("UpsertFindingInstance failed: %v", err)
		}

		findingRepo := repository.NewFindingInstanceRepository(db)
		instance, err := findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, asset.ID, definitionUID)
		if err != nil {
			t.Fatalf("Failed to get finding instance: %v", err)
		}

		// Should have policy revision = 1 (initial default)
		if instance.EffectiveRevision != 1 {
			t.Errorf("Expected effective revision 1, got %d", instance.EffectiveRevision)
		}
	})

	t.Run("handles nil evidence gracefully", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		asset := createTestFindingAsset(t, db, tenantID, source, "test-asset-no-evidence.example.com")
		definitionUID := createTestDefinition(t, db, source, "def-no-evidence")

		scannedAt := time.Now().UTC()
		vmFinding := &VMFinding{
			Asset: VMAsset{
				Hostname: "test-asset-no-evidence.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "def-no-evidence",
				Title:        "No Evidence Test",
				Severity:     "Info",
			},
			Status:     "open",
			LastFound:  scannedAt,
			FirstFound: scannedAt,
			Evidence:   nil, // No evidence
		}

		err := UpsertFindingInstance(ctx, db, tenantID, asset.ID, definitionUID, vmFinding)
		if err != nil {
			t.Fatalf("UpsertFindingInstance failed: %v", err)
		}

		findingRepo := repository.NewFindingInstanceRepository(db)
		instance, err := findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, asset.ID, definitionUID)
		if err != nil {
			t.Fatalf("Failed to get finding instance: %v", err)
		}

		// Evidence should be nil or empty map
		if instance.EvidenceJSON != nil && len(instance.EvidenceJSON) > 0 {
			t.Errorf("Expected nil or empty evidence, got %v", instance.EvidenceJSON)
		}
	})
}

// Helper functions

// createTestFindingAsset creates a test asset for finding instance tests
func createTestFindingAsset(t *testing.T, db *sqlx.DB, tenantID int64, source, hostname string) *repository.Asset {
	t.Helper()

	scannedAt := time.Now().UTC()
	vmAsset := &VMAsset{
		Hostname: hostname,
	}

	result, err := UpsertAsset(context.Background(), db, tenantID, source, vmAsset, scannedAt)
	if err != nil {
		t.Fatalf("Failed to create test asset: %v", err)
	}

	return result.Asset
}

// createTestDefinition creates a test finding definition
func createTestDefinition(t *testing.T, db *sqlx.DB, source, definitionID string) string {
	t.Helper()

	vmFinding := &VMFinding{
		Finding: VMFindingDetails{
			DefinitionID: definitionID,
			Title:        "Test Definition " + definitionID,
			Severity:     "High",
		},
		Asset: VMAsset{
			Hostname: "test-" + definitionID + ".example.com",
		},
	}

	err := UpsertDefinitionWithAliases(context.Background(), db, source, vmFinding)
	if err != nil {
		t.Fatalf("Failed to create test definition: %v", err)
	}

	return GenerateDefinitionUID(source, definitionID)
}
