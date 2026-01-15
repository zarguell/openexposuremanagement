# Unified Query Framework Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enable cross-entity queries (assets + software, assets + findings) with 2-way JOINs to answer complex exposure questions like "assets without CrowdStrike" or "internet-exposed endpoints with exploitable CVEs".

**Architecture:** Extend existing query service (`internal/services/query/`) to support LEFT JOIN between two entities with table aliases, join conditions, and performance guardrails (5,000 row limit, 5-second timeout). Implement unified query endpoint, template library, and UI components for query building and results display.

**Tech Stack:** Go 1.21+ (API), PostgreSQL 16 (JOINs, indexes), React + TypeScript (UI query builder), existing query service framework

---

## Task 1: Design Unified Query API Specification

**Files:**
- Modify: `docs/architecture.md` (append unified query section)

**Step 1: Write unified query JSON schema (documentation only)**

Add to `docs/architecture.md`:

```markdown
## Unified Query API (Cross-Entity Correlation)

### Query JSON Schema
```json
{
  "primary_entity": "assets",
  "join": {
    "entity": "software_inventory",
    "type": "left",
    "on": {
      "primary": "id",
      "joined": "asset_id"
    }
  },
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true}
  ],
  "limit": 100
}
```

### Supported Entity Relationships
- `assets` LEFT JOIN `software_inventory` ON `assets.id = software_inventory.asset_id`
- `assets` LEFT JOIN `findings` ON `assets.id = findings.asset_id`

### Performance Requirements
- Max 5,000 result rows
- 5-second query timeout
- Primary entity filters required
```

**Step 2: Run markdownlint**

```bash
markdownlint docs/architecture.md
```

Expected: PASS (no errors)

**Step 3: Commit**

```bash
git add docs/architecture.md
git commit -m "docs: add unified query API specification"
```

---

## Task 2: Extend Query Types for Join Support

**Files:**
- Modify: `api/internal/services/query/types.go`
- Test: `api/internal/services/query/types_test.go`

**Step 1: Write failing test for Join type**

Create `types_test.go`:

```go
package query

import (
	"testing"
)

func TestJoinValidation(t *testing.T) {
	t.Run("valid left join passes", func(t *testing.T) {
		join := Join{
			Entity: "software_inventory",
			Type:   "left",
			On: JoinCondition{
				Primary: "id",
				Joined:  "asset_id",
			},
		}

		if err := join.Validate(); err != nil {
			t.Errorf("expected valid join, got error: %v", err)
		}
	})

	t.Run("invalid join type fails", func(t *testing.T) {
		join := Join{
			Entity: "software_inventory",
			Type:   "inner",
			On: JoinCondition{
				Primary: "id",
				Joined:  "asset_id",
			},
		}

		if err := join.Validate(); err == nil {
			t.Error("expected error for invalid join type, got nil")
		}
	})

	t.Run("circular reference fails", func(t *testing.T) {
		// This will be validated at query level
		q := &Query{
			Filters: []Filter{},
			Join: &Join{
				Entity: "assets",
				Type:   "left",
			},
		}

		// assets joining to assets should fail
		if err := q.Validate(); err == nil {
			t.Error("expected error for circular reference, got nil")
		}
	})
}
```

**Step 2: Run test to verify it fails**

```bash
cd api && go test ./internal/services/query/... -run TestJoinValidation -v
```

Expected: FAIL with "undefined type Join"

**Step 3: Write minimal implementation - add Join types**

Add to `api/internal/services/query/types.go`:

```go
// JoinCondition defines the ON clause for a join
type JoinCondition struct {
	Primary string `json:"primary"` // Field from primary entity
	Joined  string `json:"joined"`  // Field from joined entity
}

// Join defines a join to another entity
type Join struct {
	Entity string        `json:"entity"` // Entity to join (software_inventory, findings)
	Type   string        `json:"type"`   // "left" only for MVP
	On     JoinCondition `json:"on"`     // Join condition
}

// Validate checks if the join configuration is valid
func (j *Join) Validate() error {
	allowedEntities := map[string]bool{
		"software_inventory": true,
		"findings":           true,
	}

	if !allowedEntities[j.Entity] {
		return fmt.Errorf("unsupported join entity: %s", j.Entity)
	}

	if j.Type != "left" {
		return fmt.Errorf("unsupported join type: %s (only 'left' allowed)", j.Type)
	}

	if j.On.Primary == "" || j.On.Joined == "" {
		return fmt.Errorf("join condition must specify both 'primary' and 'joined' fields")
	}

	return nil
}
```

