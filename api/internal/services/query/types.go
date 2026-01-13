package query

import (
	"errors"
	"strings"
)

// Query represents a JSON query definition
type Query struct {
	Filters      []Filter      `json:"filters"`
	Aggregations []Aggregation `json:"aggregations,omitempty"`
	Sort         []Sort        `json:"sort,omitempty"`
	Limit        *int          `json:"limit,omitempty"`
	Offset       *int          `json:"offset,omitempty"`
}

// Filter represents a single filter condition
type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// Aggregation represents an aggregation operation
type Aggregation struct {
	Type  string `json:"type"` // count, sum, max, min, group_by
	Field string `json:"field,omitempty"`
}

// Sort represents sort order
type Sort struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc, desc
}

// Validate performs basic validation on the query
func (q *Query) Validate() error {
	if len(q.Filters) == 0 {
		return errors.New("query must have at least one filter")
	}

	for _, f := range q.Filters {
		if strings.TrimSpace(f.Field) == "" {
			return errors.New("filter field cannot be empty")
		}
		if strings.TrimSpace(f.Operator) == "" {
			return errors.New("filter operator cannot be empty")
		}
	}

	return nil
}
