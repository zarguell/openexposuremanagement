package query

import (
	"errors"
	"fmt"
)

// allowedFields defines whitelisted fields per entity
// These fields must match the columns in the database views:
// - findings: uses the "findings" view which joins finding_instances, finding_definitions, assets, and intel_cve
// - assets: uses the "assets_extended" view which joins assets with asset_identifiers
var allowedFields = map[string]map[string]bool{
	"findings": {
		"id":                 true,
		"tenant_id":          true,
		"asset_id":           true,
		"definition_uid":     true,
		"scanner_status":     true,
		"effective_status":   true,
		"effective_reason":   true,
		"effective_revision": true,
		"first_observed_at":  true,
		"last_observed_at":   true,
		"evidence_json":      true,
		"created_at":         true,
		"updated_at":         true,
		"source":             true,
		"severity":           true,
		"title":              true,
		"asset_name":         true,
		"epss_score":         true,
		"epss_percentile":    true,
		"is_kev":             true,
		"kev_date_added":     true,
		"kev_due_date":       true,
		"cve":                true,
		"has_cve":            true,
	},
	"assets": {
		"id":             true,
		"tenant_id":      true,
		"canonical_name": true,
		"first_seen_at":  true,
		"last_seen_at":   true,
		"owner_team_id":  true,
		"is_active":      true,
		"created_at":     true,
		"updated_at":     true,
		"hostname_norm":  true,
		"shortname_norm": true,
		"ipv4":           true,
	},
}

// allowedOperators defines whitelisted operators
// Note: Only operators implemented by Translator should be listed here
var allowedOperators = map[string]bool{
	"eq":         true,
	"neq":        true,
	"in":         true,
	"like":       true,
	"gt":         true,
	"gte":        true,
	"lt":         true,
	"lte":        true,
	"is_null":    true,
	"is_not_null": true,
}

// Validator validates queries against whitelisted fields and operators
type Validator struct{}

// NewValidator creates a new Validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks if the query is valid for the given entity type
func (v *Validator) Validate(entityType string, q *Query) error {
	// Add nil check
	if q == nil {
		return errors.New("query cannot be nil")
	}

	// Check entity type
	fields, ok := allowedFields[entityType]
	if !ok {
		return fmt.Errorf("invalid entity type: %s", entityType)
	}

	// Validate each filter
	for _, f := range q.Filters {
		// Check field is allowed (O(1) lookup)
		if !fields[f.Field] {
			return fmt.Errorf("field '%s' not allowed for entity %s", f.Field, entityType)
		}

		// Check operator is allowed (O(1) lookup)
		if !allowedOperators[f.Operator] {
			return fmt.Errorf("operator '%s' not allowed", f.Operator)
		}
	}

	return nil
}
