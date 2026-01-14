# Query Framework Implementation Plan - Phase 1: Backend

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a generalized query framework that accepts JSON query definitions, validates them, translates to safe SQL, and executes with proper error handling and observability.

**Architecture:** The query framework consists of three layers: 1) Validator that whitelists fields/operators, 2) SQL Translator that builds parameterized queries, 3) Executor that runs queries and returns structured data. Request ID middleware enables tracing.

**Tech Stack:** Go 1.21+, PostgreSQL 16, sqlx, zerolog, gorilla/mux, golang-migrate

---

## Task 1: Add Request ID Middleware for Tracing

**Files:**
- Create: `api/internal/middleware/request_id.go`
- Modify: `api/internal/server/server.go` (wire middleware)
- Test: `api/internal/middleware/request_id_test.go`

**Step 1: Write the failing test**

Create `api/internal/middleware/request_id_test.go`:

```go
package middleware_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/openexposuremanagement/api/internal/middleware"
)

func TestRequestID(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqID := r.Context().Value("request_id")
        if reqID == nil {
            t.Fatal("request_id not in context")
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    })

    h := middleware.RequestID(handler)
    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()

    h.ServeHTTP(w, req)

    resp := w.Result()
    if resp.StatusCode != http.StatusOK {
        t.Errorf("got status %d, want 200", resp.StatusCode)
    }

    reqIDHdr := resp.Header.Get("X-Request-ID")
    if reqIDHdr == "" {
        t.Error("X-Request-ID header not set")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/middleware/... -v`
Expected: FAIL with "undefined: middleware.RequestID"

**Step 3: Write minimal implementation**

Create `api/internal/middleware/request_id.go`:

```go
package middleware

import (
    "context"
    "net/http"

    "github.com/google/uuid"
)

// RequestID middleware adds a unique request ID to context and response header
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqID := uuid.New().String()
        w.Header().Set("X-Request-ID", reqID)

        ctx := context.WithValue(r.Context(), "request_id", reqID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Step 4: Run test to verify it passes**

Run: `cd api && go test ./internal/middleware/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add api/internal/middleware/request_id.go api/internal/middleware/request_id_test.go
git commit -m "feat: add request ID middleware for tracing"
```

---

## Task 2: Create Query Schema Types

**Files:**
- Create: `api/internal/services/query/types.go`
- Test: `api/internal/services/query/types_test.go`

**Step 1: Write the failing test**

Create `api/internal/services/query/types_test.go`:

```go
package query_test

import (
    "testing"

    "github.com/openexposuremanagement/api/internal/services/query"
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
}
```

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/services/query/... -v`
Expected: FAIL with "undefined: query.Query"

**Step 3: Write minimal implementation**

Create `api/internal/services/query/types.go`:

```go
package query

import (
    "errors"
    "strings"
)

// Query represents a JSON query definition
type Query struct {
    Filters       []Filter       `json:"filters"`
    Aggregations  []Aggregation  `json:"aggregations,omitempty"`
    Sort          []Sort         `json:"sort,omitempty"`
    Limit         *int           `json:"limit,omitempty"`
    Offset        *int           `json:"offset,omitempty"`
}

// Filter represents a single filter condition
type Filter struct {
    Field    string      `json:"field"`
    Operator string      `json:"operator"`
    Value    interface{} `json:"value"`
}

// Aggregation represents an aggregation operation
type Aggregation struct {
    Type  string `json:"type"`  // count, sum, max, min, group_by
    Field string `json:"field,omitempty"`
}

// Sort represents sort order
type Sort struct {
    Field string `json:"field"`
    Order string `json:"order"` // asc, desc
}

// Validate performs basic validation on the query
func (q *Query) Validate() error {
    if len(q.Filters) == 0 {
        return errors.New("query must have at least one filter")
    }

    for _, f := range q.Filters {
        if strings.TrimSpace(f.Field) == "" {
            return errors.New("filter field cannot be empty")
        }
        if strings.TrimSpace(f.Operator) == "" {
            return errors.New("filter operator cannot be empty")
        }
    }

    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd api && go test ./internal/services/query/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add api/internal/services/query/types.go api/internal/services/query/types_test.go
git commit -m "feat: add query types with basic validation"
```

---

## Task 3: Implement Query Validator

