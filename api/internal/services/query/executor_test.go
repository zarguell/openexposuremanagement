package query

import (
	"context"
	"testing"
)

func TestExecutor_MaxLimitValidation(t *testing.T) {
	executor := NewExecutor(nil) // DB not needed for validation test

	t.Run("rejects limit exceeding maximum", func(t *testing.T) {
		limit := 1001
		q := &Query{
			Filters: []Filter{
				{Field: "effective_status", Operator: "eq", Value: "open"},
			},
			Limit: &limit,
		}

		_, err := executor.Execute(context.Background(), "test-tenant", "findings", q)
		if err == nil {
			t.Fatal("expected error for limit exceeding maximum, got nil")
		}

		if err.Error() != "limit exceeds maximum of 1000" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("accepts limit at maximum", func(t *testing.T) {
		limit := 1000
		q := &Query{
			Filters: []Filter{
				{Field: "effective_status", Operator: "eq", Value: "open"},
			},
			Limit: &limit,
		}

		// Catch panic from nil DB
		defer func() {
			if r := recover(); r != nil {
				// Expected - DB is nil, but validation passed
			}
		}()

		// Will fail at DB execution (nil DB), but should pass validation
		_, err := executor.Execute(context.Background(), "test-tenant", "findings", q)
		// We expect an error (nil DB panic), but NOT the limit error
		if err != nil && err.Error() == "limit exceeds maximum of 1000" {
			t.Error("limit at maximum should be accepted")
		}
	})

	t.Run("accepts limit below maximum", func(t *testing.T) {
		limit := 100
		q := &Query{
			Filters: []Filter{
				{Field: "effective_status", Operator: "eq", Value: "open"},
			},
			Limit: &limit,
		}

		// Catch panic from nil DB
		defer func() {
			if r := recover(); r != nil {
				// Expected - DB is nil, but validation passed
			}
		}()

		// Will fail at DB execution (nil DB), but should pass validation
		_, err := executor.Execute(context.Background(), "test-tenant", "findings", q)
		// We expect an error (nil DB panic), but NOT the limit error
		if err != nil && err.Error() == "limit exceeds maximum of 1000" {
			t.Error("limit below maximum should be accepted")
		}
	})
}
