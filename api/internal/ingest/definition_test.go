package ingest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUpsertDefinition tests the definition upsert logic
func TestUpsertDefinition(t *testing.T) {
	t.Run("generates definition UID from source and source ID", func(t *testing.T) {
		source := "tenable"
		sourceDefID := "12345"

		uid := GenerateDefinitionUID(source, sourceDefID)

		expected := "tenable-12345"
		assert.Equal(t, expected, uid)
	})

	t.Run("handles different sources", func(t *testing.T) {
		tests := []struct {
			name       string
			source     string
			sourceDefID string
			expected   string
		}{
			{"tenable", "tenable", "12345", "tenable-12345"},
			{"qualys", "qualys", "67890", "qualys-67890"},
			{"rapid7", "rapid7", "abc-123", "rapid7-abc-123"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uid := GenerateDefinitionUID(tt.source, tt.sourceDefID)
				assert.Equal(t, tt.expected, uid)
			})
		}
	})
}

// TestConvertToDefinition tests converting VMFindingDetails to FindingDefinition
func TestConvertToDefinition(t *testing.T) {
	t.Run("converts basic finding details", func(t *testing.T) {
		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				DefinitionID: "12345",
				Title:        "Test Vulnerability",
				Severity:     "High",
				References:   []string{"https://example.com/1"},
				Description:  "A test vulnerability",
				Solution:     "Apply patch",
			},
		}

		def := ConvertToDefinition("tenable", vmFinding)

		assert.Equal(t, "tenable-12345", def.DefinitionUID)
		assert.Equal(t, "tenable", def.Source)
		assert.Equal(t, "12345", def.SourceDefinitionID)
		assert.Equal(t, "Test Vulnerability", def.Title)
		assert.Equal(t, "High", def.SeverityDefault)
		assert.Equal(t, []string{"https://example.com/1"}, def.ReferencesJSON)
	})

	t.Run("handles empty references", func(t *testing.T) {
		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				DefinitionID: "12345",
				Title:        "Test Vulnerability",
				Severity:     "Medium",
			},
		}

		def := ConvertToDefinition("qualys", vmFinding)

		assert.Nil(t, def.ReferencesJSON)
	})
}

// TestExtractCVEAliases tests CVE alias extraction
func TestExtractCVEAliases(t *testing.T) {
	t.Run("extracts CVE IDs from finding", func(t *testing.T) {
		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				CVEs: []string{"CVE-2021-44228", "CVE-2021-45046"},
			},
		}

		aliases := ExtractCVEAliases("test-def-1", vmFinding)

		assert.Len(t, aliases, 2)
		assert.Equal(t, "CVE", aliases[0].AliasType)
		assert.Equal(t, "CVE-2021-44228", aliases[0].AliasValue)
		assert.Equal(t, "CVE-2021-45046", aliases[1].AliasValue)
	})

	t.Run("returns empty slice when no CVEs", func(t *testing.T) {
		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				CVEs: []string{},
			},
		}

		aliases := ExtractCVEAliases("test-def-1", vmFinding)

		assert.Empty(t, aliases)
	})

	t.Run("handles nil CVEs slice", func(t *testing.T) {
		vmFinding := &VMFinding{
			Finding: VMFindingDetails{
				CVEs: nil,
			},
		}

		aliases := ExtractCVEAliases("test-def-1", vmFinding)

		assert.Empty(t, aliases)
	})
}

// TestUpsertDefinitionWithAliases tests the full upsert flow
func TestUpsertDefinitionWithAliases(t *testing.T) {
	t.Run("upserts definition and aliases", func(t *testing.T) {
		// TODO: Implement integration test with real database
		t.Skip("Integration test - requires database setup")
	})

	t.Run("is idempotent on repeated calls", func(t *testing.T) {
		// TODO: Test that calling upsert twice with same data
		// doesn't create duplicates
		t.Skip("Integration test - requires database setup")
	})
}

// TestNormalizeSeverity tests severity normalization
func TestNormalizeSeverity(t *testing.T) {
	t.Run("passes through valid severities", func(t *testing.T) {
		validSeverities := []string{
			"Critical", "High", "Medium", "Low", "Info",
		}

		for _, severity := range validSeverities {
			result := NormalizeSeverity(severity)
			assert.Equal(t, severity, result)
		}
	})

	t.Run("handles case variations", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"critical", "Critical"},
			{"CRITICAL", "Critical"},
			{"high", "High"},
			{"HIGH", "High"},
			{"medium", "Medium"},
			{"low", "Low"},
			{"info", "Info"},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				result := NormalizeSeverity(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("handles unknown severities", func(t *testing.T) {
		// Unknown severities should be normalized to "Info" or "Unknown"
		result := NormalizeSeverity("unknown-severity")
		assert.Equal(t, "Info", result)
	})
}

// Table-driven tests for edge cases
func TestDefinitionUpsert_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		vmFinding   *VMFinding
		source      string
		expectError bool
	}{
		{
			name: "valid finding with all fields",
			vmFinding: &VMFinding{
				Finding: VMFindingDetails{
					DefinitionID: "12345",
					Title:        "Test",
					Severity:     "High",
				},
			},
			source:      "tenable",
			expectError: false,
		},
		{
			name: "missing definition ID",
			vmFinding: &VMFinding{
				Finding: VMFindingDetails{
					Title:    "Test",
					Severity: "High",
				},
			},
			source:      "tenable",
			expectError: true,
		},
		{
			name: "missing title",
			vmFinding: &VMFinding{
				Finding: VMFindingDetails{
					DefinitionID: "12345",
					Severity:     "High",
				},
			},
			source:      "tenable",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This documents expected behavior
			// Actual implementation will validate these constraints
			if tt.expectError {
				// Should return error when trying to convert
			} else {
				def := ConvertToDefinition(tt.source, tt.vmFinding)
				assert.NotNil(t, def)
			}
		})
	}
}