Also update `Query` struct:

```go
// Query represents a JSON query definition
type Query struct {
	Filters      []Filter      `json:"filters"`
	Aggregations []Aggregation `json:"aggregations,omitempty"`
	Sort         []Sort        `json:"sort,omitempty"`
	Limit        *int          `json:"limit,omitempty"`
	Offset       *int          `json:"offset,omitempty"`
	Join         *Join         `json:"join,omitempty"` // NEW: support 2-way joins
}
```

**Step 4: Run test to verify it passes**

```bash
cd api && go test ./internal/services/query/... -run TestJoinValidation -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add api/internal/services/query/types.go api/internal/services/query/types_test.go
git commit -m "feat: add Join type to query types with validation"
```

---

## Task 3: Extend Query Translator for JOIN SQL Generation

**Files:**
- Modify: `api/internal/services/query/translator.go`
- Test: `api/internal/services/query/translator_test.go`

**Step 1: Write failing test for JOIN SQL generation**

Add to `translator_test.go`:

```go
func TestTranslatorJoins(t *testing.T) {
	t.Run("generates LEFT JOIN SQL", func(t *testing.T) {
		translator := NewTranslator()

		q := &Query{
			Filters: []Filter{
				{Field: "is_active", Operator: "eq", Value: true},
			},
			Join: &Join{
				Entity: "software_inventory",
				Type:   "left",
				On: JoinCondition{
					Primary: "id",
					Joined:  "asset_id",
				},
			},
			Limit: intPtr(100),
		}

		sql, args, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check for LEFT JOIN clause
		if !strings.Contains(sql, "LEFT JOIN software_inventory") {
			t.Errorf("expected LEFT JOIN in SQL, got: %s", sql)
		}

		// Check for ON clause
		if !strings.Contains(sql, "ON assets.id = software_inventory.asset_id") {
			t.Errorf("expected ON clause in SQL, got: %s", sql)
		}

		// Check tenant filter injection
		if !strings.Contains(sql, "WHERE") {
			t.Errorf("expected WHERE clause, got: %s", sql)
		}

		// Verify args
		if len(args) != 2 { // is_active value + limit
			t.Errorf("expected 2 args, got %d: %v", len(args), args)
		}
	})

	t.Run("generates findings join SQL", func(t *testing.T) {
		translator := NewTranslator()

		q := &Query{
			Filters: []Filter{
				{Field: "severity", Operator: "in", Value: []string{"critical", "high"}},
			},
			Join: &Join{
				Entity: "findings",
				Type:   "left",
				On: JoinCondition{
					Primary: "id",
					Joined:  "asset_id",
				},
			},
		}

		sql, _, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(sql, "LEFT JOIN findings") {
			t.Errorf("expected LEFT JOIN findings, got: %s", sql)
		}
	})

	t.Run("handles NULL values from LEFT JOIN correctly", func(t *testing.T) {
		translator := NewTranslator()

		q := &Query{
			Filters: []Filter{
				{Field: "vendor", Operator: "is_null"},
			},
			Join: &Join{
				Entity: "software_inventory",
				Type:   "left",
				On: JoinCondition{
					Primary: "id",
					Joined:  "asset_id",
				},
			},
		}

		sql, _, err := translator.Translate("assets", q)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should use table alias for joined fields
		if !strings.Contains(sql, "software_inventory.vendor IS NULL") {
			t.Errorf("expected table alias for joined field, got: %s", sql)
		}
	})
}

func intPtr(i int) *int {
	return &i
}
```

