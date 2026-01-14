package query_test

import (
	"testing"

	"github.com/openexposuremanagement/oem/internal/services/query"
)

func TestValidator(t *testing.T) {
	v := query.NewValidator()

	t.Run("valid findings query", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "in", Value: []string{"critical", "high"}},
				{Field: "has_cve", Operator: "eq", Value: true},
			},
		}
		err := v.Validate("findings", q)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid assets query", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "is_active", Operator: "eq", Value: true},
			},
		}
		err := v.Validate("assets", q)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid field for findings", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "invalid_field", Operator: "eq", Value: "test"},
			},
		}
		err := v.Validate("findings", q)
		if err == nil {
			t.Error("expected error for invalid field")
		}
	})

	t.Run("invalid operator", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "invalid_op", Value: "test"},
			},
		}
		err := v.Validate("findings", q)
		if err == nil {
			t.Error("expected error for invalid operator")
		}
	})

	t.Run("invalid entity type", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "severity", Operator: "eq", Value: "test"},
			},
		}
		err := v.Validate("invalid_entity", q)
		if err == nil {
			t.Error("expected error for invalid entity type")
		}
	})

	t.Run("nil query returns error", func(t *testing.T) {
		err := v.Validate("findings", nil)
		if err == nil {
			t.Error("expected error for nil query")
		}
		if err.Error() != "query cannot be nil" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("removed operators are rejected", func(t *testing.T) {
		removedOperators := []string{"not_in", "not_like", "between"}

		for _, op := range removedOperators {
			q := &query.Query{
				Filters: []query.Filter{
					{Field: "severity", Operator: op, Value: "critical"},
				},
			}
			err := v.Validate("findings", q)
			if err == nil {
				t.Errorf("%s operator should be rejected", op)
			}
		}
	})
}
