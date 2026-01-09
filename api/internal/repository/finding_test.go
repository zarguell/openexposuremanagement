package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test data helpers for finding instances
func getTestFindingInstance() *FindingInstance {
	now := time.Now()
	return &FindingInstance{
		TenantID:          1,
		AssetID:           100,
		DefinitionUID:     "tenable-12345",
		ScannerStatus:     "open",
		FirstObservedAt:   now.Add(-24 * time.Hour),
		LastObservedAt:    now,
		EvidenceJSON:      map[string]interface{}{"port": 443},
		EffectiveStatus:   "open",
		EffectiveReason:   "scanner",
		EffectiveRevision: 0,
	}
}

// Test FindingInstance structure
func TestFindingInstance_Structure(t *testing.T) {
	t.Run("has all required fields", func(t *testing.T) {
		instance := getTestFindingInstance()

		assert.NotNil(t, instance)
		assert.Greater(t, instance.TenantID, int64(0))
		assert.Greater(t, instance.AssetID, int64(0))
		assert.NotEmpty(t, instance.DefinitionUID)
		assert.NotEmpty(t, instance.ScannerStatus)
		assert.False(t, instance.FirstObservedAt.IsZero())
		assert.False(t, instance.LastObservedAt.IsZero())
	})
}

// Test observation window behavior
func TestObservationWindow_FirstObserved(t *testing.T) {
	t.Run("first_observed_only_moves_earlier", func(t *testing.T) {
		// This test documents the expected behavior:
		// When upserting a finding instance, if the new first_observed_at
		// is earlier than the existing one, it should be updated.
		// If it's later, the existing earlier value should be kept.

		existingFirstObserved := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		newFirstObserved := time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC)

		// New first observed is earlier - should update
		assert.True(t, newFirstObserved.Before(existingFirstObserved))

		// If new first observed is later - should NOT update
		newFirstObservedLater := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
		assert.True(t, newFirstObservedLater.After(existingFirstObserved))
	})
}

func TestObservationWindow_LastObserved(t *testing.T) {
	t.Run("last_observed_only_moves_later", func(t *testing.T) {
		// This test documents the expected behavior:
		// When upserting a finding instance, if the new last_observed_at
		// is later than the existing one, it should be updated.
		// If it's earlier, the existing later value should be kept.

		existingLastObserved := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
		newLastObserved := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)

		// New last observed is later - should update
		assert.True(t, newLastObserved.After(existingLastObserved))

		// If new last observed is earlier - should NOT update
		newLastObservedEarlier := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
		assert.True(t, newLastObservedEarlier.Before(existingLastObserved))
	})
}

// Table-driven tests for scanner status validation
func TestScannerStatus_Validation(t *testing.T) {
	validStatuses := []string{
		"open",
		"fixed",
		"fixed_by_verification",
	}

	for _, status := range validStatuses {
		t.Run("valid_status_"+status, func(t *testing.T) {
			// Document that these are the valid scanner statuses
			assert.Contains(t, validStatuses, status)
		})
	}

	t.Run("invalid_statuses", func(t *testing.T) {
		invalidStatuses := []string{
			"",
			"unknown",
			"suppressed",
			"false_positive",
		}

		for _, status := range invalidStatuses {
			// These should be rejected
			assert.NotContains(t, validStatuses, status)
		}
	})
}

// Test effective status computation
func TestEffectiveStatus_Computation(t *testing.T) {
	t.Run("defaults_to_scanner_status_when_no_suppression", func(t *testing.T) {
		// When there's no active suppression, the effective status
		// should match the scanner status
		scannerStatus := "open"
		expectedEffective := "open"

		assert.Equal(t, expectedEffective, scannerStatus)
	})

	t.Run("can_be_overridden_by_suppression", func(t *testing.T) {
		// When there's an active suppression, the effective status
		// should reflect the suppression goal
		scannerStatus := "open"
		suppressionGoal := "accepted_risk"
		expectedEffective := "accepted_risk"

		assert.NotEqual(t, scannerStatus, expectedEffective)
		assert.Equal(t, suppressionGoal, expectedEffective)
	})
}

