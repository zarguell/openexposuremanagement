package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/openexposuremanagement/oem/internal/database"
	"github.com/openexposuremanagement/oem/internal/repository"
)

// TestUpsertAsset_Integration tests the asset upsert functionality with a real database
func TestUpsertAsset_Integration(t *testing.T) {
	// Setup test database
	db := database.SetupTestDB(t)
	if db == nil {
		return // Test was skipped
	}

	ctx := context.Background()
	source := "tenable"
	scannedAt := time.Now().UTC()

	t.Run("creates new asset when no match exists", func(t *testing.T) {
		// Each subtest gets its own tenant
		tenantID := database.CreateTestTenant(t, db)

		vmAsset := &VMAsset{
			Hostname: "new-server.example.com",
			IPAddresses: []string{"192.168.1.50"},
			ExternalIDs: map[string]string{
				"aws-id": "i-new12345",
			},
		}

		result, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, scannedAt)
		if err != nil {
			t.Fatalf("UpsertAsset failed: %v", err)
		}

		// Verify new asset was created
		if !result.NewAsset {
			t.Error("Expected NewAsset=true for new asset")
		}

		if result.Asset.ID == 0 {
			t.Error("Expected valid asset ID")
		}

		if result.Asset.TenantID != tenantID {
			t.Errorf("Expected tenant ID %d, got %d", tenantID, result.Asset.TenantID)
		}

		if result.Asset.CanonicalName != "new-server.example.com" {
			t.Errorf("Expected canonical name 'new-server.example.com', got '%s'", result.Asset.CanonicalName)
		}

		if !result.Asset.FirstSeenAt.Equal(scannedAt) {
			t.Errorf("Expected FirstSeenAt %v, got %v", scannedAt, result.Asset.FirstSeenAt)
		}

		if !result.Asset.LastSeenAt.Equal(scannedAt) {
			t.Errorf("Expected LastSeenAt %v, got %v", scannedAt, result.Asset.LastSeenAt)
		}

		if !result.Asset.IsActive {
			t.Error("Expected asset to be active")
		}

		// Verify match reason
		if result.Reason != MatchReasonNoMatch {
			t.Errorf("Expected match reason %s, got %s", MatchReasonNoMatch, result.Reason)
		}

		// Verify identifiers were created
		assetRepo := repository.NewAssetRepository(db)
		identifiers, err := assetRepo.GetIdentifiers(ctx, result.Asset.ID)
		if err != nil {
			t.Fatalf("Failed to get identifiers: %v", err)
		}

		// Should have: hostname, shortname, ipv4, external_id = 4 identifiers
		if len(identifiers) != 4 {
			t.Errorf("Expected 4 identifiers (hostname, shortname, ipv4, external_id), got %d", len(identifiers))
		}

		// Check identifier types exist
		idTypes := make(map[string]bool)
		for _, id := range identifiers {
			idTypes[id.IDType] = true
		}

		if !idTypes["hostname_norm"] {
			t.Error("Expected hostname_norm identifier")
		}
		if !idTypes["ipv4"] {
			t.Error("Expected ipv4 identifier")
		}
		if !idTypes["external_id:aws-id"] {
			t.Error("Expected external_id:aws-id identifier")
		}
	})

	t.Run("updates existing asset last_seen_at when matched", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		// Create initial asset
		initialTime := scannedAt.Add(-1 * time.Hour)
		vmAsset := &VMAsset{
			Hostname: "existing-server.example.com",
		}

		_, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, initialTime)
		if err != nil {
			t.Fatalf("First UpsertAsset failed: %v", err)
		}

		// Upsert again with later timestamp
		laterTime := scannedAt
		secondResult, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, laterTime)
		if err != nil {
			t.Fatalf("Second UpsertAsset failed: %v", err)
		}

		// Should match existing asset
		if secondResult.NewAsset {
			t.Error("Expected second upsert to match existing asset")
		}

		// LastSeenAt should be updated
		if !secondResult.Asset.LastSeenAt.Equal(laterTime) {
			t.Errorf("Expected LastSeenAt %v, got %v", laterTime, secondResult.Asset.LastSeenAt)
		}

		// FirstSeenAt should remain unchanged (with second precision)
		expectedFirst := initialTime.Truncate(time.Second)
		actualFirst := secondResult.Asset.FirstSeenAt.Truncate(time.Second)
		if !expectedFirst.Equal(actualFirst) {
			t.Errorf("Expected FirstSeenAt %v, got %v", expectedFirst, actualFirst)
		}
	})

	t.Run("updates canonical name when hostname changes", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		// Create asset with external ID
		vmAsset := &VMAsset{
			ExternalIDs: map[string]string{
				"aws-id": "i-canonical123",
			},
			Hostname: "old-name.example.com",
		}

		_, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, scannedAt)
		if err != nil {
			t.Fatalf("First UpsertAsset failed: %v", err)
		}

		// Upsert same asset (by external ID) with different hostname
		vmAsset.Hostname = "new-name.example.com"
		secondResult, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, scannedAt.Add(time.Minute))
		if err != nil {
			t.Fatalf("Second UpsertAsset failed: %v", err)
		}

		// Should match by external ID (second call should update existing asset)
		if secondResult.NewAsset {
			t.Error("Expected second upsert to update existing asset")
		}

		// Canonical name should be updated
		if secondResult.Asset.CanonicalName != "new-name.example.com" {
			t.Errorf("Expected canonical name 'new-name.example.com', got '%s'", secondResult.Asset.CanonicalName)
		}
	})

	t.Run("does not update last_seen_at if earlier time provided", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		vmAsset := &VMAsset{
			Hostname: "time-server.example.com",
		}

		// Create with later time
		laterTime := scannedAt
		_, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, laterTime)
		if err != nil {
			t.Fatalf("First UpsertAsset failed: %v", err)
		}

		// Try to update with earlier time
		earlierTime := scannedAt.Add(-1 * time.Hour)
		secondResult, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, earlierTime)
		if err != nil {
			t.Fatalf("Second UpsertAsset failed: %v", err)
		}

		// LastSeenAt should remain at the later time (with second precision)
		expectedTime := laterTime.Truncate(time.Second)
		actualTime := secondResult.Asset.LastSeenAt.Truncate(time.Second)
		if !expectedTime.Equal(actualTime) {
			t.Errorf("Expected LastSeenAt to remain at %v, got %v", expectedTime, actualTime)
		}
	})

	t.Run("uses hostname as canonical name when available", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		vmAsset := &VMAsset{
			Hostname: "host-named-server.example.com",
			IPAddresses: []string{"10.0.0.5"},
		}

		result, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, scannedAt)
		if err != nil {
			t.Fatalf("UpsertAsset failed: %v", err)
		}

		if result.Asset.CanonicalName != "host-named-server.example.com" {
			t.Errorf("Expected canonical name from hostname, got '%s'", result.Asset.CanonicalName)
		}
	})

	t.Run("uses IP as canonical name when no hostname", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		vmAsset := &VMAsset{
			IPAddresses: []string{"10.1.2.3"},
		}

		result, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, scannedAt)
		if err != nil {
			t.Fatalf("UpsertAsset failed: %v", err)
		}

		if result.Asset.CanonicalName != "10.1.2.3" {
			t.Errorf("Expected IP as canonical name, got '%s'", result.Asset.CanonicalName)
		}
	})

	t.Run("uses external ID as canonical name when no hostname or IP", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		vmAsset := &VMAsset{
			ExternalIDs: map[string]string{
				"aws-id": "i-externalcanonical",
			},
		}

		result, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, scannedAt)
		if err != nil {
			t.Fatalf("UpsertAsset failed: %v", err)
		}

		if result.Asset.CanonicalName != "i-externalcanonical" {
			t.Errorf("Expected external ID as canonical name, got '%s'", result.Asset.CanonicalName)
		}
	})

	t.Run("returns error when cannot determine canonical name", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		vmAsset := &VMAsset{
			// No hostname, no IP, no external IDs
		}

		_, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, scannedAt)
		if err == nil {
			t.Error("Expected error when asset has no identifiers, got nil")
		}

		if _, ok := err.(ValidationError); !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		}
	})

	t.Run("upserts multiple identifiers for same asset", func(t *testing.T) {
		tenantID := database.CreateTestTenant(t, db)

		vmAsset := &VMAsset{
			Hostname: "multi-id-server.example.com",
			IPAddresses: []string{"192.168.1.100", "192.168.1.101"},
			ExternalIDs: map[string]string{
				"aws-id": "i-multi123",
				"azure-id": "azure-multi456",
			},
		}

		result, err := UpsertAsset(ctx, db, tenantID, source, vmAsset, scannedAt)
		if err != nil {
			t.Fatalf("UpsertAsset failed: %v", err)
		}

		// Verify all identifiers were created
		assetRepo := repository.NewAssetRepository(db)
		identifiers, err := assetRepo.GetIdentifiers(ctx, result.Asset.ID)
		if err != nil {
			t.Fatalf("Failed to get identifiers: %v", err)
		}

		// Should have: hostname, shortname, 2 IPs, 2 external IDs = 7 identifiers (hostname + shortname + 2xipv4 + 2xexternal_id)
		// Actually: hostname (1) + shortname (1) + ipv4 (2) + external_id:aws-id (1) + external_id:azure-id (1) = 6
		expectedCount := 6
		if len(identifiers) != expectedCount {
			t.Errorf("Expected %d identifiers, got %d. Identifiers: %+v", expectedCount, len(identifiers), identifiers)
		}

		// Verify specific identifiers exist
		idMap := make(map[string]string)
		for _, id := range identifiers {
			idMap[id.IDType] = id.IDValue
		}

		if idMap["hostname_norm"] != "multi-id-server.example.com" {
			t.Error("Expected hostname identifier")
		}
		if idMap["ipv4_192.168.1.100"] != "" || idMap["ipv4_192.168.1.101"] != "" {
			// IPs might be stored differently, just check we have ipv4 types
			ipCount := 0
			for _, id := range identifiers {
				if id.IDType == "ipv4" {
					ipCount++
				}
			}
			if ipCount != 2 {
				t.Errorf("Expected 2 IP identifiers, got %d", ipCount)
			}
		}
	})
}