**Step 2: Run test to verify it fails**

```bash
cd api && go test ./internal/services/query/... -run TestTranslatorJoins -v
```

Expected: FAIL (JOIN not implemented in translator)

**Step 3: Write minimal implementation - extend translator**

Update `Translate` function in `translator.go`:

```go
// Translate converts a validated Query object to parameterized SQL
func (tr *Translator) Translate(entityType string, q *Query) (string, []interface{}, error) {
	var whereParts []string
	var args []interface{}
	argPos := 1

	// Determine primary table and alias
	primaryTable := entityToTable[entityType]
	if primaryTable == "" {
		primaryTable = entityType
	}
	primaryAlias := entityType

	// Build FROM clause with optional JOIN
	var fromClause string
	if q.Join != nil {
		// Add LEFT JOIN with table aliases
		joinedTable := entityToTable[q.Join.Entity]
		if joinedTable == "" {
			joinedTable = q.Join.Entity
		}
		joinedAlias := q.Join.Entity

		fromClause = fmt.Sprintf("%s AS %s LEFT JOIN %s AS %s ON %s.%s = %s.%s",
			primaryTable, primaryAlias,
			joinedTable, joinedAlias,
			primaryAlias, q.Join.On.Primary,
			joinedAlias, q.Join.On.Joined)
	} else {
		fromClause = fmt.Sprintf("%s AS %s", primaryTable, primaryAlias)
	}

	// Build WHERE clause with table aliases
	for _, f := range q.Filters {
		var part string
		var newArgs []interface{}

		// Determine which table this field belongs to
		fieldRef := tr.qualifyField(f.Field, primaryAlias, q.Join)

		switch f.Operator {
		case "eq":
			part = fmt.Sprintf("%s = $%d", fieldRef, argPos)
			newArgs = []interface{}{f.Value}
			argPos++
		case "neq":
			part = fmt.Sprintf("%s != $%d", fieldRef, argPos)
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
			part = fmt.Sprintf("%s IN (%s)", fieldRef, strings.Join(placeholders, ", "))
		// ... (handle other operators similarly with fieldRef)
		default:
			return "", nil, fmt.Errorf("unsupported operator: %s", f.Operator)
		}

		whereParts = append(whereParts, part)
		args = append(args, newArgs...)
	}

	// Build rest of query (aggregations, sort, limit, offset)
	// ... (similar to existing implementation)

	// Assemble final query
	var queryParts []string
	queryParts = append(queryParts, fmt.Sprintf("SELECT * FROM %s", fromClause))
	if len(whereParts) > 0 {
		queryParts = append(queryParts, "WHERE "+strings.Join(whereParts, " AND "))
	}
	// ... (add ORDER BY, LIMIT, OFFSET)

	return strings.Join(queryParts, " "), args, nil
}

// qualifyField adds table alias to field name if needed
func (tr *Translator) qualifyField(field string, primaryAlias string, join *Join) string {
	// Known fields from joined entity
	joinedFields := map[string][]string{
		"software_inventory": {"vendor", "product_name", "version", "cpe_string", "install_path"},
		"findings":           {"severity", "scanner_status", "effective_status", "cve", "epss_score", "is_kev"},
	}

	if join != nil {
		if fields, ok := joinedFields[join.Entity]; ok {
			for _, f := range fields {
				if field == f {
					return fmt.Sprintf("%s.%s", join.Entity, field)
				}
			}
		}
	}

	// Default to primary entity
	return fmt.Sprintf("%s.%s", primaryAlias, field)
}
```

**Step 4: Run test to verify it passes**

```bash
cd api && go test ./internal/services/query/... -run TestTranslatorJoins -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add api/internal/services/query/translator.go api/internal/services/query/translator_test.go
git commit -m "feat: extend translator to generate LEFT JOIN SQL"
```

---

## Task 4: Add Performance Guardrails to Query Executor

**Files:**
- Modify: `api/internal/services/query/executor.go`
- Test: `api/internal/services/query/executor_test.go`

**Step 1: Write failing test for guardrails**

