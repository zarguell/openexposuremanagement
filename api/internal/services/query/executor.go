package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// MaxQueryLimit is the maximum number of rows allowed in a single query
const MaxQueryLimit = 5000

// QueryTimeout is the maximum duration for query execution
const QueryTimeout = 5 * time.Second

// SlowQueryThreshold is the duration above which queries are logged as slow
const SlowQueryThreshold = 1 * time.Second

// QueryResult represents the result of a query execution
type QueryResult struct {
	Data []map[string]interface{} `json:"data"`
	Meta *QueryMeta                `json:"meta"`
}

// QueryMeta contains metadata about the query execution
type QueryMeta struct {
	TotalRows       int64  `json:"total_rows"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
	HasMore         bool   `json:"has_more"`
}

// Executor executes validated queries
type Executor struct {
	db         *sqlx.DB
	validator  *Validator
	translator *Translator
}

// NewExecutor creates a new Executor
func NewExecutor(db *sqlx.DB) *Executor {
	return &Executor{
		db:         db,
		validator:  NewValidator(),
		translator: NewTranslator(),
	}
}

// Execute runs a query and returns results
func (e *Executor) Execute(ctx context.Context, tenantID, entityType string, q *Query) (*QueryResult, error) {
	// Validate query
	if err := e.validator.Validate(entityType, q); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Enforce maximum query limit to prevent DoS
	if q.Limit != nil && *q.Limit > MaxQueryLimit {
		return nil, fmt.Errorf("limit exceeds maximum of %d", MaxQueryLimit)
	}

	// Only add timeout if parent doesn't have a deadline
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		ctx, cancel = context.WithTimeout(ctx, QueryTimeout)
		defer cancel()
	}

	// Translate to SQL
	sql, args, err := e.translator.Translate(entityType, q)
	if err != nil {
		return nil, fmt.Errorf("translation error: %w", err)
	}

	// Add tenant_id to filters (security)
	tenantFilter := fmt.Sprintf("tenant_id = $%d", len(args)+1)
	args = append(args, tenantID)

	// Inject tenant filter into SQL
	// Must inject before LIMIT/OFFSET clauses if they exist
	if !strings.Contains(sql, "WHERE") {
		// No WHERE clause yet, need to add one
		// But must insert before LIMIT/OFFSET if they exist
		if idx := strings.Index(sql, " LIMIT"); idx != -1 {
			// Insert WHERE before LIMIT
			sql = sql[:idx] + " WHERE " + tenantFilter + sql[idx:]
		} else if idx := strings.Index(sql, " OFFSET"); idx != -1 {
			// Insert WHERE before OFFSET (shouldn't happen without LIMIT, but handle it)
			sql = sql[:idx] + " WHERE " + tenantFilter + sql[idx:]
		} else {
			// No LIMIT/OFFSET, just append WHERE
			sql = sql + " WHERE " + tenantFilter
		}
	} else {
		// WHERE exists, prepend tenant filter to existing conditions
		sql = strings.Replace(sql, "WHERE", "WHERE "+tenantFilter+" AND ", 1)
	}

	// Start timer
	startTime := time.Now()

	// Execute query
	rows, err := e.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		// Check if context was cancelled
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("query timeout after %v", QueryTimeout)
		}
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	// Collect results
	var data []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Calculate execution time
	executionTime := time.Since(startTime)

	// Log slow queries
	if executionTime > SlowQueryThreshold {
		log.Warn().
			Str("entity_type", entityType).
			Str("tenant_id", tenantID).
			Dur("duration_ms", executionTime).
			Str("sql", sql).
			Msg("slow query detected")
	}

	return &QueryResult{
		Data: data,
		Meta: &QueryMeta{
			TotalRows:       int64(len(data)),
			ExecutionTimeMs: executionTime.Milliseconds(),
			HasMore:         q.Limit != nil && len(data) == *q.Limit,
		},
	}, nil
}
