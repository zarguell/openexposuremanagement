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

func TestJoinValidation(t *testing.T) {
	t.Run("valid left join passes", func(t *testing.T) {
		join := query.Join{
			Entity: "software_inventory",
			Type:   "left",
			On: query.JoinCondition{
				Primary: "id",
				Joined:  "asset_id",
			},
		}

		if err := join.Validate(); err != nil {
			t.Errorf("expected valid join, got error: %v", err)
		}
	})

	t.Run("invalid join type fails", func(t *testing.T) {
		join := query.Join{
			Entity: "software_inventory",
			Type:   "inner",
			On: query.JoinCondition{
				Primary: "id",
				Joined:  "asset_id",
			},
		}

		if err := join.Validate(); err == nil {
			t.Error("expected error for invalid join type, got nil")
		}
	})

	t.Run("unsupported join entity fails", func(t *testing.T) {
		join := query.Join{
			Entity: "unsupported_entity",
			Type:   "left",
			On: query.JoinCondition{
				Primary: "id",
				Joined:  "asset_id",
			},
		}

		if err := join.Validate(); err == nil {
			t.Error("expected error for unsupported entity, got nil")
		}
	})

	t.Run("empty join condition fails", func(t *testing.T) {
		join := query.Join{
			Entity: "software_inventory",
			Type:   "left",
			On: query.JoinCondition{
				Primary: "",
				Joined:  "",
			},
		}

		if err := join.Validate(); err == nil {
			t.Error("expected error for empty join condition, got nil")
		}
	})

	t.Run("query with valid join passes", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "is_active", Operator: "eq", Value: true},
			},
			Join: &query.Join{
				Entity: "software_inventory",
				Type:   "left",
				On: query.JoinCondition{
					Primary: "id",
					Joined:  "asset_id",
				},
			},
		}

		if err := q.Validate(); err != nil {
			t.Errorf("expected valid query with join, got error: %v", err)
		}
	})
}
