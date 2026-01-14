package query_test

import (
	"testing"

	"github.com/openexposuremanagement/oem/internal/services/query"
)

func TestTranslator(t *testing.T) {
	translator := query.NewTranslator()

	t.Run("simple filter query", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
		}
		sql, args, err := translator.Translate("findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql == "" {
			t.Error("sql is empty")
		}
		if len(args) == 0 {
			t.Error("args is empty")
		}
		// Check it's parameterized
		if sql == "severity = 'critical'" {
			t.Error("SQL should be parameterized, not contain literal values")
		}
	})

	t.Run("filter with in operator", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "in", Value: []string{"critical", "high"}},
			},
		}
		sql, args, err := translator.Translate("findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sql == "" {
			t.Error("sql is empty")
		}
		if len(args) != 2 {
			t.Errorf("expected 2 args, got %d", len(args))
		}
	})

	t.Run("query with aggregations", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "effective_status", Operator: "eq", Value: "open"},
			},
			Aggregations: []query.Aggregation{
				{Type: "group_by", Field: "severity"},
				{Type: "count"},
			},
		}
		sql, _, err := translator.Translate("findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should have GROUP BY
		if !contains(sql, "GROUP BY") {
			t.Error("SQL should contain GROUP BY")
		}
		// Should have COUNT
		if !contains(sql, "COUNT") {
			t.Error("SQL should contain COUNT")
		}
	})

	t.Run("query with sort", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
			Sort: []query.Sort{
				{Field: "last_observed_at", Order: "desc"},
			},
		}
		sql, _, err := translator.Translate("findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should have ORDER BY
		if !contains(sql, "ORDER BY") {
			t.Error("SQL should contain ORDER BY")
		}
		// Should have DESC
		if !contains(sql, "DESC") {
			t.Error("SQL should contain DESC")
		}
	})

	t.Run("query with limit and offset", func(t *testing.T) {
		limit := 50
		offset := 10
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
			Limit:  &limit,
			Offset: &offset,
		}
		sql, args, err := translator.Translate("findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should have LIMIT and OFFSET
		if !contains(sql, "LIMIT") {
			t.Error("SQL should contain LIMIT")
		}
		if !contains(sql, "OFFSET") {
			t.Error("SQL should contain OFFSET")
		}
		// Limit and offset should be parameterized
		if len(args) < 3 {
			t.Errorf("expected at least 3 args, got %d", len(args))
		}
	})

	t.Run("unsupported operator returns error", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "invalid_op", Value: "critical"},
			},
		}
		_, _, err := translator.Translate("findings", q)
		if err == nil {
			t.Error("expected error for unsupported operator, got nil")
		}
	})

	t.Run("unsupported aggregation type returns error", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
			Aggregations: []query.Aggregation{
				{Type: "invalid_agg_type", Field: "severity"},
			},
		}
		_, _, err := translator.Translate("findings", q)
		if err == nil {
			t.Error("expected error for unsupported aggregation type, got nil")
		}
	})

	t.Run("in operator with non-array value returns error", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "in", Value: "not_an_array"},
			},
		}
		_, _, err := translator.Translate("findings", q)
		if err == nil {
			t.Error("expected error for non-array value with in operator, got nil")
		}
	})

	t.Run("is_null operator", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "last_observed_at", Operator: "is_null"},
			},
		}
		sql, args, err := translator.Translate("findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "IS NULL") {
			t.Error("SQL should contain IS NULL")
		}
		// IS NULL should not add args
		if len(args) != 0 {
			t.Errorf("expected 0 args for IS NULL, got %d", len(args))
		}
	})

	t.Run("is_not_null operator", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "last_observed_at", Operator: "is_not_null"},
			},
		}
		sql, args, err := translator.Translate("findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "IS NOT NULL") {
			t.Error("SQL should contain IS NOT NULL")
		}
		// IS NOT NULL should not add args
		if len(args) != 0 {
			t.Errorf("expected 0 args for IS NOT NULL, got %d", len(args))
		}
	})

	t.Run("multiple filters with AND", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
				{Field: "effective_status", Operator: "eq", Value: "open"},
			},
		}
		sql, args, err := translator.Translate("findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !contains(sql, "WHERE") {
			t.Error("SQL should contain WHERE")
		}
		// Should have 2 args
		if len(args) != 2 {
			t.Errorf("expected 2 args, got %d", len(args))
		}
	})

	t.Run("all comparison operators", func(t *testing.T) {
		operators := []string{"eq", "neq", "gt", "gte", "lt", "lte", "like"}
		for _, op := range operators {
			t.Run(op, func(t *testing.T) {
				q := &query.Query{
					Filters: []query.Filter{
						{Field: "severity", Operator: op, Value: "test"},
					},
				}
				_, args, err := translator.Translate("findings", q)
				if err != nil {
					t.Fatalf("unexpected error for operator %s: %v", op, err)
				}
				if len(args) != 1 {
					t.Errorf("expected 1 arg for operator %s, got %d", op, len(args))
				}
			})
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (
		s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
