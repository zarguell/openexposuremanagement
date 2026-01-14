package query

import (
	"fmt"
	"strings"
)

// Translator converts Query objects to SQL
type Translator struct{}

// NewTranslator creates a new Translator
func NewTranslator() *Translator {
	return &Translator{}
}

// Translate converts a query to SQL and returns sql + args
func (tr *Translator) Translate(entityType string, q *Query) (string, []interface{}, error) {
	var whereParts []string
	var args []interface{}
	argPos := 1

	// Build WHERE clause
	for _, f := range q.Filters {
		var part string
		var newArgs []interface{}

		switch f.Operator {
		case "eq":
			part = fmt.Sprintf("%s = $%d", f.Field, argPos)
			newArgs = []interface{}{f.Value}
			argPos++
		case "neq":
			part = fmt.Sprintf("%s != $%d", f.Field, argPos)
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
			part = fmt.Sprintf("%s IN (%s)", f.Field, strings.Join(placeholders, ", "))
		case "like":
			part = fmt.Sprintf("%s LIKE $%d", f.Field, argPos)
			newArgs = []interface{}{f.Value}
			argPos++
		case "gt":
			part = fmt.Sprintf("%s > $%d", f.Field, argPos)
			newArgs = []interface{}{f.Value}
			argPos++
		case "gte":
			part = fmt.Sprintf("%s >= $%d", f.Field, argPos)
			newArgs = []interface{}{f.Value}
			argPos++
		case "lt":
			part = fmt.Sprintf("%s < $%d", f.Field, argPos)
			newArgs = []interface{}{f.Value}
			argPos++
		case "lte":
			part = fmt.Sprintf("%s <= $%d", f.Field, argPos)
			newArgs = []interface{}{f.Value}
			argPos++
		case "is_null":
			part = fmt.Sprintf("%s IS NULL", f.Field)
		case "is_not_null":
			part = fmt.Sprintf("%s IS NOT NULL", f.Field)
		default:
			return "", nil, fmt.Errorf("unsupported operator: %s", f.Operator)
		}

		whereParts = append(whereParts, part)
		args = append(args, newArgs...)
	}

	// Build base query
	var selectClause string
	var groupByClause string
	var orderByClause string
	var limitClause string
	var offsetClause string

	// Handle aggregations
	if len(q.Aggregations) > 0 {
		var selectParts []string
		var groupByFields []string

		for _, agg := range q.Aggregations {
			switch agg.Type {
			case "count":
				if agg.Field != "" {
					selectParts = append(selectParts, fmt.Sprintf("COUNT(%s)", agg.Field))
				} else {
					selectParts = append(selectParts, "COUNT(*)")
				}
			case "group_by":
				selectParts = append(selectParts, agg.Field)
				groupByFields = append(groupByFields, agg.Field)
			}
		}
		selectClause = strings.Join(selectParts, ", ")
		if len(groupByFields) > 0 {
			groupByClause = "GROUP BY " + strings.Join(groupByFields, ", ")
		}
	} else {
		selectClause = "*"
	}

	// Handle sort
	if len(q.Sort) > 0 {
		var sortParts []string
		for _, s := range q.Sort {
			sortParts = append(sortParts, fmt.Sprintf("%s %s", s.Field, strings.ToUpper(s.Order)))
		}
		orderByClause = "ORDER BY " + strings.Join(sortParts, ", ")
	}

	// Handle limit
	if q.Limit != nil {
		limitClause = fmt.Sprintf("LIMIT $%d", argPos)
		args = append(args, *q.Limit)
		argPos++
	}

	// Handle offset
	if q.Offset != nil {
		offsetClause = fmt.Sprintf("OFFSET $%d", argPos)
		args = append(args, *q.Offset)
	}

	// Assemble final query
	var queryParts []string
	queryParts = append(queryParts, fmt.Sprintf("SELECT %s FROM %s", selectClause, entityType))
	if len(whereParts) > 0 {
		queryParts = append(queryParts, "WHERE "+strings.Join(whereParts, " AND "))
	}
	if groupByClause != "" {
		queryParts = append(queryParts, groupByClause)
	}
	if orderByClause != "" {
		queryParts = append(queryParts, orderByClause)
	}
	if limitClause != "" {
		queryParts = append(queryParts, limitClause)
	}
	if offsetClause != "" {
		queryParts = append(queryParts, offsetClause)
	}

	return strings.Join(queryParts, " "), args, nil
}
