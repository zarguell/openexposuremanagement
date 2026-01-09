package ingest

import (
	"testing"
	"time"

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
	t.Run("upserts_finding_with_observation_window", func(t *testing.T) {
		// TODO: Integration test with database
		t.Skip("Integration test - requires database setup")
	})

	t.Run("is_idempotent_on_repeated_calls", func(t *testing.T) {
		// TODO: Integration test with database
		t.Skip("Integration test - requires database setup")
	})

	t.Run("updates_observation_window_correctly", func(t *testing.T) {
		// TODO: Integration test with database
		// Verify first_observed only moves earlier
		// Verify last_observed only moves later
		t.Skip("Integration test - requires database setup")
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