**Files:**
- Create: `api/internal/services/query/validator.go`
- Test: `api/internal/services/query/validator_test.go`

**Step 1: Write the failing test**

Create `api/internal/services/query/validator_test.go`:

```go
package query_test

import (
    "testing"

    "github.com/openexposuremanagement/api/internal/services/query"
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
}
```

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/services/query/... -v`
Expected: FAIL with "undefined: query.NewValidator"

**Step 3: Write minimal implementation**

Create `api/internal/services/query/validator.go`:

```go
package query

import (
    "errors"
    "fmt"
)

// allowedFields defines whitelisted fields per entity
var allowedFields = map[string][]string{
    "findings": {
        "severity", "scanner_status", "effective_status",
        "cve", "source", "asset_name", "first_observed_at",
        "last_observed_at", "epss_score", "is_kev", "has_cve",
    },
    "assets": {
        "canonical_name", "hostname_norm", "shortname_norm",
        "ipv4", "first_seen_at", "last_seen_at", "is_active",
    },
}

// allowedOperators defines whitelisted operators
var allowedOperators = []string{
    "eq", "neq", "in", "not_in", "like", "not_like",
    "gt", "gte", "lt", "lte", "between", "is_null", "is_not_null",
}

// Validator validates queries against whitelisted fields and operators
type Validator struct{}

// NewValidator creates a new Validator
func NewValidator() *Validator {
    return &Validator{}
}