Add to `executor_test.go`:

```go
func TestQueryGuardrails(t *testing.T) {
	t.Run("enforces 5000 row limit", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		executor := NewExecutor(db)

		limit := 6000 // exceeds max
		q := &Query{
			Filters: []Filter{{Field: "is_active", Operator: "eq", Value: true}},
			Limit:   &limit,
		}

		_, err := executor.Execute(context.Background(), "1", "assets", q)
		if err == nil {
			t.Error("expected error for excessive limit, got nil")
		}
	})

	t.Run("enforces 5 second timeout", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		// Slow query that simulates timeout
		executor := NewExecutor(db)

		// This will need a mock or slow DB to test properly
		// For now, test that timeout context is passed
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Query with pg_sleep (if supported)
		q := &Query{
			Filters: []Filter{},
		}

		start := time.Now()
		_, err := executor.Execute(ctx, "1", "assets", q)
		duration := time.Since(start)

		if err == nil && duration > 200*time.Millisecond {
			t.Errorf("query exceeded timeout duration: %v", duration)
		}
	})
}
```

**Step 2: Run test to verify it fails**

```bash
cd api && go test ./internal/services/query/... -run TestQueryGuardrails -v
```

Expected: FAIL (limit is 1000, not 5000)

**Step 3: Write minimal implementation - update executor**

Update `executor.go`:

```go
const (
	MaxQueryLimit   = 5000 // Increased for unified queries
	QueryTimeout    = 5 * time.Second
)

func (e *Executor) Execute(ctx context.Context, tenantID, entityType string, q *Query) (*QueryResult, error) {
	// Validate query
	if err := e.validator.Validate(entityType, q); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Enforce maximum query limit
	if q.Limit != nil && *q.Limit > MaxQueryLimit {
		return nil, fmt.Errorf("limit exceeds maximum of %d", MaxQueryLimit)
	}

	// Add timeout context
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	// ... rest of existing implementation
}
```

**Step 4: Run test to verify it passes**

```bash
cd api && go test ./internal/services/query/... -run TestQueryGuardrails -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add api/internal/services/query/executor.go api/internal/services/query/executor_test.go
git commit -m "feat: add performance guardrails (5000 row limit, 5s timeout)"
```

---

## Task 5: Add Database Indexes for Join Performance

**Files:**
- Create: `db/migrations/000017_add_unified_query_indexes.up.sql`
- Create: `db/migrations/000017_add_unified_query_indexes.down.sql`

**Step 1: Write migration (no test needed, verified by EXPLAIN)**

Create `db/migrations/000017_add_unified_query_indexes.up.sql`:

```sql
-- +migrate Up
-- Add indexes for unified query JOIN performance

-- Composite index for assets LEFT JOIN software_inventory
-- Covers: SELECT * FROM assets LEFT JOIN software_inventory ON assets.id = software_inventory.asset_id
CREATE INDEX idx_asset_software_join_covering ON asset_software(tenant_id, asset_id, software_id)
INCLUDE (last_seen_at);

-- Composite index for assets LEFT JOIN findings
-- Covers: SELECT * FROM assets LEFT JOIN findings ON assets.id = findings.asset_id
CREATE INDEX idx_finding_instances_asset_join ON finding_instances(tenant_id, asset_id, definition_uid)
INCLUDE (effective_status, last_observed_at);

-- Index for software vendor/product queries (common filter)
CREATE INDEX idx_software_vendor_product_lookup ON software(vendor, product_name, version);
```

Create `db/migrations/000017_add_unified_query_indexes.down.sql`:

```sql
-- +migrate Down
DROP INDEX IF EXISTS idx_asset_software_join_covering;
DROP INDEX IF EXISTS idx_finding_instances_asset_join;
DROP INDEX IF EXISTS idx_software_vendor_product_lookup;
```

**Step 2: Run migration**

```bash
make migrate-up
```

Expected: Migration succeeds

**Step 3: Verify with EXPLAIN ANALYZE**

