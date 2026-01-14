package query

import (
	"errors"
	"fmt"
)

// allowedFields defines whitelisted fields per entity
var allowedFields = map[string]map[string]bool{
	"findings": {
		"severity":          true,
		"scanner_status":    true,
		"effective_status":  true,
		"cve":               true,
		"source":            true,
		"asset_name":        true,
		"first_observed_at": true,
		"last_observed_at":  true,
		"epss_score":        true,
		"is_kev":            true,
		"has_cve":           true,
	},
	"assets": {
		"canonical_name": true,
		"hostname_norm":  true,
		"shortname_norm": true,
		"ipv4":           true,
		"first_seen_at":  true,
		"last_seen_at":   true,
		"is_active":      true,
	},
}

// allowedOperators defines whitelisted operators
var allowedOperators = map[string]bool{
	"eq":         true,
	"neq":        true,
	"in":         true,
	"not_in":     true,
	"like":       true,
	"not_like":   true,
	"gt":         true,
	"gte":        true,
	"lt":         true,
	"lte":        true,
	"between":    true,
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
