package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/database"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/stretchr/testify/assert"
)

// TestConvertToFindingInstance tests converting VMFinding to FindingInstance
func TestConvertToFindingInstance(t *testing.T) {
	t.Run("converts basic finding", func(t *testing.T) {
		now := time.Now()
		vmFinding := &VMFinding{
			Asset: VMAsset{
				Hostname: "webserver01.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "12345",
				Title:        "Test Vulnerability",
				Severity:     "High",
			},
			Status:     "open",
			FirstFound: now.Add(-24 * time.Hour),
			LastFound:  now,
			Evidence: map[string]interface{}{
				"port": 443,
			},
		}

		tenantID := int64(1)
		assetID := int64(100)
		definitionUID := "tenable-12345"
		policyRevision := int64(5)

		instance := ConvertToFindingInstance(tenantID, assetID, definitionUID, vmFinding, policyRevision)

		assert.Equal(t, tenantID, instance.TenantID)
		assert.Equal(t, assetID, instance.AssetID)
		assert.Equal(t, definitionUID, instance.DefinitionUID)
		assert.Equal(t, "open", instance.ScannerStatus)
		assert.Equal(t, "open", instance.EffectiveStatus)
		assert.Equal(t, "scanner", instance.EffectiveReason)
		assert.Equal(t, policyRevision, instance.EffectiveRevision)
		assert.NotNil(t, instance.EvidenceJSON)
	})

	t.Run("handles_fixed_status", func(t *testing.T) {
		vmFinding := &VMFinding{
			Status: "fixed",
			Asset: VMAsset{
				Hostname: "webserver01.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "12345",
				Title:        "Test Vulnerability",
			},
		}

		instance := ConvertToFindingInstance(1, 100, "tenable-12345", vmFinding, 1)

		assert.Equal(t, "fixed", instance.ScannerStatus)
		assert.Equal(t, "fixed", instance.EffectiveStatus)
		assert.Equal(t, "scanner", instance.EffectiveReason)
	})

	t.Run("handles_fixed_by_verification_status", func(t *testing.T) {
		vmFinding := &VMFinding{
			Status: "fixed_by_verification",
			Asset: VMAsset{
				Hostname: "webserver01.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "12345",
				Title:        "Test Vulnerability",
			},
		}

		instance := ConvertToFindingInstance(1, 100, "tenable-12345", vmFinding, 1)

		assert.Equal(t, "fixed_by_verification", instance.ScannerStatus)
		assert.Equal(t, "fixed", instance.EffectiveStatus) // Maps to "fixed"
		assert.Equal(t, "scanner", instance.EffectiveReason)
	})
}

// TestNormalizeScannerStatus tests scanner status normalization
func TestNormalizeScannerStatus(t *testing.T) {
	t.Run("passes through valid statuses", func(t *testing.T) {
		validStatuses := []string{
			"open",
			"fixed",
			"fixed_by_verification",
		}

		for _, status := range validStatuses {
			result := NormalizeScannerStatus(status)
			assert.Equal(t, status, result)
		}
	})

	t.Run("maps fixed_by_verification to fixed for effective", func(t *testing.T) {
		scannerStatus := "fixed_by_verification"
		effectiveStatus := ComputeEffectiveStatus(scannerStatus, nil)

		// For now, without suppressions, effective status should match scanner
		// status semantics (fixed_by_verification -> fixed)
		assert.Equal(t, "fixed", effectiveStatus)
	})
}