// Validate checks if the query is valid for the given entity type
func (v *Validator) Validate(entityType string, q *Query) error {
    // Check entity type
    fields, ok := allowedFields[entityType]
    if !ok {
        return fmt.Errorf("invalid entity type: %s", entityType)
    }

    // Validate each filter
    for _, f := range q.Filters {
        // Check field is allowed
        fieldAllowed := false
        for _, allowed := range fields {
            if f.Field == allowed {
                fieldAllowed = true
                break
            }
        }
        if !fieldAllowed {
            return fmt.Errorf("field '%s' not allowed for entity %s", f.Field, entityType)
        }

        // Check operator is allowed
        opAllowed := false
        for _, allowed := range allowedOperators {
            if f.Operator == allowed {
                opAllowed = true
                break
            }
        }
        if !opAllowed {
            return fmt.Errorf("operator '%s' not allowed", f.Operator)
        }
    }

    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd api && go test ./internal/services/query/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add api/internal/services/query/validator.go api/internal/services/query/validator_test.go
git commit -m "feat: add query validator with field/operator whitelisting"
```

---

## Task 4: Implement SQL Translator

**Files:**
- Create: `api/internal/services/query/translator.go`
- Test: `api/internal/services/query/translator_test.go`

**Step 1: Write the failing test**

Create `api/internal/services/query/translator_test.go`:

```go
package query_test

import (
    "testing"

    "github.com/openexposuremanagement/api/internal/services/query"
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
```

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/services/query/... -v`
Expected: FAIL with "undefined: query.NewTranslator"

**Step 3: Write minimal implementation**

Create `api/internal/services/query/translator.go`:

```go
package query

import (
    "fmt"
    "strings"
)

// Translator converts Query objects to SQL
type Translator struct{}

// NewTranslator creates a new Translator
func NewTranslator() *Translator {
    return &Translator{}
}

// Translate converts a query to SQL and returns sql + args
func (tr *Translator) Translate(entityType string, q *Query) (string, []interface{}, error) {
    var whereParts []string
    var args []interface{}
    argPos := 1

    // Build WHERE clause
    for _, f := range q.Filters {
        var part string
        var newArgs []interface{}

        switch f.Operator {
        case "eq":
            part = fmt.Sprintf("%s = $%d", f.Field, argPos)
            newArgs = []interface{}{f.Value}
            argPos++
        case "neq":
            part = fmt.Sprintf("%s != $%d", f.Field, argPos)
            newArgs = []interface{}{f.Value}
            argPos++
        case "in":
            values, ok := f.Value.([]string)
            if !ok {
                return "", nil, fmt.Errorf("in operator requires array of strings")
            }
            placeholders := make([]string, len(values))
            newArgs = make([]interface{}, len(values))
            for i, v := range values {
                placeholders[i] = fmt.Sprintf("$%d", argPos)
                newArgs[i] = v
                argPos++
            }
            part = fmt.Sprintf("%s IN (%s)", f.Field, strings.Join(placeholders, ", "))
        case "like":
            part = fmt.Sprintf("%s LIKE $%d", f.Field, argPos)
            newArgs = []interface{}{f.Value}
            argPos++
        case "gt":
            part = fmt.Sprintf("%s > $%d", f.Field, argPos)
            newArgs = []interface{}{f.Value}
            argPos++
        case "gte":
            part = fmt.Sprintf("%s >= $%d", f.Field, argPos)
            newArgs = []interface{}{f.Value}
            argPos++
        case "lt":
            part = fmt.Sprintf("%s < $%d", f.Field, argPos)
            newArgs = []interface{}{f.Value}
            argPos++
        case "lte":
            part = fmt.Sprintf("%s <= $%d", f.Field, argPos)
            newArgs = []interface{}{f.Value}
            argPos++
        case "is_null":
            part = fmt.Sprintf("%s IS NULL", f.Field)
        case "is_not_null":
            part = fmt.Sprintf("%s IS NOT NULL", f.Field)
        default:
            return "", nil, fmt.Errorf("unsupported operator: %s", f.Operator)
        }

        whereParts = append(whereParts, part)
        args = append(args, newArgs...)
    }

    // Build base query
    var selectClause string
    var groupByClause string
    var orderByClause string
    var limitClause string
    var offsetClause string

    // Handle aggregations
    if len(q.Aggregations) > 0 {
        var selectParts []string
        var groupByFields []string

        for _, agg := range q.Aggregations {
            switch agg.Type {
            case "count":
                if agg.Field != "" {
                    selectParts = append(selectParts, fmt.Sprintf("COUNT(%s)", agg.Field))
                } else {
                    selectParts = append(selectParts, "COUNT(*)")
                }
            case "group_by":
                selectParts = append(selectParts, agg.Field)
                groupByFields = append(groupByFields, agg.Field)
            }
        }
        selectClause = strings.Join(selectParts, ", ")
        if len(groupByFields) > 0 {
            groupByClause = "GROUP BY " + strings.Join(groupByFields, ", ")
        }
    } else {
        selectClause = "*"
    }

    // Handle sort
    if len(q.Sort) > 0 {
        var sortParts []string
        for _, s := range q.Sort {
            sortParts = append(sortParts, fmt.Sprintf("%s %s", s.Field, strings.ToUpper(s.Order)))
        }
        orderByClause = "ORDER BY " + strings.Join(sortParts, ", ")
    }

    // Handle limit
    if q.Limit != nil {
        limitClause = fmt.Sprintf("LIMIT $%d", argPos)
        args = append(args, *q.Limit)
        argPos++
    }

    // Handle offset
    if q.Offset != nil {
        offsetClause = fmt.Sprintf("OFFSET $%d", argPos)
        args = append(args, *q.Offset)
    }

    // Assemble final query
    var queryParts []string
    queryParts = append(queryParts, fmt.Sprintf("SELECT %s FROM %s", selectClause, entityType))
    if len(whereParts) > 0 {
        queryParts = append(queryParts, "WHERE "+strings.Join(whereParts, " AND "))
    }
    if groupByClause != "" {
        queryParts = append(queryParts, groupByClause)
    }
    if orderByClause != "" {
        queryParts = append(queryParts, orderByClause)
    }
    if limitClause != "" {
        queryParts = append(queryParts, limitClause)
    }
    if offsetClause != "" {
        queryParts = append(queryParts, offsetClause)
    }

    return strings.Join(queryParts, " "), args, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd api && go test ./internal/services/query/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add api/internal/services/query/translator.go api/internal/services/query/translator_test.go
git commit -m "feat: add SQL translator for query framework"
```

---

## Task 5: Add Database Migration for Saved Queries

**Files:**
- Create: `api/migrations/000013_saved_queries.up.sql`
- Create: `api/migrations/000013_saved_queries.down.sql`

**Step 1: Create up migration**

Create `api/migrations/000013_saved_queries.up.sql`:

```sql
-- Saved queries table for dashboard widgets and future user-saved queries
CREATE TABLE saved_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('assets', 'findings')),
    query_json JSONB NOT NULL,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    UNIQUE(tenant_id, name)
);

-- Index for tenant lookups
CREATE INDEX idx_saved_queries_tenant ON saved_queries(tenant_id);

-- Index for system queries (used by dashboard)
CREATE INDEX idx_saved_queries_system ON saved_queries(is_system) WHERE is_system = true;

-- Index for entity type lookups
CREATE INDEX idx_saved_queries_entity_type ON saved_queries(tenant_id, entity_type);
```

**Step 2: Create down migration**

Create `api/migrations/000013_saved_queries.down.sql`:

```sql
DROP INDEX IF EXISTS idx_saved_queries_entity_type;
DROP INDEX IF EXISTS idx_saved_queries_system;
DROP INDEX IF EXISTS idx_saved_queries_tenant;
DROP TABLE IF EXISTS saved_queries;
```

**Step 3: Run migration to verify**

Run: `make migrate-up`
Expected: Migration applies successfully

Verify: `psql "$DATABASE_URL" -c "\d saved_queries"`
Expected: Table structure matches migration

**Step 4: Commit**

```bash
git add api/migrations/000013_saved_queries.*
git commit -m "feat: add saved_queries table for query framework"
```

---

## Task 6: Implement Query Executor Service

**Files:**
- Create: `api/internal/services/query/executor.go`
- Create: `api/internal/services/query/executor_integration_test.go`
- Modify: `api/internal/server/server.go` (wire up endpoints later)

**Step 1: Write the failing integration test**

Create `api/internal/services/query/executor_integration_test.go`:

```go
package query_test

import (
    "context"
    "testing"

    "github.com/jmoiron/sqlx"
    "github.com/openexposuremanagement/api/internal/services/query"
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
```

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/services/query/... -v`
Expected: FAIL with "undefined: query.NewExecutor"

**Step 3: Write minimal implementation**

Create `api/internal/services/query/executor.go`:

```go
package query

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/jmoiron/sqlx"
)

// QueryResult represents the result of a query execution
type QueryResult struct {
    Data []map[string]interface{} `json:"data"`
    Meta *QueryMeta                `json:"meta"`
}

// QueryMeta contains metadata about the query execution
type QueryMeta struct {
    TotalRows      int64     `json:"total_rows"`
    ExecutionTimeMs int64    `json:"execution_time_ms"`
    HasMore        bool      `json:"has_more"`
}

// Executor executes validated queries
type Executor struct {
    db        *sqlx.DB
    validator *Validator
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
    if !contains(sql, "WHERE") {
        sql = sql + " WHERE " + tenantFilter
    } else {
        sql = replaceFirst(sql, "WHERE", "WHERE "+tenantFilter+" AND")
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

func contains(s, substr string) bool {
    return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return i
        }
    }
    return -1
}

func replaceFirst(s, old, new string) string {
    idx := indexOf(s, old)
    if idx < 0 {
        return s
    }
    return s[:idx] + new + s[idx+len(old):]
}
```

**Step 4: Run test to verify it passes**

Run: `cd api && go test ./internal/services/query/... -v`
Expected: May PASS or need test data adjustments

**Step 5: Commit**

```bash
git add api/internal/services/query/executor.go api/internal/services/query/executor_integration_test.go
git commit -m "feat: add query executor service"
```

---

## Task 7: Add Structured Error Response Types

**Files:**
- Create: `api/internal/api/errors.go`
- Test: `api/internal/api/errors_test.go`

**Step 1: Write the failing test**

Create `api/internal/api/errors_test.go`:

```go
package api_test

import (
    "encoding/json"
    "net/http/httptest"
    "testing"

    "github.com/openexposuremanagement/api/internal/api"
)

func TestErrorResponse(t *testing.T) {
    t.Run("validation error response", func(t *testing.T) {
        err := &api.QueryError{
            Code:    "VALIDATION_ERROR",
            Message: "Invalid query parameter",
            Details: map[string]interface{}{
                "field": "severity",
                "issue": "must be one of: critical, high, medium, low",
            },
        }

        w := httptest.NewRecorder()
        api.WriteErrorResponse(w, err, "abc-123")

        if w.Code != 400 {
            t.Errorf("got status %d, want 400", w.Code)
        }

        var resp map[string]interface{}
        if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
            t.Fatalf("failed to decode response: %v", err)
        }

        errResp, ok := resp["error"].(map[string]interface{})
        if !ok {
            t.Fatal("response missing error object")
        }

        if errResp["code"] != "VALIDATION_ERROR" {
            t.Errorf("got code %v, want VALIDATION_ERROR", errResp["code"])
        }

        if errResp["request_id"] != "abc-123" {
            t.Errorf("got request_id %v, want abc-123", errResp["request_id"])
        }
    })
}
```

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/api/... -v`
Expected: FAIL with "undefined: api.QueryError"