```bash
psql "$DATABASE_URL" -c "EXPLAIN ANALYZE SELECT * FROM assets LEFT JOIN software_inventory ON assets.id = software_inventory.asset_id WHERE assets.tenant_id = 1 LIMIT 100;"
```

Expected: Index shows in query plan

**Step 4: Commit**

```bash
git add db/migrations/000017_add_unified_query_indexes.*
git commit -m "feat: add indexes for unified query JOIN performance"
```

---

## Task 6: Implement POST /api/v1/query/unified Endpoint

**Files:**
- Modify: `api/internal/handlers/query.go`
- Test: `api/internal/handlers/query_test.go`

**Step 1: Write failing test for unified endpoint**

Add to `query_test.go`:

```go
func TestQueryUnified(t *testing.T) {
	t.Run("valid join query returns results", func(t *testing.T) {
		executor := &mockQueryExecutor{
			result: &query.QueryResult{
				Data: []map[string]interface{}{
					{
						"id":             int64(1),
						"canonical_name": "server1.local",
						"vendor":         "Microsoft",
						"product_name":   "Windows Server",
					},
				},
				Meta: &query.QueryMeta{
					TotalRows:       1,
					ExecutionTimeMs: 25,
					HasMore:         false,
				},
			},
		}

		handler := NewQueryHandler(executor)

		body := `{
			"primary_entity": "assets",
			"join": {
				"entity": "software_inventory",
				"type": "left",
				"on": {"primary": "id", "joined": "asset_id"}
			},
			"filters": [
				{"field": "vendor", "operator": "eq", "value": "Microsoft"}
			],
			"limit": 100
		}`

		req := httptest.NewRequest("POST", "/api/v1/query/unified", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req = setQueryUserContextWithRequestID(req, 1, "test-req-unified")

		w := httptest.NewRecorder()
		handler.QueryUnified(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want 200", w.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["data"] == nil {
			t.Fatal("expected data field in response")
		}
	})
}
```

**Step 2: Run test to verify it fails**

```bash
cd api && go test ./internal/handlers/... -run TestQueryUnified -v
```