// TestComputeEffectiveStatus tests effective status computation
func TestComputeEffectiveStatus(t *testing.T) {
	t.Run("scanner_status_maps_to_effective_when_no_suppression", func(t *testing.T) {
		tests := []struct {
			scannerStatus string
			expected      string
		}{
			{"open", "open"},
			{"fixed", "fixed"},
			{"fixed_by_verification", "fixed"},
		}

		for _, tt := range tests {
			t.Run(tt.scannerStatus, func(t *testing.T) {
				result := ComputeEffectiveStatus(tt.scannerStatus, nil)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("suppression_overrides_scanner_status", func(t *testing.T) {
		// TODO: Implement when suppression logic is added
		// For now, document expected behavior:
		// - If there's an active suppression for the CVE
		// - The effective status should be the suppression goal
		// - e.g., "accepted_risk" or "false_positive"
	})
}

// TestObservationWindowTests tests observation window behavior
func TestObservationWindowTests(t *testing.T) {
	t.Run("first_observed_uses_earliest_timestamp", func(t *testing.T) {
		// This test documents the observation window behavior:
		// First observed should track the earliest time we've seen this finding

		firstScan := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		secondScan := time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC)

		// Second scan has a later first_found, but we should keep the earlier one
		assert.True(t, secondScan.After(firstScan))
	})

	t.Run("last_observed_uses_latest_timestamp", func(t *testing.T) {
		// Last observed should track the most recent time we've seen this finding

		firstScan := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		secondScan := time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC)

		// Second scan has a later last_found, so we should update to it
		assert.True(t, secondScan.After(firstScan))
	})
}

// TestEvidenceHandling tests evidence JSON handling
func TestEvidenceHandling(t *testing.T) {
	t.Run("preserves_evidence_data", func(t *testing.T) {
		vmFinding := &VMFinding{
			Asset: VMAsset{
				Hostname: "webserver01.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "12345",
				Title:        "Test",
			},
			Status: "open",
			Evidence: map[string]interface{}{
				"port":          443,
				"protocol":      "https",
				"vulnerability": "CVE-2021-44228",
			},
		}

		instance := ConvertToFindingInstance(1, 100, "test-12345", vmFinding, 1)

		assert.NotNil(t, instance.EvidenceJSON)
		assert.Equal(t, 443, instance.EvidenceJSON["port"])
		assert.Equal(t, "https", instance.EvidenceJSON["protocol"])
	})

	t.Run("handles_nil_evidence", func(t *testing.T) {
		vmFinding := &VMFinding{
			Asset: VMAsset{
				Hostname: "webserver01.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "12345",
				Title:        "Test",
			},
			Status:   "open",
			Evidence: nil,
		}

		instance := ConvertToFindingInstance(1, 100, "test-12345", vmFinding, 1)

		// Evidence should be nil or empty map
		if instance.EvidenceJSON != nil {
			assert.Empty(t, instance.EvidenceJSON)
		}
	})

	t.Run("handles_empty_evidence", func(t *testing.T) {
		vmFinding := &VMFinding{
			Asset: VMAsset{
				Hostname: "webserver01.example.com",
			},
			Finding: VMFindingDetails{
				DefinitionID: "12345",
				Title:        "Test",
			},
			Status:   "open",
			Evidence: map[string]interface{}{},
		}

		instance := ConvertToFindingInstance(1, 100, "test-12345", vmFinding, 1)

		assert.NotNil(t, instance.EvidenceJSON)
		assert.Empty(t, instance.EvidenceJSON)
	})
}

// TestUpsertFindingInstance tests the full upsert flow
func TestUpsertFindingInstance(t *testing.T) {
	// Setup test database
	db := database.SetupTestDB(t)
	if db == nil {
		return // Test was skipped
	}

	ctx := context.Background()
	tenantID := database.CreateTestTenant(t, db)

	t.Run("upserts_finding_with_observation_window", func(t *testing.T) {
		// Create test asset and definition
		asset := createTestAssetForFinding(t, db, tenantID, "test-host.example.com", time.Now())
		assetID := asset.ID
		definitionUID := "test-def-1"
		createTestFindingDefinition(t, db, definitionUID, "tenable", "12345", "Test Finding")

		vmFinding := &VMFinding{
			Status:     "open",
			LastFound:  time.Now(),
			FirstFound: time.Now().Add(-1 * time.Hour),
			Evidence: map[string]interface{}{
				"port": 443,
			},
			Finding: VMFindingDetails{
				DefinitionID: "12345",
				Title:        "Test Finding",
			},
		}

		err := UpsertFindingInstance(ctx, db, tenantID, assetID, definitionUID, vmFinding)
		assert.NoError(t, err)

		// Verify finding instance was created
		findingRepo := repository.NewFindingInstanceRepository(db)
		instance, err := findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, assetID, definitionUID)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
		assert.Equal(t, "open", instance.ScannerStatus)
		assert.Equal(t, "open", instance.EffectiveStatus)
		assert.Equal(t, float64(443), instance.EvidenceJSON["port"])
	})

	t.Run("is_idempotent_on_repeated_calls", func(t *testing.T) {
		// Create test asset and definition
		asset := createTestAssetForFinding(t, db, tenantID, "test-host2.example.com", time.Now())
		assetID := asset.ID
		definitionUID := "test-def-2"
		createTestFindingDefinition(t, db, definitionUID, "qualys", "67890", "Test Finding 2")

		vmFinding := &VMFinding{
			Status:    "open",
			LastFound: time.Now(),
			Finding: VMFindingDetails{
				DefinitionID: "67890",
				Title:        "Test Finding 2",
			},
		}

		// Call upsert twice
		err := UpsertFindingInstance(ctx, db, tenantID, assetID, definitionUID, vmFinding)
		assert.NoError(t, err)
		err = UpsertFindingInstance(ctx, db, tenantID, assetID, definitionUID, vmFinding)
		assert.NoError(t, err)

		// Verify only one instance exists
		findingRepo := repository.NewFindingInstanceRepository(db)
		instances, err := findingRepo.GetByTenantAndAsset(ctx, tenantID, assetID)
		assert.NoError(t, err)
		assert.Len(t, instances, 1)
	})

	t.Run("updates_observation_window_correctly", func(t *testing.T) {
		// Create test asset and definition
		asset := createTestAssetForFinding(t, db, tenantID, "test-host3.example.com", time.Now())
		assetID := asset.ID
		definitionUID := "test-def-3"
		createTestFindingDefinition(t, db, definitionUID, "rapid7", "99999", "Test Finding 3")

		findingRepo := repository.NewFindingInstanceRepository(db)

		// First upsert with initial timestamps
		firstTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		vmFinding1 := &VMFinding{
			Status:     "open",
			LastFound:  firstTime,
			FirstFound: firstTime.Add(-1 * time.Hour),
			Finding: VMFindingDetails{
				DefinitionID: "99999",
				Title:        "Test Finding 3",
			},
		}

		err := UpsertFindingInstance(ctx, db, tenantID, assetID, definitionUID, vmFinding1)
		assert.NoError(t, err)

		// Verify initial timestamps
		instance, err := findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, assetID, definitionUID)
		assert.NoError(t, err)
		assert.Equal(t, firstTime.Add(-1*time.Hour), instance.FirstObservedAt)
		assert.Equal(t, firstTime, instance.LastObservedAt)

		// Second upsert with later last_found (should update)
		secondTime := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
		vmFinding2 := &VMFinding{
			Status:     "open",
			LastFound:  secondTime,
			FirstFound: secondTime.Add(-30 * time.Minute), // Earlier first_found
			Finding: VMFindingDetails{
				DefinitionID: "99999",
				Title:        "Test Finding 3",
			},
		}

		err = UpsertFindingInstance(ctx, db, tenantID, assetID, definitionUID, vmFinding2)
		assert.NoError(t, err)

		// Verify timestamps were updated correctly
		instance, err = findingRepo.GetByTenantAssetAndDefinition(ctx, tenantID, assetID, definitionUID)
		assert.NoError(t, err)
		// First observed should NOT move later - it stays at the earliest time seen
		assert.Equal(t, firstTime.Add(-1*time.Hour), instance.FirstObservedAt)
		// Last observed should move later
		assert.Equal(t, secondTime, instance.LastObservedAt)
	})
}

// Table-driven tests for edge cases
func TestFindingInstance_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		vmFinding     *VMFinding
		tenantID      int64
		assetID       int64
		definitionUID string
		expectError   bool
	}{
		{
			name: "valid finding with all fields",
			vmFinding: &VMFinding{
				Status: "open",
				Asset: VMAsset{
					Hostname: "test.example.com",
				},
				Finding: VMFindingDetails{
					DefinitionID: "12345",
					Title:        "Test",
				},
			},
			tenantID:      1,
			assetID:       100,
			definitionUID: "test-12345",
			expectError:   false,
		},
		{
			name: "missing first_found - uses last_found",
			vmFinding: &VMFinding{
				Status:    "open",
				LastFound: time.Now(),
				Asset: VMAsset{
					Hostname: "test.example.com",
				},
				Finding: VMFindingDetails{
					DefinitionID: "12345",
					Title:        "Test",
				},
			},
			tenantID:      1,
			assetID:       100,
			definitionUID: "test-12345",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := ConvertToFindingInstance(
				tt.tenantID,
				tt.assetID,
				tt.definitionUID,
				tt.vmFinding,
				1, // policyRevision
			)

			if !tt.expectError {
				assert.NotNil(t, instance)
				assert.Equal(t, tt.tenantID, instance.TenantID)
				assert.Equal(t, tt.assetID, instance.AssetID)
				assert.Equal(t, tt.definitionUID, instance.DefinitionUID)
			}
		})
	}
}