**Step 3: Write minimal implementation**

Create `api/internal/api/errors.go`:

```go
package api

import (
    "encoding/json"
    "net/http"
    "time"
)

// QueryError represents a structured error response
type QueryError struct {
    Code      string                 `json:"code"`
    Message   string                 `json:"message"`
    Details   map[string]interface{} `json:"details,omitempty"`
    RequestID string                 `json:"request_id"`
    Timestamp string                 `json:"timestamp"`
}

// WriteErrorResponse writes a structured error response
func WriteErrorResponse(w http.ResponseWriter, err *QueryError, requestID string) {
    err.RequestID = requestID
    err.Timestamp = time.Now().Format(time.RFC3339)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)

    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": err,
    })
}
```

**Step 4: Run test to verify it passes**

Run: `cd api && go test ./internal/api/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add api/internal/api/errors.go api/internal/api/errors_test.go
git commit -m "feat: add structured error response types"
```

---

## Task 8: Implement Query HTTP Handlers

**Files:**
- Create: `api/internal/handlers/query.go`
- Test: `api/internal/handlers/query_test.go`

**Step 1: Write the failing test**

Create `api/internal/handlers/query_test.go`:

```go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/openexposuremanagement/api/internal/handlers"
    "github.com/openexposuremanagement/api/internal/services/query"
)

func TestQueryHandler(t *testing.T) {
    t.Run("valid findings query returns 200", func(t *testing.T) {
        executor := &mockExecutor{}
        handler := handlers.NewQueryHandler(executor)

        q := &query.Query{
            Filters: []query.Filter{
                {Field: "severity", Operator: "eq", Value: "critical"},
            },
        }
        body, _ := json.Marshal(q)

        req := httptest.NewRequest("POST", "/api/v1/query/findings", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()

        handler.QueryFindings(w, req)

        if w.Code != 200 {
            t.Errorf("got status %d, want 200", w.Code)
        }
    })

    t.Run("invalid query returns 400", func(t *testing.T) {
        executor := &mockExecutor{}
        handler := handlers.NewQueryHandler(executor)

        body := []byte(`{"invalid": "json"}`)

        req := httptest.NewRequest("POST", "/api/v1/query/findings", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()

        handler.QueryFindings(w, req)

        if w.Code != 400 {
            t.Errorf("got status %d, want 400", w.Code)
        }
    })
}

type mockExecutor struct{}

func (m *mockExecutor) Execute(ctx interface{}, tenantID, entityType string, q *query.Query) (*query.QueryResult, error) {
    return &query.QueryResult{
        Data: []map[string]interface{}{
            {"severity": "critical"},
        },
        Meta: &query.QueryMeta{
            TotalRows:       1,
            ExecutionTimeMs: 5,
            HasMore:         false,
        },
    }, nil
}
```

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/handlers/... -v`
Expected: FAIL with "undefined: handlers.NewQueryHandler"

**Step 3: Write minimal implementation**

Create `api/internal/handlers/query.go`:

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/openexposuremanagement/api/internal/api"
    "github.com/openexposuremanagement/api/internal/services/query"
)

// QueryHandler handles query requests
type QueryHandler struct {
    executor query.ExecutorInterface
}

// ExecutorInterface defines the executor interface for testing
type ExecutorInterface interface {
    Execute(ctx interface{}, tenantID, entityType string, q *query.Query) (*query.QueryResult, error)
}

// NewQueryHandler creates a new QueryHandler
func NewQueryHandler(executor query.ExecutorInterface) *QueryHandler {
    return &QueryHandler{
        executor: executor,
    }
}

// QueryFindings handles POST /api/v1/query/findings
func (h *QueryHandler) QueryFindings(w http.ResponseWriter, r *http.Request) {
    // Parse request body
    var q query.Query
    if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
        api.WriteErrorResponse(w, &api.QueryError{
            Code:    "INVALID_JSON",
            Message: "Failed to parse request body",
        }, getRequestID(r))
        return
    }

    // Get tenant ID from context (would be set by auth middleware)
    tenantID := "demo-tenant"

    // Execute query
    result, err := h.executor.Execute(r.Context(), tenantID, "findings", &q)
    if err != nil {
        api.WriteErrorResponse(w, &api.QueryError{
            Code:    "QUERY_ERROR",
            Message: err.Error(),
        }, getRequestID(r))
        return
    }

    // Return result
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

// QueryAssets handles POST /api/v1/query/assets
func (h *QueryHandler) QueryAssets(w http.ResponseWriter, r *http.Request) {
    // Parse request body
    var q query.Query
    if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
        api.WriteErrorResponse(w, &api.QueryError{
            Code:    "INVALID_JSON",
            Message: "Failed to parse request body",
        }, getRequestID(r))
        return
    }

    // Get tenant ID from context
    tenantID := "demo-tenant"

    // Execute query
    result, err := h.executor.Execute(r.Context(), tenantID, "assets", &q)
    if err != nil {
        api.WriteErrorResponse(w, &api.QueryError{
            Code:    "QUERY_ERROR",
            Message: err.Error(),
        }, getRequestID(r))
        return
    }

    // Return result
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

func getRequestID(r *http.Request) string {
    if reqID := r.Context().Value("request_id"); reqID != nil {
        return reqID.(string)
    }
    return "unknown"
}
```

