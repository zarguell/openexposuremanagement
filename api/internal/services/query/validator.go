package query

import (
	"errors"
	"fmt"
	"strings"
)

// allowedFields defines whitelisted fields per entity
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
	"software_inventory": {
		"id":              true,
		"tenant_id":       true,
		"asset_id":        true,
		"software_id":     true,
		"source":          true,
		"install_path":    true,
		"first_seen_at":   true,
		"last_seen_at":    true,
		"created_at":      true,
		"updated_at":      true,
		"cpe_string":      true,
		"vendor":          true,
		"product_name":    true,
		"version":         true,
		"edition":         true,
		"target_hw":       true,
		"lang":            true,
		"title_formatted": true,
		"asset_name":      true,
		"asset_is_active": true,
	},
}

// Entity field mappings for dot-walking syntax
var entityFieldMappings = map[string]map[string]bool{
	"software": {
		"vendor":        true,
		"product_name":  true,
		"version":       true,
		"cpe_string":    true,
		"install_path":  true,
		"first_seen_at": true,
		"last_seen_at":  true,
	},
	"findings": {
		"severity":          true,
		"scanner_status":    true,
		"effective_status":  true,
		"cve":               true,
		"epss_score":        true,
		"is_kev":            true,
		"first_observed_at": true,
		"last_observed_at":  true,
	},
}

// allowedOperators defines whitelisted operators
var allowedOperators = map[string]bool{
	"eq":          true,
	"neq":         true,
	"in":          true,
	"like":        true,
	"gt":          true,
	"gte":         true,
	"lt":          true,
	"lte":         true,
	"is_null":     true,
	"is_not_null": true,
}

// Validator validates queries against whitelisted fields and operators
type Validator struct{}

// NewValidator creates a new Validator
func NewValidator() *Validator {
	return &Validator{}
}

// parseField parses a field name to extract entity prefix
func (v *Validator) parseField(field string) (entityPrefix, fieldName string) {
	parts := strings.SplitN(field, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", field
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
		entityPrefix, fieldName := v.parseField(f.Field)

		// Determine which field whitelist to use
		var validFields map[string]bool
		if entityPrefix == "" {
			// Primary entity field
			validFields = fields
		} else {
			// Related entity field (dot-walking syntax)
			var ok bool
			validFields, ok = entityFieldMappings[entityPrefix]
			if !ok {
				return fmt.Errorf("unknown entity prefix: %s", entityPrefix)
			}
		}

		// Check field is allowed
		if !validFields[fieldName] {
			return fmt.Errorf("field '%s' not allowed for entity %s", f.Field, entityType)
		}

		// Check operator is allowed
		if !allowedOperators[f.Operator] {
			return fmt.Errorf("operator '%s' not allowed", f.Operator)
		}
	}

	// Validate aggregations if present
	for _, agg := range q.Aggregations {
		if strings.TrimSpace(agg.Type) == "" {
			return errors.New("aggregation type cannot be empty")
		}

		// Validate aggregation field
		if agg.Field != "" {
			entityPrefix, fieldName := v.parseField(agg.Field)

			var validFields map[string]bool
			if entityPrefix == "" {
				validFields = fields
			} else {
				var ok bool
				validFields, ok = entityFieldMappings[entityPrefix]
				if !ok {
					return fmt.Errorf("unknown entity prefix in aggregation: %s", entityPrefix)
				}
			}

			if !validFields[fieldName] {
				return fmt.Errorf("aggregation field '%s' not allowed", agg.Field)
			}
		}
	}

	// Validate sort fields if present
	for _, s := range q.Sort {
		entityPrefix, fieldName := v.parseField(s.Field)

		var validFields map[string]bool
		if entityPrefix == "" {
			validFields = fields
		} else {
			var ok bool
			validFields, ok = entityFieldMappings[entityPrefix]
			if !ok {
				return fmt.Errorf("unknown entity prefix in sort: %s", entityPrefix)
			}
		}

		if !validFields[fieldName] {
			return fmt.Errorf("sort field '%s' not allowed", s.Field)
		}

		order := strings.ToLower(strings.TrimSpace(s.Order))
		if order != "asc" && order != "desc" {
			return errors.New("sort order must be 'asc' or 'desc'")
		}
	}

	// Note: We no longer validate the old Join syntax
	// The dot-walking syntax is simpler and doesn't require explicit join configuration

	return nil
}
