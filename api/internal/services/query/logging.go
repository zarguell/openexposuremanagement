package query

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// QueryExecutor interface defines the Execute method for query executors
type QueryExecutor interface {
	Execute(ctx context.Context, tenantID, entityType string, q *Query) (*QueryResult, error)
}

// LoggingExecutor wraps an Executor and adds logging
type LoggingExecutor struct {
	executor      QueryExecutor
	logger        zerolog.Logger
	slowThreshold time.Duration // Queries slower than this are logged as warnings
}

// WithLogging wraps an executor with debug logging
func WithLogging(executor QueryExecutor, logger zerolog.Logger) *LoggingExecutor {
	return &LoggingExecutor{
		executor:      executor,
		logger:        logger,
		slowThreshold: 1000 * time.Millisecond, // 1 second default
	}
}

// Execute executes a query with logging
func (le *LoggingExecutor) Execute(ctx context.Context, tenantID, entityType string, q *Query) (*QueryResult, error) {
	startTime := time.Now()

	// Log query start
	le.logger.Debug().
		Str("tenant_id", tenantID).
		Str("entity_type", entityType).
		Int("filter_count", len(q.Filters)).
		Int("aggregation_count", len(q.Aggregations)).
		Msg("executing query")

	// Execute query
	result, err := le.executor.Execute(ctx, tenantID, entityType, q)

	executionTime := time.Since(startTime)
	executionTimeMs := executionTime.Milliseconds()

	if err != nil {
		// Log query error
		le.logger.Error().
			Str("tenant_id", tenantID).
			Str("entity_type", entityType).
			Int64("execution_time_ms", executionTimeMs).
			Err(err).
			Msg("query_failed")
		return nil, err
	}

	// Log query completion
	event := le.logger.Info().
		Str("tenant_id", tenantID).
		Str("entity_type", entityType).
		Int64("execution_time_ms", executionTimeMs).
		Int64("total_rows", result.Meta.TotalRows)

	// Warn for slow queries
	if executionTime > le.slowThreshold {
		event = le.logger.Warn()
		event.Str("slow_query", "true")
	}

	event.Msg("query_completed")

	return result, nil
}
