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

	t.Run("empty aggregation type fails", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
			Aggregations: []query.Aggregation{
				{Type: "", Field: "severity"},
			},
		}
		if err := q.Validate(); err == nil {
			t.Error("expected error for empty aggregation type")
		}
	})

	t.Run("empty sort field fails", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
			Sort: []query.Sort{
				{Field: "", Order: "asc"},
			},
		}
		if err := q.Validate(); err == nil {
			t.Error("expected error for empty sort field")
		}
	})

	t.Run("invalid sort order fails", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
			Sort: []query.Sort{
				{Field: "severity", Order: "invalid"},
			},
		}
		if err := q.Validate(); err == nil {
			t.Error("expected error for invalid sort order")
		}
	})

	t.Run("accepts uppercase sort order", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
			Sort: []query.Sort{
				{Field: "severity", Order: "ASC"},
			},
		}
		if err := q.Validate(); err != nil {
			t.Errorf("should accept ASC: %v", err)
		}
	})

	t.Run("accepts mixed case sort order", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
			Sort: []query.Sort{
				{Field: "severity", Order: "DeSc"},
			},
		}
		if err := q.Validate(); err != nil {
			t.Errorf("should accept DeSc: %v", err)
		}
	})
}