Also update executor.go to implement the interface:

Add to `api/internal/services/query/executor.go`:

```go
// ExecutorInterface defines the executor interface
type ExecutorInterface interface {
    Execute(ctx context.Context, tenantID, entityType string, q *Query) (*QueryResult, error)
}

// Ensure Executor implements the interface
var _ ExecutorInterface = (*Executor)(nil)
```

**Step 4: Run test to verify it passes**

Run: `cd api && go test ./internal/handlers/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add api/internal/handlers/query.go api/internal/handlers/query_test.go
git add api/internal/services/query/executor.go
git commit -m "feat: add query HTTP handlers"
```

---

## Task 9: Wire Query Endpoints into Router

**Files:**
- Modify: `api/internal/server/server.go`
- Modify: `api/internal/server/routes.go` (if exists)

**Step 1: Update server initialization**

Modify `api/internal/server/server.go` to wire the query handler:

```go
// Add imports
import (
    "github.com/openexposuremanagement/api/internal/handlers"
    "github.com/openexposuremanagement/api/internal/services/query"
)

// In Server struct or setup function
func (s *Server) setupQueryHandler() {
    executor := query.NewExecutor(s.db)
    queryHandler := handlers.NewQueryHandler(executor)

    // Register routes
    s.router.HandleFunc("/api/v1/query/findings", s.withAuth(queryHandler.QueryFindings)).Methods("POST")
    s.router.HandleFunc("/api/v1/query/assets", s.withAuth(queryHandler.QueryAssets)).Methods("POST")
}
```

