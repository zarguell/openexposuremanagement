package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// MaxQueryLimit is the maximum number of rows allowed in a single query
const MaxQueryLimit = 1000

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

	// Translate to SQL
	sql, args, err := e.translator.Translate(entityType, q)
	if err != nil {
		return nil, fmt.Errorf("translation error: %w", err)
	}

	// Add tenant_id to filters (security)
	tenantFilter := fmt.Sprintf("tenant_id = $%d", len(args)+1)
	args = append(args, tenantID)

	// Inject tenant filter into SQL
	// For now, simple approach - prepend to WHERE or create WHERE if none exists
	if !strings.Contains(sql, "WHERE") {
		sql = sql + " WHERE " + tenantFilter
	} else {
		sql = strings.Replace(sql, "WHERE", "WHERE "+tenantFilter+" AND ", 1)
	}

	// Start timer
	startTime := time.Now()

	// Execute query
	rows, err := e.db.QueryxContext(ctx, sql, args...)
	if err != nil {
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
	executionTime := time.Since(startTime).Milliseconds()

	return &QueryResult{
		Data: data,
		Meta: &QueryMeta{
			TotalRows:       int64(len(data)),
			ExecutionTimeMs: executionTime,
			HasMore:         q.Limit != nil && len(data) == *q.Limit,
		},
	}, nil
}