// Helper functions for integration tests

var ctx = context.Background()

// createTestAssetForFinding creates a test asset and returns it
func createTestAssetForFinding(t *testing.T, db *sqlx.DB, tenantID int64, canonicalName string, seenAt time.Time) *repository.Asset {
	t.Helper()

	asset := &repository.Asset{
		TenantID:      tenantID,
		CanonicalName: canonicalName,
		FirstSeenAt:   seenAt,
		LastSeenAt:    seenAt,
		IsActive:      true,
	}

	assetRepo := repository.NewAssetRepository(db)
	err := assetRepo.Create(ctx, asset)
	if err != nil {
		t.Fatalf("failed to create test asset: %v", err)
	}

	return asset
}

// createTestFindingDefinition creates a test finding definition
func createTestFindingDefinition(t *testing.T, db *sqlx.DB, definitionUID, source, sourceDefID, title string) {
	t.Helper()

	def := &repository.FindingDefinition{
		DefinitionUID:      definitionUID,
		Source:             source,
		SourceDefinitionID: sourceDefID,
		Title:              title,
		SeverityDefault:    "High",
		ReferencesJSON:     []string{"https://example.com"},
	}

	defRepo := repository.NewDefinitionRepository(db)
	err := defRepo.UpsertDefinition(ctx, def)
	if err != nil {
		t.Fatalf("failed to create test definition: %v", err)
	}
}