**Step 2: Test endpoints manually**

Run: `make dev`

Test findings endpoint:
```bash
curl -X POST http://localhost:8080/api/v1/query/findings \
  -H "Content-Type: application/json" \
  -d '{
    "filters": [
      {"field": "severity", "operator": "eq", "value": "critical"}
    ],
    "limit": 10
  }'
```

Expected: JSON response with data array

**Step 3: Commit**

```bash
git add api/internal/server/server.go
git commit -m "feat: wire query endpoints into router"
```

---

## Task 10: Add Query Debug Logging

**Files:**
- Modify: `api/internal/services/query/executor.go`

**Step 1: Add logging to executor**

Update `api/internal/services/query/executor.go` to add debug logging:

```go
import (
    "github.com/rs/zerolog/log"
)

// In Execute method, add logging:
func (e *Executor) Execute(ctx context.Context, tenantID, entityType string, q *Query) (*QueryResult, error) {
    // ... existing validation ...

    // Log the generated SQL (for debugging)
    log.Debug().
        Str("tenant_id", tenantID).
        Str("entity_type", entityType).
        Str("sql", sql).
        Int("arg_count", len(args)).
        Msg("Executing query")

    // ... rest of execution ...

    // Log slow queries
    if executionTime > 1000 { // > 1 second
        log.Warn().
            Str("tenant_id", tenantID).
            Str("entity_type", entityType).
            Int64("execution_time_ms", executionTime).
            Int("row_count", len(data)).
            Msg("Slow query detected")
    }

    // ... return result ...
}
```

