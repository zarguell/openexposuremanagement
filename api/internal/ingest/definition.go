package ingest

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/repository"
	"github.com/rs/zerolog/log"
)

// GenerateDefinitionUID generates a unique identifier for a definition
// Format: {source}-{source_definition_id}
func GenerateDefinitionUID(source, sourceDefID string) string {
	return source + "-" + sourceDefID
}

// ConvertToDefinition converts a VMFinding to a FindingDefinition
func ConvertToDefinition(source string, vmFinding *VMFinding) *repository.FindingDefinition {
	uid := GenerateDefinitionUID(source, vmFinding.Finding.DefinitionID)

	return &repository.FindingDefinition{
		DefinitionUID:      uid,
		Source:             source,
		SourceDefinitionID: vmFinding.Finding.DefinitionID,
		Title:              vmFinding.Finding.Title,
		SeverityDefault:    NormalizeSeverity(vmFinding.Finding.Severity),
		ReferencesJSON:     vmFinding.Finding.References,
	}
}

// NormalizeSeverity normalizes a severity value to a standard format
func NormalizeSeverity(severity string) string {
	// Convert to title case
	severity = strings.TrimSpace(severity)
	severity = strings.ToLower(severity)

	// Map to standard values
	switch severity {
	case "critical":
		return "Critical"
	case "high":
		return "High"
	case "medium":
		return "Medium"
	case "low":
		return "Low"
	case "info", "informational", "information":
		return "Info"
	default:
		// Unknown severities default to Info
		return "Info"
	}
}

// ExtractCVEAliases extracts CVE aliases from a VM finding
func ExtractCVEAliases(definitionUID string, vmFinding *VMFinding) []repository.FindingDefinitionAlias {
	if len(vmFinding.Finding.CVEs) == 0 {
		return []repository.FindingDefinitionAlias{}
	}

	aliases := make([]repository.FindingDefinitionAlias, 0, len(vmFinding.Finding.CVEs))
	for _, cve := range vmFinding.Finding.CVEs {
		cve = strings.TrimSpace(cve)
		if cve == "" {
			continue
		}
		aliases = append(aliases, repository.FindingDefinitionAlias{
			DefinitionUID: definitionUID,
			AliasType:     "CVE",
			AliasValue:    cve,
		})
	}

	return aliases
}

// UpsertDefinitionWithAliases upserts a definition and its CVE aliases
func UpsertDefinitionWithAliases(ctx context.Context, db *sqlx.DB, source string, vmFinding *VMFinding) error {
	// Convert to definition
	def := ConvertToDefinition(source, vmFinding)

	// Upsert definition
	repo := repository.NewDefinitionRepository(db)
	err := repo.UpsertDefinition(ctx, def)
	if err != nil {
		log.Error().
			Str("definition_uid", def.DefinitionUID).
			Err(err).
			Msg("Failed to upsert definition")
		return err
	}

	log.Debug().
		Str("definition_uid", def.DefinitionUID).
		Str("title", def.Title).
		Msg("Upserted definition")

	// Extract and upsert CVE aliases
	aliases := ExtractCVEAliases(def.DefinitionUID, vmFinding)
	for _, alias := range aliases {
		err := repo.UpsertAlias(ctx, &alias)
		if err != nil {
			log.Error().
				Str("definition_uid", def.DefinitionUID).
				Str("alias_type", alias.AliasType).
				Str("alias_value", alias.AliasValue).
				Err(err).
				Msg("Failed to upsert alias")
			return err
		}
	}

	if len(aliases) > 0 {
		log.Debug().
			Str("definition_uid", def.DefinitionUID).
			Int("count", len(aliases)).
			Msg("Upserted CVE aliases")
	}

	return nil
}