Expected: FAIL (method doesn't exist)

**Step 3: Write minimal implementation - add unified handler**

Add to `query.go`:

```go
// QueryUnified handles POST /api/v1/query/unified
func (h *QueryHandler) QueryUnified(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)

	// Get user context
	userCtx := r.Context().Value(auth.UserContextKey)
	if userCtx == nil {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "UNAUTHORIZED",
			Message: "User context not found",
		}, requestID, http.StatusUnauthorized)
		return
	}

	user, ok := userCtx.(*auth.UserContext)
	if !ok {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "INVALID_CONTEXT",
			Message: "Invalid user context",
		}, requestID, http.StatusInternalServerError)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "INVALID_REQUEST",
			Message: "Failed to read request body",
			Details: map[string]interface{}{"error": err.Error()},
		}, requestID, http.StatusBadRequest)
		return
	}

	// Parse unified query
	var unifiedQuery struct {
		PrimaryEntity string         `json:"primary_entity"`
		Filters       []query.Filter `json:"filters"`
		Join          *query.Join    `json:"join"`
		Aggregations  []query.Aggregation `json:"aggregations,omitempty"`
		Sort          []query.Sort   `json:"sort,omitempty"`
		Limit         *int           `json:"limit,omitempty"`
		Offset        *int           `json:"offset,omitempty"`
	}

	if err := json.Unmarshal(body, &unifiedQuery); err != nil {
		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "INVALID_JSON",
			Message: "Invalid JSON in request body",
			Details: map[string]interface{}{"error": err.Error()},
		}, requestID, http.StatusBadRequest)
		return
	}

	// Build Query object
	q := &query.Query{
		Filters:      unifiedQuery.Filters,
		Join:         unifiedQuery.Join,
		Aggregations: unifiedQuery.Aggregations,
		Sort:         unifiedQuery.Sort,
		Limit:        unifiedQuery.Limit,
		Offset:       unifiedQuery.Offset,
	}

	// Execute query
	tenantID := strconv.FormatInt(user.TenantID, 10)
	result, err := h.executor.Execute(r.Context(), tenantID, unifiedQuery.PrimaryEntity, q)
	if err != nil {
		log.Error().Err(err).
			Str("request_id", requestID).
			Str("tenant_id", tenantID).
			Str("primary_entity", unifiedQuery.PrimaryEntity).
			Msg("unified query execution failed")

		errMsg := err.Error()
		if strings.HasPrefix(errMsg, "validation error:") {
			api.WriteErrorResponse(w, &api.QueryError{
				Code:    "VALIDATION_ERROR",
				Message: "Query validation failed",
				Details: map[string]interface{}{"error": errMsg},
			}, requestID, http.StatusUnprocessableEntity)
			return
		}

		api.WriteErrorResponse(w, &api.QueryError{
			Code:    "QUERY_FAILED",
			Message: "Query execution failed",
			Details: map[string]interface{}{"error": err.Error()},
		}, requestID, http.StatusInternalServerError)
		return
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": result.Data,
		"meta": result.Meta,
	})); err != nil {
		log.Error().Err(err).
			Str("request_id", requestID).
			Msg("failed to encode unified query response")
	}
}
```

Also need to update the routing in `api/internal/server/server.go` to add the route.

**Step 4: Run test to verify it passes**

```bash
cd api && go test ./internal/handlers/... -run TestQueryUnified -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add api/internal/handlers/query.go api/internal/handlers/query_test.go
git commit -m "feat: add POST /api/v1/query/unified endpoint"
```

---

## Task 7: Create Unified Query Templates

**Files:**
- Create: `api/internal/services/query/templates.go`
- Test: `api/internal/services/query/templates_test.go`

**Step 1: Write failing test for templates**

Create `templates_test.go`:

```go
package query

import (
	"testing"
)

func TestQueryTemplates(t *testing.T) {
	t.Run("load missing software template", func(t *testing.T) {
		tmpl := GetTemplate("missing_software")
		if tmpl == nil {
			t.Fatal("expected template, got nil")
		}

		if tmpl.Name != "Assets Missing Critical Software" {
			t.Errorf("got name %s, want 'Assets Missing Critical Software'", tmpl.Name)
		}
	})

	t.Run("load exploitable CVEs template", func(t *testing.T) {
		tmpl := GetTemplate("exploitable_cves")
		if tmpl == nil {
			t.Fatal("expected template, got nil")
		}

		if tmpl.PrimaryEntity != "assets" {
			t.Errorf("got primary_entity %s, want 'assets'", tmpl.PrimaryEntity)
		}

		if tmpl.Join == nil {
			t.Error("expected join in exploitable CVEs template")
		}
	})
}
```

**Step 2: Run test to verify it fails**

```bash
cd api && go test ./internal/services/query/... -run TestQueryTemplates -v
```

Expected: FAIL (templates don't exist)

**Step 3: Write minimal implementation**

Create `templates.go`:

```go
package query

// Template represents a pre-built query template
type Template struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	PrimaryEntity string   `json:"primary_entity"`
	Join          *Join    `json:"join,omitempty"`
	Filters       []Filter `json:"filters"`
	Parameters    []Parameter `json:"parameters"`
}

// Parameter defines a template parameter
type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // string, array, number
	Description string `json:"description"`
	Default     interface{} `json:"default,omitempty"`
}

// templates is the template registry
var templates = map[string]*Template{
	"missing_software": {
		ID:            "missing_software",
		Name:          "Assets Missing Critical Software",
		Description:   "Find assets that do not have a specific software installed (e.g., CrowdStrike, antivirus)",
		PrimaryEntity: "assets",
		Join: &Join{
			Entity: "software_inventory",
			Type:   "left",
			On: JoinCondition{
				Primary: "id",
				Joined:  "asset_id",
			},
		},
		Filters: []Filter{
			{Field: "product_name", Operator: "eq", Value: "{{software_name}}"},
		},
		Parameters: []Parameter{
			{
				Name:        "software_name",
				Type:        "string",
				Description: "Name of the software to check (e.g., 'CrowdStrike Falcon')",
				Default:     "CrowdStrike Falcon",
			},
		},
	},
	"exploitable_cves": {
		ID:            "exploitable_cves",
		Name:          "Assets with Exploitable CVEs",
		Description:   "Find assets that have CVEs with high EPSS scores or in CISA KEV",
		PrimaryEntity: "assets",
		Join: &Join{
			Entity: "findings",
			Type:   "left",
			On: JoinCondition{
				Primary: "id",
				Joined:  "asset_id",
			},
		},
		Filters: []Filter{
			{Field: "epss_score", Operator: "gte", Value: "{{epss_threshold}}"},
		},
		Parameters: []Parameter{
			{
				Name:        "epss_threshold",
				Type:        "number",
				Description: "Minimum EPSS score (0.0-1.0)",
				Default:     0.9,
			},
		},
	},
	"software_vulnerabilities": {
		ID:            "software_vulnerabilities",
		Name:          "Vulnerabilities by Software",
		Description:   "Find findings affecting specific software products",
		PrimaryEntity: "findings",
		Join:          nil, // findings already joined to assets via view
		Filters: []Filter{
			{Field: "cve", Operator: "is_not_null"},
		},
		Parameters: []Parameter{
			{
				Name:        "vendor",
				Type:        "string",
				Description: "Software vendor (e.g., 'Microsoft')",
			},
			{
				Name:        "product",
				Type:        "string",
				Description: "Software product (e.g., 'Windows Server')",
			},
		},
	},
}

// GetTemplate retrieves a template by ID
func GetTemplate(id string) *Template {
	return templates[id]
}

// ListTemplates returns all available templates
func ListTemplates() []*Template {
	result := make([]*Template, 0, len(templates))
	for _, tmpl := range templates {
		result = append(result, tmpl)
	}
	return result
}
```

**Step 4: Run test to verify it passes**

```bash
cd api && go test ./internal/services/query/... -run TestQueryTemplates -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add api/internal/services/query/templates.go api/internal/services/query/templates_test.go
git commit -m "feat: add unified query templates library"
```

---

## Task 8: Add Unified Query Documentation and Examples

**Files:**
- Create: `docs/unified-queries.md`

**Step 1: Write documentation**

Create `docs/unified-queries.md`:

```markdown
# Unified Queries

## Overview

The unified query API enables cross-entity correlation queries using 2-way LEFT JOINs between assets, software inventory, and findings.

## API Endpoint

```
POST /api/v1/query/unified
```

## Query Syntax

### Basic Structure

```json
{
  "primary_entity": "assets",
  "join": {
    "entity": "software_inventory",
    "type": "left",
    "on": {
      "primary": "id",
      "joined": "asset_id"
    }
  },
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true}
  ],
  "limit": 100
}
```

### Supported Relationships

| Primary Entity | Join Entity | On Condition |
|----------------|-------------|--------------|
| assets | software_inventory | assets.id = software_inventory.asset_id |
| assets | findings | assets.id = findings.asset_id |

### Common Use Cases

#### Assets Missing Specific Software

```json
{
  "primary_entity": "assets",
  "join": {
    "entity": "software_inventory",
    "type": "left",
    "on": {"primary": "id", "joined": "asset_id"}
  },
  "filters": [
    {"field": "product_name", "operator": "eq", "value": "CrowdStrike Falcon"}
  ],
  "limit": 100
}
```

Returns assets where `product_name` is NULL (software not installed).

#### Assets with Exploitable CVEs

```json
{
  "primary_entity": "assets",
  "join": {
    "entity": "findings",
    "type": "left",
    "on": {"primary": "id", "joined": "asset_id"}
  },
  "filters": [
    {"field": "epss_score", "operator": "gte", "value": 0.9}
  ],
  "limit": 100
}
```

#### Software Vulnerability Correlation

```json
{
  "primary_entity": "findings",
  "filters": [
    {"field": "cve", "operator": "eq", "value": "CVE-2021-44228"}
  ],
  "limit": 50
}
```

Findings view already includes asset information.

## Performance Guidelines

- **Max rows**: 5,000 (enforced)
- **Timeout**: 5 seconds
- **Join type**: LEFT JOIN only (prevents data explosion)
- **Filters**: Required on primary entity for performance
- **Indexes**: Composite indexes support common join patterns

## Limitations

- 2-way joins only (no multi-hop joins in MVP)
- LEFT JOIN only (no INNER, RIGHT, CROSS)
- Filter fields must belong to primary or joined entity
- Aggregations not yet supported on joined fields

## Templates

The API provides pre-built query templates:

| Template ID | Name | Description |
|-------------|------|-------------|
| `missing_software` | Assets Missing Critical Software | Find assets without specific software |
| `exploitable_cves` | Assets with Exploitable CVEs | High EPSS or CISA KEV |
| `software_vulnerabilities` | Vulnerabilities by Software | Findings affecting specific products |

Use `GET /api/v1/query/templates` to list all templates.
```