**Step 2: Test logging**

Run queries and verify logs appear with `LOG_LEVEL=debug`

**Step 3: Commit**

```bash
git add api/internal/services/query/executor.go
git commit -m "feat: add query debug logging and slow query warnings"
```

---

## Task 11: Add Health Check Endpoints

**Files:**
- Create: `api/internal/handlers/health.go`
- Test: `api/internal/handlers/health_test.go`

**Step 1: Write the failing test**

Create `api/internal/handlers/health_test.go`:

```go
package handlers_test

import (
    "net/http/httptest"
    "testing"

    "github.com/openexposuremanagement/api/internal/handlers"
)

func TestHealthHandler(t *testing.T) {
    handler := handlers.NewHealthHandler(nil)

    t.Run("healthz returns 200", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/healthz", nil)
        w := httptest.NewRecorder()

        handler.HealthZ(w, req)

        if w.Code != 200 {
            t.Errorf("got status %d, want 200", w.Code)
        }
    })

    t.Run("healthz/db returns 200 when db is up", func(t *testing.T) {
        // This would require a real DB connection
        // For now, test handler exists
        req := httptest.NewRequest("GET", "/healthz/db", nil)
        w := httptest.NewRecorder()

        handler.HealthZDB(w, req)

        // Might be 200 or 503 depending on DB state
        if w.Code != 200 && w.Code != 503 {
            t.Errorf("got status %d, want 200 or 503", w.Code)
        }
    })
}
```

**Step 2: Run test to verify it fails**

Run: `cd api && go test ./internal/handlers/... -v`
Expected: FAIL with "undefined: handlers.NewHealthHandler"

**Step 3: Write minimal implementation**

Create `api/internal/handlers/health.go`:

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/jmoiron/sqlx"
)

// HealthHandler handles health check requests
type HealthHandler struct {
    db *sqlx.DB
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(db *sqlx.DB) *HealthHandler {
    return &HealthHandler{db: db}
}

// HealthZ handles GET /healthz
func (h *HealthHandler) HealthZ(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
    })
}

// HealthZDB handles GET /healthz/db
func (h *HealthHandler) HealthZDB(w http.ResponseWriter, r *http.Request) {
    if h.db == nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "error",
            "error":  "database not configured",
        })
        return
    }

    // Try to ping the database
    if err := h.db.Ping(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "error",
            "error":  err.Error(),
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
    })
}
```

**Step 4: Run test to verify it passes**

Run: `cd api && go test ./internal/handlers/... -v`
Expected: PASS

**Step 5: Wire into router**

Update `api/internal/server/server.go`:

```go
func (s *Server) setupHealthHandler() {
    healthHandler := handlers.NewHealthHandler(s.db)

    s.router.HandleFunc("/healthz", healthHandler.HealthZ).Methods("GET")
    s.router.HandleFunc("/healthz/db", healthHandler.HealthZDB).Methods("GET")
}
```

**Step 6: Commit**

```bash
git add api/internal/handlers/health.go api/internal/handlers/health_test.go
git commit -m "feat: add health check endpoints"
```

---

## Final Tasks

### Run Full Test Suite
```bash
cd api && go test ./... -v
cd api && go test ./... -cover
```

### Update Documentation
- Update README.md with new query endpoints
- Add examples of query JSON structures
- Document error response format

### Milestone Complete Checklist
- [ ] Query validator rejects invalid fields/operators
- [ ] SQL translator generates correct SQL for various queries
- [ ] Aggregation queries produce expected GROUP BY clauses
- [ ] Filter operators generate correct WHERE clauses
- [ ] End-to-end query execution with test database
- [ ] Request ID appears in logs and response headers
- [ ] Error responses include request_id and structured details
- [ ] All tests pass with ≥80% coverage

---

**End of Phase 1: Backend Query Framework**

Next phases (Frontend Query Infrastructure, Query Pages, Dashboard Migration, etc.) should be saved to separate plan documents after Phase 1 is complete.
