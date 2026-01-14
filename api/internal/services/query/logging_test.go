package query

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestQueryLogging(t *testing.T) {
	t.Run("logs query execution details", func(t *testing.T) {
		// Create buffer to capture log output
		var logBuf bytes.Buffer
		logger := zerolog.New(&logBuf).Level(zerolog.DebugLevel)

		// Create mock executor
		mockExec := &mockExecutor{
			result: &QueryResult{
				Data: []map[string]interface{}{{"id": 1}},
				Meta: &QueryMeta{
					TotalRows: 1,
				},
			},
		}

		// Wrap executor with logging
		executor := WithLogging(mockExec, logger)

		// Execute query
		q := &Query{
			Filters: []Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
		}

		_, err := executor.Execute(context.Background(), "tenant-1", "findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify log output contains expected fields
		logOutput := logBuf.String()

		expectedFields := []string{
			"tenant_id", "tenant-1",
			"entity_type", "findings",
			"filter_count", "1",
			"execution_time_ms",
		}

		for _, field := range expectedFields {
			if !contains(logOutput, field) {
				t.Errorf("log output missing expected field: %s\nLog: %s", field, logOutput)
			}
		}
	})

	t.Run("logs query errors", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := zerolog.New(&logBuf).Level(zerolog.ErrorLevel)

		mockExec := &errorExecutor{
			err: &ValidationError{},
		}

		executor := WithLogging(mockExec, logger)

		q := &Query{
			Filters: []Filter{
				{Field: "severity", Operator: "eq", Value: "invalid"},
			},
		}

		_, err := executor.Execute(context.Background(), "tenant-1", "findings", q)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		logOutput := logBuf.String()

		if !contains(logOutput, "query_failed") {
			t.Error("error log should contain query_failed field")
		}

		if !contains(logOutput, "error") {
			t.Error("error log should contain error field")
		}
	})

	t.Run("logs slow queries", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := zerolog.New(&logBuf).Level(zerolog.WarnLevel)

		mockExec := &slowExecutor{
			delay: 1500 * time.Millisecond,
			result: &QueryResult{
				Data: []map[string]interface{}{{"id": 1}},
				Meta: &QueryMeta{
					TotalRows: 1,
				},
			},
		}

		executor := WithLogging(mockExec, logger)

		q := &Query{
			Filters: []Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
		}

		_, err := executor.Execute(context.Background(), "tenant-1", "findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		logOutput := logBuf.String()

		if !contains(logOutput, "slow_query") {
			t.Error("slow query should be logged with warning level")
		}
	})

	t.Run("logs aggregations count", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := zerolog.New(&logBuf).Level(zerolog.DebugLevel)

		mockExec := &mockExecutor{
			result: &QueryResult{
				Data: []map[string]interface{}{{"count": 100}},
				Meta: &QueryMeta{
					TotalRows: 1,
				},
			},
		}

		executor := WithLogging(mockExec, logger)

		q := &Query{
			Filters: []Filter{
				{Field: "severity", Operator: "eq", Value: "critical"},
			},
			Aggregations: []Aggregation{
				{Type: "count", Field: "id"},
				{Type: "group_by", Field: "severity"},
			},
		}

		_, err := executor.Execute(context.Background(), "tenant-1", "findings", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		logOutput := logBuf.String()

		if !contains(logOutput, "aggregation_count") {
			t.Error("log should contain aggregation_count when aggregations present")
		}

		if !contains(logOutput, "2") {
			t.Error("log should show correct aggregation count")
		}
	})
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// Mock executor that returns successful result
type mockExecutor struct {
	result *QueryResult
}

func (m *mockExecutor) Execute(ctx context.Context, tenantID, entityType string, q *Query) (*QueryResult, error) {
	return m.result, nil
}

// Mock executor that returns error
type errorExecutor struct {
	err error
}

func (m *errorExecutor) Execute(ctx context.Context, tenantID, entityType string, q *Query) (*QueryResult, error) {
	return nil, m.err
}

// Slow executor for testing slow query logging
type slowExecutor struct {
	delay  time.Duration
	result *QueryResult
}

func (m *slowExecutor) Execute(ctx context.Context, tenantID, entityType string, q *Query) (*QueryResult, error) {
	time.Sleep(m.delay)
	return m.result, nil
}

// ValidationError is a test error
type ValidationError struct{}

func (e *ValidationError) Error() string {
	return "validation error"
}
