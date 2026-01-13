package query_test

import (
	"testing"

	"github.com/openexposuremanagement/oem/internal/services/query"
)

func TestQueryValidation(t *testing.T) {
	t.Run("valid query with filters", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
		}
		if err := q.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid query with aggregations", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "effective_status", Operator: "eq", Value: "open"},
			},
			Aggregations: []query.Aggregation{
				{Type: "group_by", Field: "severity"},
				{Type: "count"},
			},
		}
		if err := q.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("query with empty field fails", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "", Operator: "eq", Value: "test"},
			},
		}
		if err := q.Validate(); err == nil {
			t.Error("expected error for empty field")
		}
	})
}
