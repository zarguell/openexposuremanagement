package query

import (
	"fmt"
	"strings"
)

// entityToTable maps entity types to their actual database table/view names
var entityToTable = map[string]string{
	"findings":           "findings",            // Uses the "findings" view
	"assets":             "assets_extended",     // Uses the "assets_extended" view
	"software_inventory": "software_inventory", // Uses the "software_inventory" view
}

// entityRelationships defines how entities relate to assets
// Maps entity prefix to table name and join condition
var entityRelationships = map[string]struct {
	table      string
	joinColumn string // Column in joined table that references assets.id
	alias      string // SQL alias for the table
}{
	"software": {
		table:      "software_inventory",
		joinColumn: "asset_id",
		alias:      "software_inventory",
	},
	"findings": {
		table:      "findings",
		joinColumn: "asset_id",
		alias:      "findings",
	},
}

// Fields that belong to related entities (for validation)
var relatedEntityFields = map[string]map[string]bool{
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

// Primary entity fields (assets)
var primaryEntityFields = map[string]bool{
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
}

// parsedField represents a field that has been parsed for entity prefix
type parsedField struct {
	entityPrefix string // "software", "findings", or "" for primary entity
	fieldName    string // The actual field name
}

// Translator converts Query objects to SQL
type Translator struct{}

// NewTranslator creates a new Translator
func NewTranslator() *Translator {
	return &Translator{}
}

// parseField parses a field name to extract entity prefix
// e.g., "software.vendor" -> {entityPrefix: "software", fieldName: "vendor"}
// e.g., "is_active" -> {entityPrefix: "", fieldName: "is_active"}
func (tr *Translator) parseField(field string) parsedField {
	parts := strings.SplitN(field, ".", 2)
	if len(parts) == 2 {
		return parsedField{
			entityPrefix: parts[0],
			fieldName:    parts[1],
		}
	}
	return parsedField{
		entityPrefix: "",
		fieldName:    field,
	}
}

// Translate converts a validated Query object to parameterized SQL using dot-walking syntax.
//
// The dot-walking syntax automatically determines JOIN logic:
// - Fields without dots (e.g., "is_active") filter on primary entity (assets)
// - Fields with dots (e.g., "software.vendor") imply a JOIN
// - Negated filters (negate: true) on related entities use NOT EXISTS
// - Non-negated filters on related entities use INNER JOIN
//
// WARNING: This function assumes the query has already been validated
// by Validator.Validate() to prevent SQL injection.
//
// Returns: SQL string with $1, $2, ... placeholders and args array
func (tr *Translator) Translate(primaryEntityType string, q *Query) (string, []interface{}, error) {
	var whereParts []string
	var notExistsParts []string
	var innerJoinEntities []string // Entities to INNER JOIN
	var args []interface{}
	argPos := 1

	// First pass: categorize filters and detect which entities to join
	for _, f := range q.Filters {
		parsed := tr.parseField(f.Field)

		if parsed.entityPrefix == "" {
			// Primary entity filter
			var part string
			var newArgs []interface{}

			qualifiedField := fmt.Sprintf("%s.%s", primaryEntityType, parsed.fieldName)

			switch f.Operator {
			case "eq":
				part = fmt.Sprintf("%s = $%d", qualifiedField, argPos)
				newArgs = []interface{}{f.Value}
				argPos++
			case "neq":
				part = fmt.Sprintf("%s != $%d", qualifiedField, argPos)
				newArgs = []interface{}{f.Value}
				argPos++
			case "in":
				values, ok := f.Value.([]string)
				if !ok {
					return "", nil, fmt.Errorf("in operator requires array of strings")
				}
				placeholders := make([]string, len(values))
				newArgs = make([]interface{}, len(values))
				for i, v := range values {
					placeholders[i] = fmt.Sprintf("$%d", argPos)
					newArgs[i] = v
					argPos++
				}
				part = fmt.Sprintf("%s IN (%s)", qualifiedField, strings.Join(placeholders, ", "))
			case "like":
				part = fmt.Sprintf("%s LIKE $%d", qualifiedField, argPos)
				newArgs = []interface{}{f.Value}
				argPos++
			case "gt":
				part = fmt.Sprintf("%s > $%d", qualifiedField, argPos)
				newArgs = []interface{}{f.Value}
				argPos++
			case "gte":
				part = fmt.Sprintf("%s >= $%d", qualifiedField, argPos)
				newArgs = []interface{}{f.Value}
				argPos++
			case "lt":
				part = fmt.Sprintf("%s < $%d", qualifiedField, argPos)
				newArgs = []interface{}{f.Value}
				argPos++
			case "lte":
				part = fmt.Sprintf("%s <= $%d", qualifiedField, argPos)
				newArgs = []interface{}{f.Value}
				argPos++
			case "is_null":
				part = fmt.Sprintf("%s IS NULL", qualifiedField)
			case "is_not_null":
				part = fmt.Sprintf("%s IS NOT NULL", qualifiedField)
			default:
				return "", nil, fmt.Errorf("unsupported operator: %s", f.Operator)
			}

			whereParts = append(whereParts, part)
			args = append(args, newArgs...)

		} else {
			// Related entity filter
			entityRel, ok := entityRelationships[parsed.entityPrefix]
			if !ok {
				return "", nil, fmt.Errorf("unknown entity prefix: %s", parsed.entityPrefix)
			}

			if f.Negate {
				// Use NOT EXISTS for negated filters
				var subqueryWhere string
				var subqueryArgs []interface{}

				qualifiedField := fmt.Sprintf("%s.%s", entityRel.alias, parsed.fieldName)

				switch f.Operator {
				case "eq":
					subqueryWhere = fmt.Sprintf("%s = $%d", qualifiedField, len(args)+1)
					subqueryArgs = []interface{}{f.Value}
				case "neq":
					subqueryWhere = fmt.Sprintf("%s != $%d", qualifiedField, len(args)+1)
					subqueryArgs = []interface{}{f.Value}
				case "like":
					subqueryWhere = fmt.Sprintf("%s LIKE $%d", qualifiedField, len(args)+1)
					subqueryArgs = []interface{}{f.Value}
				case "in":
					values, ok := f.Value.([]string)
					if !ok {
						return "", nil, fmt.Errorf("in operator requires array of strings")
					}
					placeholders := make([]string, len(values))
					subqueryArgs = make([]interface{}, len(values))
					for i, v := range values {
						placeholders[i] = fmt.Sprintf("$%d", len(args)+i+1)
						subqueryArgs[i] = v
					}
					subqueryWhere = fmt.Sprintf("%s IN (%s)", qualifiedField, strings.Join(placeholders, ", "))
				default:
					return "", nil, fmt.Errorf("unsupported operator for negated filter: %s", f.Operator)
				}

				// Build NOT EXISTS subquery
				onClause := fmt.Sprintf("%s.%s = %s.%s",
					primaryEntityType, "id",
					entityRel.alias, entityRel.joinColumn)
				notExists := fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s AS %s WHERE %s AND %s)",
					entityRel.table, entityRel.alias, onClause, subqueryWhere)
				notExistsParts = append(notExistsParts, notExists)
				args = append(args, subqueryArgs...)

			} else {
				// Use INNER JOIN for non-negated filters
				// Track that we need to INNER JOIN this entity
				innerJoinEntities = append(innerJoinEntities, parsed.entityPrefix)

				// Add the filter condition
				var part string
				var newArgs []interface{}

				qualifiedField := fmt.Sprintf("%s.%s", entityRel.alias, parsed.fieldName)

				switch f.Operator {
				case "eq":
					part = fmt.Sprintf("%s = $%d", qualifiedField, argPos)
					newArgs = []interface{}{f.Value}
					argPos++
				case "neq":
					part = fmt.Sprintf("%s != $%d", qualifiedField, argPos)
					newArgs = []interface{}{f.Value}
					argPos++
				case "in":
					values, ok := f.Value.([]string)
					if !ok {
						return "", nil, fmt.Errorf("in operator requires array of strings")
					}
					placeholders := make([]string, len(values))
					newArgs = make([]interface{}, len(values))
					for i, v := range values {
						placeholders[i] = fmt.Sprintf("$%d", argPos)
						newArgs[i] = v
						argPos++
					}
					part = fmt.Sprintf("%s IN (%s)", qualifiedField, strings.Join(placeholders, ", "))
				case "like":
					part = fmt.Sprintf("%s LIKE $%d", qualifiedField, argPos)
					newArgs = []interface{}{f.Value}
					argPos++
				case "gt":
					part = fmt.Sprintf("%s > $%d", qualifiedField, argPos)
					newArgs = []interface{}{f.Value}
					argPos++
				case "gte":
					part = fmt.Sprintf("%s >= $%d", qualifiedField, argPos)
					newArgs = []interface{}{f.Value}
					argPos++
				case "lt":
					part = fmt.Sprintf("%s < $%d", qualifiedField, argPos)
					newArgs = []interface{}{f.Value}
					argPos++
				case "lte":
					part = fmt.Sprintf("%s <= $%d", qualifiedField, argPos)
					newArgs = []interface{}{f.Value}
					argPos++
				case "is_null":
					part = fmt.Sprintf("%s IS NULL", qualifiedField)
				case "is_not_null":
					part = fmt.Sprintf("%s IS NOT NULL", qualifiedField)
				default:
					return "", nil, fmt.Errorf("unsupported operator: %s", f.Operator)
				}

				whereParts = append(whereParts, part)
				args = append(args, newArgs...)
			}
		}
	}

	// Build SELECT clause
	// Add DISTINCT if we have INNER JOINs to prevent duplicate rows
	selectClause := "SELECT *"
	if len(innerJoinEntities) > 0 {
		selectClause = fmt.Sprintf("SELECT DISTINCT %s.*", primaryEntityType)
	}

	// Build FROM clause
	tableName := entityToTable[primaryEntityType]
	if tableName == "" {
		tableName = primaryEntityType
	}
	fromClause := fmt.Sprintf("FROM %s AS %s", tableName, primaryEntityType)

	// Add INNER JOINs (deduplicate)
	seenEntities := make(map[string]bool)
	for _, entityPrefix := range innerJoinEntities {
		if !seenEntities[entityPrefix] {
			seenEntities[entityPrefix] = true
			entityRel := entityRelationships[entityPrefix]
			onClause := fmt.Sprintf("%s.%s = %s.%s",
				primaryEntityType, "id",
				entityRel.alias, entityRel.joinColumn)
			fromClause += fmt.Sprintf(" INNER JOIN %s AS %s ON %s",
				entityRel.table, entityRel.alias, onClause)
		}
	}

	// Combine WHERE parts
	var allWhereParts []string
	allWhereParts = append(allWhereParts, whereParts...)
	allWhereParts = append(allWhereParts, notExistsParts...)

	// Build full query
	var queryParts []string
	queryParts = append(queryParts, fmt.Sprintf("%s %s", selectClause, fromClause))
	if len(allWhereParts) > 0 {
		queryParts = append(queryParts, "WHERE "+strings.Join(allWhereParts, " AND "))
	}

	// Handle aggregations
	if len(q.Aggregations) > 0 {
		var selectParts []string
		var groupByFields []string

		for _, agg := range q.Aggregations {
			switch agg.Type {
			case "count":
				if agg.Field != "" {
					parsed := tr.parseField(agg.Field)
					if parsed.entityPrefix == "" {
						selectParts = append(selectParts, fmt.Sprintf("COUNT(%s.%s)", primaryEntityType, parsed.fieldName))
					} else {
						entityRel := entityRelationships[parsed.entityPrefix]
						selectParts = append(selectParts, fmt.Sprintf("COUNT(%s.%s)", entityRel.alias, parsed.fieldName))
					}
				} else {
					selectParts = append(selectParts, "COUNT(*)")
				}
			case "group_by":
				parsed := tr.parseField(agg.Field)
				if parsed.entityPrefix == "" {
					selectParts = append(selectParts, fmt.Sprintf("%s.%s", primaryEntityType, parsed.fieldName))
					groupByFields = append(groupByFields, fmt.Sprintf("%s.%s", primaryEntityType, parsed.fieldName))
				} else {
					entityRel := entityRelationships[parsed.entityPrefix]
					selectParts = append(selectParts, fmt.Sprintf("%s.%s", entityRel.alias, parsed.fieldName))
					groupByFields = append(groupByFields, fmt.Sprintf("%s.%s", entityRel.alias, parsed.fieldName))
				}
			default:
				return "", nil, fmt.Errorf("unsupported aggregation type: %s", agg.Type)
			}
		}

		if len(selectParts) > 0 {
			queryParts[0] = "SELECT " + strings.Join(selectParts, ", ") + " " + fromClause
			// Re-add INNER JOINs if needed for aggregations
			for entityPrefix := range seenEntities {
				entityRel := entityRelationships[entityPrefix]
				onClause := fmt.Sprintf("%s.%s = %s.%s",
					primaryEntityType, "id",
					entityRel.alias, entityRel.joinColumn)
				queryParts[0] += fmt.Sprintf(" INNER JOIN %s AS %s ON %s",
					entityRel.table, entityRel.alias, onClause)
			}
			if len(groupByFields) > 0 {
				queryParts = append(queryParts, "GROUP BY "+strings.Join(groupByFields, ", "))
			}
		}
	}

	// Handle sort
	if len(q.Sort) > 0 {
		var sortParts []string
		for _, s := range q.Sort {
			parsed := tr.parseField(s.Field)
			if parsed.entityPrefix == "" {
				sortParts = append(sortParts, fmt.Sprintf("%s.%s %s", primaryEntityType, parsed.fieldName, strings.ToUpper(s.Order)))
			} else {
				// Sorting on related entity fields - may need INNER JOIN
				entityRel := entityRelationships[parsed.entityPrefix]
				sortParts = append(sortParts, fmt.Sprintf("%s.%s %s", entityRel.alias, parsed.fieldName, strings.ToUpper(s.Order)))
			}
		}
		if len(sortParts) > 0 {
			queryParts = append(queryParts, "ORDER BY "+strings.Join(sortParts, ", "))
		}
	}

	// Handle limit and offset
	if q.Limit != nil {
		queryParts = append(queryParts, fmt.Sprintf("LIMIT $%d", argPos))
		args = append(args, *q.Limit)
		argPos++
	}

	if q.Offset != nil {
		queryParts = append(queryParts, fmt.Sprintf("OFFSET $%d", argPos))
		args = append(args, *q.Offset)
	}

	return strings.Join(queryParts, " "), args, nil
}