**Step 2: Run markdownlint**

```bash
markdownlint docs/unified-queries.md
```

Expected: PASS

**Step 3: Commit**

```bash
git add docs/unified-queries.md
git commit -m "docs: add unified query documentation"
```

---

## Task 9: Milestone Refactor & Optimization Pass

**Files:**
- Modify: Multiple (refactoring)
- Create: `db/migrations/000018_add_unified_query_materialized_views.up.sql`

**Step 1: Review query performance**

```bash
# Test dashboard widget queries
make demo-smoke
ab -n 100 -c 10 http://localhost:8080/api/v1/query/unified
```

Expected: All requests < 2 seconds

**Step 2: Create materialized views for slow widgets**

If needed, create migration:

```sql
-- +migrate Up
-- Materialized views for dashboard widgets

CREATE MATERIALIZED VIEW mv_assets_missing_software AS
SELECT DISTINCT a.id, a.tenant_id, a.canonical_name
FROM assets a
WHERE a.is_active = TRUE
  AND NOT EXISTS (
    SELECT 1 FROM asset_software asw
    JOIN software s ON s.id = asw.software_id
    WHERE asw.asset_id = a.id
  );

CREATE UNIQUE INDEX idx_mv_assets_missing_software ON mv_assets_missing_software(id);
CREATE INDEX idx_mv_assets_missing_software_tenant ON mv_assets_missing_software(tenant_id);
```

