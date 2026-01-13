package query

import (
	"fmt"
)

// allowedFields defines whitelisted fields per entity
var allowedFields = map[string][]string{
	"findings": {
		"severity", "scanner_status", "effective_status",
		"cve", "source", "asset_name", "first_observed_at",
		"last_observed_at", "epss_score", "is_kev", "has_cve",
	},
	"assets": {
		"canonical_name", "hostname_norm", "shortname_norm",
		"ipv4", "first_seen_at", "last_seen_at", "is_active",
	},
}

// allowedOperators defines whitelisted operators
var allowedOperators = []string{
	"eq", "neq", "in", "not_in", "like", "not_like",
	"gt", "gte", "lt", "lte", "between", "is_null", "is_not_null",
}

// Validator validates queries against whitelisted fields and operators
type Validator struct{}

// NewValidator creates a new Validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks if the query is valid for the given entity type
func (v *Validator) Validate(entityType string, q *Query) error {
	// Check entity type
	fields, ok := allowedFields[entityType]
	if !ok {
		return fmt.Errorf("invalid entity type: %s", entityType)
	}

	// Validate each filter
	for _, f := range q.Filters {
		// Check field is allowed
		fieldAllowed := false
		for _, allowed := range fields {
			if f.Field == allowed {
				fieldAllowed = true
				break
			}
		}
		if !fieldAllowed {
			return fmt.Errorf("field '%s' not allowed for entity %s", f.Field, entityType)
		}

		// Check operator is allowed
		opAllowed := false
		for _, allowed := range allowedOperators {
			if f.Operator == allowed {
				opAllowed = true
				break
			}
		}
		if !opAllowed {
			return fmt.Errorf("operator '%s' not allowed", f.Operator)
		}
	}

	return nil
}
