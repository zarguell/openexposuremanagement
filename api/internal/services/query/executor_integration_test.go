package query_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/openexposuremanagement/oem/internal/services/query"
	_ "github.com/lib/pq"
)

// This is an integration test - requires test database
func TestExecutorIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Connect to test database
	db, err := sqlx.Connect("postgres", "dbname=oem_test sslmode=disable")
	if err != nil {
		t.Skipf("cannot connect to test db: %v", err)
	}
	defer db.Close()

	executor := query.NewExecutor(db)

	t.Run("execute simple findings query", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "effective_status", Operator: "eq", Value: "open"},
			},
			Limit: intPtr(10),
		}

		result, err := executor.Execute(context.Background(), "test-tenant-id", "findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("result is nil")
		}

		if result.Meta == nil {
			t.Error("meta is nil")
		} else {
			if result.Meta.ExecutionTimeMs == 0 {
				t.Error("execution time should be recorded")
			}
		}
	})

	t.Run("execute aggregation query", func(t *testing.T) {
		q := &query.Query{
			Filters: []query.Filter{
				{Field: "effective_status", Operator: "eq", Value: "open"},
			},
			Aggregations: []query.Aggregation{
				{Type: "group_by", Field: "severity"},
				{Type: "count"},
			},
		}

		result, err := executor.Execute(context.Background(), "test-tenant-id", "findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("result is nil")
		}
	})
}

func intPtr(i int) *int {
	return &i
}