**Step 3: Add query logging**

Update `executor.go` to log slow queries:

```go
if executionTime > 1000 { // > 1 second
	log.Warn().
		Str("tenant_id", tenantID).
		Str("entity_type", entityType).
		Int64("duration_ms", executionTime).
		Msg("slow query detected")
}
```

**Step 4: Update documentation**

Add limitations section to `docs/unified-queries.md`:

```markdown
## Known Limitations (MVP)

- No N-way joins (3+ entities)
- No aggregations on joined fields
- No subquery pushdown optimization
- Template parameter substitution not yet implemented
- UI query builder not yet implemented (backend-only API)
```

**Step 5: Run smoke test**

```bash
make demo-smoke
```

Expected: PASS

**Step 6: Commit**

```bash
git add .
git commit -m "refactor: optimize unified queries and add materialized views"
```

---

## Execution Handoff

**Plan complete and saved to `docs/plans/2025-01-14-unified-query-framework.md`. Two execution options:**

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**

---

## Summary

This plan implements the **Unified Query Framework** milestone with:

- **API Design**: JSON schema for 2-way LEFT JOIN queries
- **Backend**: Extended query service with JOIN support
- **Guardrails**: 5,000 row limit, 5-second timeout
- **Templates**: Pre-built queries for common scenarios
- **Documentation**: Complete API reference and examples
- **Optimization**: Database indexes and materialized views

**Backend tasks (9 total)**: TDD approach, commit after each step, validation commands for quality.

**UI tasks (deferred)**: Query builder, results display, dashboard widgets - not in this plan.

**Estimated time**: ~4-6 hours for backend implementation following TDD cycle.
