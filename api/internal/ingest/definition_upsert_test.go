package ingest

import (
	"context"
	"testing"

	"github.com/openexposuremanagement/oem/internal/database"
	"github.com/openexposuremanagement/oem/internal/repository"
)

// TestUpsertDefinitionWithAliases_Integration tests the definition and alias upsert functionality with a real database
func TestUpsertDefinitionWithAliases_Integration(t *testing.T) {
	// Setup test database
	db := database.SetupTestDB(t)
	if db == nil {
		return // Test was skipped
	}

	ctx := context.Background()
	source := "tenable"

	t.Run("creates new definition with CVE aliases", func(t *testing.T) {
		_ = database.CreateTestTenant(t, db) // Tenant not used but needed for DB setup

		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				DefinitionID: "12345",
				Title:        "Apache Log4j Remote Code Execution Vulnerability",
				Severity:     "Critical",
				CVEs:         []string{"CVE-2021-44228", "CVE-2021-45046"},
			},
			Asset: VMAsset{
				Hostname: "test.example.com",
			},
		}

		err := UpsertDefinitionWithAliases(ctx, db, source, vmFinding)
		if err != nil {
			t.Fatalf("UpsertDefinitionWithAliases failed: %v", err)
		}

		// Verify definition was created
		defRepo := repository.NewDefinitionRepository(db)
		expectedUID := GenerateDefinitionUID(source, vmFinding.Finding.DefinitionID)
		def, err := defRepo.GetDefinition(ctx, expectedUID)
		if err != nil {
			t.Fatalf("Failed to get definition: %v", err)
		}

		if def.Source != source {
			t.Errorf("Expected source '%s', got '%s'", source, def.Source)
		}

		if def.SourceDefinitionID != "12345" {
			t.Errorf("Expected source definition ID '12345', got '%s'", def.SourceDefinitionID)
		}

		if def.Title != "Apache Log4j Remote Code Execution Vulnerability" {
			t.Errorf("Expected title 'Apache Log4j Remote Code Execution Vulnerability', got '%s'", def.Title)
		}

		if def.SeverityDefault != "Critical" {
			t.Errorf("Expected severity 'Critical', got '%s'", def.SeverityDefault)
		}

		// Verify CVE aliases were created
		aliases, err := defRepo.GetAliasesForDefinition(ctx, expectedUID)
		if err != nil {
			t.Fatalf("Failed to get aliases: %v", err)
		}

		if len(aliases) != 2 {
			t.Errorf("Expected 2 CVE aliases, got %d", len(aliases))
		}

		aliasMap := make(map[string]string)
		for _, alias := range aliases {
			aliasMap[alias.AliasValue] = alias.AliasType
		}

		if aliasMap["CVE-2021-44228"] != "CVE" {
			t.Error("Expected CVE-2021-44228 alias")
		}
		if aliasMap["CVE-2021-45046"] != "CVE" {
			t.Error("Expected CVE-2021-45046 alias")
		}
	})

	t.Run("updates existing definition metadata", func(t *testing.T) {
		_ = database.CreateTestTenant(t, db)

		// Create initial definition
		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				DefinitionID: "54321",
				Title:        "Original Title",
				Severity:     "High",
				CVEs:         []string{"CVE-2022-1234"},
			},
			Asset: VMAsset{
				Hostname: "test2.example.com",
			},
		}

		err := UpsertDefinitionWithAliases(ctx, db, source, vmFinding)
		if err != nil {
			t.Fatalf("First UpsertDefinitionWithAliases failed: %v", err)
		}

		// Update with new metadata
		updatedFinding := &VMFinding{
			Finding: VMFindingDetails{
				DefinitionID: "54321",
				Title:        "Updated Title",
				Severity:     "Critical",
				CVEs:         []string{"CVE-2022-1234", "CVE-2022-5678"},
			},
			Asset: VMAsset{
				Hostname: "test2.example.com",
			},
		}

		err = UpsertDefinitionWithAliases(ctx, db, source, updatedFinding)
		if err != nil {
			t.Fatalf("Second UpsertDefinitionWithAliases failed: %v", err)
		}

		// Verify definition was updated
		defRepo := repository.NewDefinitionRepository(db)
		expectedUID := GenerateDefinitionUID(source, vmFinding.Finding.DefinitionID)
		def, err := defRepo.GetDefinition(ctx, expectedUID)
		if err != nil {
			t.Fatalf("Failed to get definition: %v", err)
		}

		if def.Title != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got '%s'", def.Title)
		}

		if def.SeverityDefault != "Critical" {
			t.Errorf("Expected severity 'Critical', got '%s'", def.SeverityDefault)
		}

		// Verify new alias was added
		aliases, err := defRepo.GetAliasesForDefinition(ctx, expectedUID)
		if err != nil {
			t.Fatalf("Failed to get aliases: %v", err)
		}

		if len(aliases) != 2 {
			t.Errorf("Expected 2 CVE aliases after update, got %d", len(aliases))
		}

		// Verify updated_at was modified
		if def.UpdatedAt.IsZero() {
			t.Error("Expected updated_at to be set")
		}
	})

	t.Run("handles duplicate CVE aliases idempotently", func(t *testing.T) {
		_ = database.CreateTestTenant(t, db)

		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				DefinitionID: "99999",
				Title:        "Test Finding",
				Severity:     "Medium",
				CVEs:         []string{"CVE-2023-1111"},
			},
			Asset: VMAsset{
				Hostname: "test3.example.com",
			},
		}

		// Upsert same definition twice
		err := UpsertDefinitionWithAliases(ctx, db, source, vmFinding)
		if err != nil {
			t.Fatalf("First UpsertDefinitionWithAliases failed: %v", err)
		}

		err = UpsertDefinitionWithAliases(ctx, db, source, vmFinding)
		if err != nil {
			t.Fatalf("Second UpsertDefinitionWithAliases failed: %v", err)
		}

		// Verify only one alias exists
		defRepo := repository.NewDefinitionRepository(db)
		expectedUID := GenerateDefinitionUID(source, vmFinding.Finding.DefinitionID)
		aliases, err := defRepo.GetAliasesForDefinition(ctx, expectedUID)
		if err != nil {
			t.Fatalf("Failed to get aliases: %v", err)
		}

		if len(aliases) != 1 {
			t.Errorf("Expected 1 CVE alias after duplicate upsert, got %d", len(aliases))
		}
	})

	t.Run("normalizes severity values", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			{"lowercase critical", "critical", "Critical"},
			{"uppercase critical", "CRITICAL", "Critical"},
			{"mixed case high", "HiGh", "High"},
			{"lowercase medium", "medium", "Medium"},
			{"uppercase low", "LOW", "Low"},
			{"mixed case info", "InFo", "Info"},
			{"already normalized", "Critical", "Critical"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_ = database.CreateTestTenant(t, db)

				vmFinding := &VMFinding{
					Finding: VMFindingDetails{
						DefinitionID: "severity-test-" + tt.name,
						Title:        "Severity Test",
						Severity:     tt.input,
					},
					Asset: VMAsset{
						Hostname: "test-severity-" + tt.name + ".example.com",
					},
				}

				err := UpsertDefinitionWithAliases(ctx, db, source, vmFinding)
				if err != nil {
					t.Fatalf("UpsertDefinitionWithAliases failed: %v", err)
				}

				defRepo := repository.NewDefinitionRepository(db)
				expectedUID := GenerateDefinitionUID(source, vmFinding.Finding.DefinitionID)
				def, err := defRepo.GetDefinition(ctx, expectedUID)
				if err != nil {
					t.Fatalf("Failed to get definition: %v", err)
				}

				if def.SeverityDefault != tt.expected {
					t.Errorf("Expected severity '%s', got '%s'", tt.expected, def.SeverityDefault)
				}
			})
		}
	})

	t.Run("generates consistent definition UID", func(t *testing.T) {
		_ = database.CreateTestTenant(t, db)

		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				DefinitionID: "uid-test-123",
				Title:        "UID Test",
				Severity:     "High",
			},
			Asset: VMAsset{
				Hostname: "test-uid.example.com",
			},
		}

		err := UpsertDefinitionWithAliases(ctx, db, source, vmFinding)
		if err != nil {
			t.Fatalf("UpsertDefinitionWithAliases failed: %v", err)
		}

		expectedUID := GenerateDefinitionUID(source, "uid-test-123")
		if expectedUID != "tenable-uid-test-123" {
			t.Errorf("Expected UID 'tenable-uid-test-123', got '%s'", expectedUID)
		}

		// Verify definition exists with expected UID
		defRepo := repository.NewDefinitionRepository(db)
		_, err = defRepo.GetDefinition(ctx, expectedUID)
		if err != nil {
			t.Fatalf("Failed to get definition with UID '%s': %v", expectedUID, err)
		}
	})

	t.Run("extracts CVEs and creates aliases", func(t *testing.T) {
		_ = database.CreateTestTenant(t, db)

		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				DefinitionID: "cve-extract-test",
				Title:        "CVE Extraction Test",
				Severity:     "High",
				CVEs:         []string{"CVE-2020-1234", "CVE-2020-5678", "CVE-2021-9999"},
			},
			Asset: VMAsset{
				Hostname: "test-cve-extract.example.com",
			},
		}

		err := UpsertDefinitionWithAliases(ctx, db, source, vmFinding)
		if err != nil {
			t.Fatalf("UpsertDefinitionWithAliases failed: %v", err)
		}

		defRepo := repository.NewDefinitionRepository(db)
		expectedUID := GenerateDefinitionUID(source, vmFinding.Finding.DefinitionID)
		aliases, err := defRepo.GetAliasesForDefinition(ctx, expectedUID)
		if err != nil {
			t.Fatalf("Failed to get aliases: %v", err)
		}

		if len(aliases) != 3 {
			t.Errorf("Expected 3 CVE aliases, got %d", len(aliases))
		}

		// Verify all CVEs are present
		aliasMap := make(map[string]bool)
		for _, alias := range aliases {
			if alias.AliasType != "CVE" {
				t.Errorf("Expected alias type 'CVE', got '%s'", alias.AliasType)
			}
			aliasMap[alias.AliasValue] = true
		}

		expectedCVEs := []string{"CVE-2020-1234", "CVE-2020-5678", "CVE-2021-9999"}
		for _, cve := range expectedCVEs {
			if !aliasMap[cve] {
				t.Errorf("Expected CVE alias '%s' not found", cve)
			}
		}
	})

	t.Run("handles definition with no CVEs", func(t *testing.T) {
		_ = database.CreateTestTenant(t, db)

		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				DefinitionID: "no-cve-test",
				Title:        "No CVE Test",
				Severity:     "Medium",
				CVEs:         []string{},
			},
			Asset: VMAsset{
				Hostname: "test-no-cve.example.com",
			},
		}

		err := UpsertDefinitionWithAliases(ctx, db, source, vmFinding)
		if err != nil {
			t.Fatalf("UpsertDefinitionWithAliases failed: %v", err)
		}

		defRepo := repository.NewDefinitionRepository(db)
		expectedUID := GenerateDefinitionUID(source, vmFinding.Finding.DefinitionID)
		def, err := defRepo.GetDefinition(ctx, expectedUID)
		if err != nil {
			t.Fatalf("Failed to get definition: %v", err)
		}

		if def.Title != "No CVE Test" {
			t.Errorf("Expected title 'No CVE Test', got '%s'", def.Title)
		}

		aliases, err := defRepo.GetAliasesForDefinition(ctx, expectedUID)
		if err != nil {
			t.Fatalf("Failed to get aliases: %v", err)
		}

		if len(aliases) != 0 {
			t.Errorf("Expected 0 aliases for definition with no CVEs, got %d", len(aliases))
		}
	})

	t.Run("stores references as JSONB", func(t *testing.T) {
		_ = database.CreateTestTenant(t, db)

		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				DefinitionID: "refs-test",
				Title:        "References Test",
				Severity:     "High",
				CVEs:         []string{"CVE-2023-7777"},
				References:   []string{"https://example.com/advisory", "https://example.com/patch"},
			},
			Asset: VMAsset{
				Hostname: "test-refs.example.com",
			},
		}

		err := UpsertDefinitionWithAliases(ctx, db, source, vmFinding)
		if err != nil {
			t.Fatalf("UpsertDefinitionWithAliases failed: %v", err)
		}

		defRepo := repository.NewDefinitionRepository(db)
		expectedUID := GenerateDefinitionUID(source, vmFinding.Finding.DefinitionID)
		def, err := defRepo.GetDefinition(ctx, expectedUID)
		if err != nil {
			t.Fatalf("Failed to get definition: %v", err)
		}

		if def.ReferencesJSON == nil {
			t.Error("Expected references_json to be set")
		}

		// Verify we have the expected number of references
		if len(def.ReferencesJSON) < 2 {
			t.Errorf("Expected at least 2 references, got %d", len(def.ReferencesJSON))
		}
	})
}
