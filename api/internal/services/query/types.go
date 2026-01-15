package query

import (
	"errors"
	"fmt"
	"strings"
)

// Query represents a JSON query definition
type Query struct {
	Filters      []Filter      `json:"filters"`
	Join         *Join         `json:"join,omitempty"`
	Aggregations []Aggregation `json:"aggregations,omitempty"`
	Sort         []Sort        `json:"sort,omitempty"`
	Limit        *int          `json:"limit,omitempty"`
	Offset       *int          `json:"offset,omitempty"`
}

// IntPtr returns a pointer to an int (helper for tests)
func IntPtr(i int) *int {
	return &i
}

// Filter represents a single filter condition
type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Negate   bool        `json:"negate,omitempty"` // If true, negates the filter (NOT EXISTS for related entities)
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

// JoinCondition defines the ON clause for a join
type JoinCondition struct {
	Primary string `json:"primary"` // Field from primary entity
	Joined  string `json:"joined"`  // Field from joined entity
}

// Join defines a join to another entity
type Join struct {
	Entity string        `json:"entity"` // Entity to join (software_inventory, findings)
	Type   string        `json:"type"`   // "left" only for MVP
	On     JoinCondition `json:"on"`     // Join condition
	Filter *Filter       `json:"filter,omitempty"` // Optional filter to apply to joined entity for anti-join patterns
}

// Validate checks if the join configuration is valid
func (j *Join) Validate() error {
	// Prevent circular references (entity joining to itself)
	// For MVP, we know the primary entity is always "assets" when this is called
	if j.Entity == "assets" {
		return fmt.Errorf("circular reference not allowed: cannot join assets to assets")
	}

	allowedEntities := map[string]bool{
		"software_inventory": true,
		"findings":           true,
	}

	if !allowedEntities[j.Entity] {
		return fmt.Errorf("unsupported join entity: %s", j.Entity)
	}

	if j.Type != "left" {
		return fmt.Errorf("unsupported join type: %s (only 'left' allowed)", j.Type)
	}

	if j.On.Primary == "" || j.On.Joined == "" {
		return fmt.Errorf("join condition must specify both 'primary' and 'joined' fields")
	}

	return nil
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

	// Validate join if present
	if q.Join != nil {
		if err := q.Join.Validate(); err != nil {
			return err
		}
	}

	// Validate aggregations if present
	for _, agg := range q.Aggregations {
		if strings.TrimSpace(agg.Type) == "" {
			return errors.New("aggregation type cannot be empty")
		}
	}

	// Validate sort if present
	for _, s := range q.Sort {
		if strings.TrimSpace(s.Field) == "" {
			return errors.New("sort field cannot be empty")
		}
		order := strings.ToLower(strings.TrimSpace(s.Order))
		if order != "asc" && order != "desc" {
			return errors.New("sort order must be 'asc' or 'desc'")
		}
	}

	return nil
}