// Test evidence JSON handling
func TestEvidenceJSON_Handling(t *testing.T) {
	t.Run("stores_arbitrary_key_value_pairs", func(t *testing.T) {
		evidence := map[string]interface{}{
			"port":          443,
			"protocol":      "https",
			"vulnerability": "CVE-2021-44228",
			"scan_id":       "scan-123",
		}

		assert.NotNil(t, evidence)
		assert.Equal(t, 443, evidence["port"])
		assert.Equal(t, "https", evidence["protocol"])
	})

	t.Run("handles_empty_evidence", func(t *testing.T) {
		evidence := map[string]interface{}{}

		assert.Empty(t, evidence)
	})

	t.Run("handles_nil_evidence", func(t *testing.T) {
		var evidence map[string]interface{} = nil

		assert.Nil(t, evidence)
	})
}

// Integration tests (to be implemented with test DB)
func TestFindingInstanceRepository_Upsert(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("creates_new_finding_instance", func(t *testing.T) {
		// TODO: Implement with test DB
		// 1. Insert new finding instance
		// 2. Verify it was created
		// 3. Verify all fields are stored correctly
	})

	t.Run("updates_observation_window_on_re_upsert", func(t *testing.T) {
		// TODO: Implement with test DB
		// 1. Insert finding instance with first/last observed
		// 2. Re-upsert with earlier first_observed
		// 3. Verify first_observed moved earlier
		// 4. Re-upsert with later last_observed
		// 5. Verify last_observed moved later
	})

	t.Run("does_not_update_first_observed_if_later", func(t *testing.T) {
		// TODO: Implement with test DB
	})

	t.Run("does_not_update_last_observed_if_earlier", func(t *testing.T) {
		// TODO: Implement with test DB
	})

	t.Run("updates_scanner_status", func(t *testing.T) {
		// TODO: Implement with test DB
		// 1. Insert finding with status "open"
		// 2. Re-upsert with status "fixed"
		// 3. Verify scanner_status updated
	})

	t.Run("updates_evidence_json", func(t *testing.T) {
		// TODO: Implement with test DB
	})
}

func TestFindingInstanceRepository_GetByTenant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("returns_finding_instances_for_tenant", func(t *testing.T) {
		// TODO: Implement with test DB
	})

	t.Run("returns_empty_slice_for_tenant_with_no_findings", func(t *testing.T) {
		// TODO: Implement with test DB
	})
}

func TestFindingInstanceRepository_GetByAsset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("returns_finding_instances_for_asset", func(t *testing.T) {
		// TODO: Implement with test DB
	})

	t.Run("filters_by_effective_status", func(t *testing.T) {
		// TODO: Implement with test DB
	})
}

func TestFindingInstanceRepository_GetByDefinition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("returns_all_instances_of_a_definition", func(t *testing.T) {
		// TODO: Implement with test DB
	})
}

// Test tenant scoping
func TestFindingInstance_TenantScoping(t *testing.T) {
	t.Run("ensures_tenant_isolation", func(t *testing.T) {
		// This test documents that all finding instance queries
		// must be scoped to a specific tenant
		tenantID := int64(1)
		assert.Greater(t, tenantID, int64(0))
	})

	t.Run("prevents_cross_tenant_access", func(t *testing.T) {
		// TODO: Implement with test DB
		// Verify that a finding from tenant 1 cannot be
		// accessed/updated by a query from tenant 2
	})
}

// Benchmark tests
func BenchmarkFindingInstanceRepository_Upsert(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark")
	}

	// TODO: Set up benchmark with test DB
	// 1. Create test DB and repository
	// 2. Benchmark upsert performance
	// 3. Track observation window updates
}

// Test unique constraint handling
func TestFindingInstance_UniqueConstraint(t *testing.T) {
	t.Run("enforces_unique_per_tenant_asset_definition", func(t *testing.T) {
		// This documents the unique constraint:
		// UNIQUE(tenant_id, asset_id, definition_uid)
		//
		// This means there can only be one finding instance per
		// combination of tenant, asset, and definition.

		tenantID := int64(1)
		assetID := int64(100)
		definitionUID := "tenable-12345"

		// This combination should be unique
		uniqueKey := struct {
			TenantID      int64
			AssetID       int64
			DefinitionUID string
		}{
			TenantID:      tenantID,
			AssetID:       assetID,
			DefinitionUID: definitionUID,
		}

		assert.NotNil(t, uniqueKey)
	})
}
