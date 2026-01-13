# Query Framework & Enhanced Dashboard Design

**Date:** 2025-01-12
**Status:** Approved
**Focus:** Polish and harden MVP - UX/DX improvements and error handling/observability

## Overview

Build a generalized query framework that serves both interactive filtering (UI) and dashboard visualizations. The dashboard becomes a composition of query widgets, and drilldowns pass query parameters to raw query forms. This replaces ad-hoc dashboard endpoints with a flexible, reusable query engine.

**Key Decisions:**
- ✅ Build severity breakdowns and rich detail views
- ✅ Create unified query framework for assets/findings
- ❌ Defer time-series/trend data until OpenSearch implementation
- ✅ Improve error handling with request tracing and structured responses

---

## Architecture

### Core Concept

Dashboards are compositions of queries. The query framework:
1. Accepts JSON query definitions (filters, aggregations, sort, pagination)
2. Validates against whitelisted fields/operators
3. Translates to safe, parameterized SQL
4. Returns structured data with metadata

**Benefits:**
- Single code path for all asset/finding queries
- Dashboards use same framework as interactive queries
- Drilldowns are trivial (just navigate with query params)
- Easy to add new visualizations without backend changes

---

## Backend Design

### JSON Query Schema

```typescript
interface Query {
  filters: Filter[];
  aggregations?: Aggregation[];  // optional - if omitted, return raw rows
  sort?: Sort[];
  limit?: number;
  offset?: number;
}

interface Filter {
  field: string;           // "severity", "asset_name", "has_cve", etc.
  operator: string;        // "eq", "in", "like", "gt", "lt", "between"
  value: any;
}

interface Aggregation {
  type: "count" | "sum" | "max" | "min" | "group_by";
  field?: string;          // required for sum/max/min, optional for count
}

interface Sort {
  field: string;
  order: "asc" | "desc";
}
```

### Example Queries

**Dashboard severity breakdown:**
```json
{
  "filters": [
    {"field": "effective_status", "operator": "eq", "value": "open"}
  ],
  "aggregations": [
    {"type": "group_by", "field": "severity"},
    {"type": "count"}
  ]
}
```

**Top exposed assets:**
```json
{
  "filters": [
    {"field": "effective_status", "operator": "eq", "value": "open"}
  ],
  "aggregations": [
    {"type": "group_by", "field": "asset_name"},
    {"type": "count", "field": "id"},
    {"type": "max", "field": "severity"}
  ],
  "sort": [{"field": "count", "order": "desc"}],
  "limit": 10
}
```

**Findings with CVEs, high severity:**
```json
{
  "filters": [
    {"field": "severity", "operator": "in", "value": ["critical", "high"]},
    {"field": "has_cve", "operator": "eq", "value": true}
  ],
  "sort": [{"field": "last_observed_at", "order": "desc"}],
  "limit": 50
}
```

### Query Executor Service

**File:** `api/internal/services/query/`

**Components:**

1. **Validator** (`validator.go`)
   ```go
   var allowedFields = map[string][]string{
       "findings": {
           "severity", "scanner_status", "effective_status",
           "cve", "source", "asset_name", "first_observed_at",
           "last_observed_at", "epss_score", "is_kev",
       },
       "assets": {
           "canonical_name", "hostname_norm", "first_seen_at",
           "last_seen_at", "is_active",
       },
   }

   var allowedOperators = []string{
       "eq", "neq", "in", "not_in", "like", "not_like",
       "gt", "gte", "lt", "lte", "between", "is_null", "is_not_null",
   }
   ```

2. **SQL Translator** (`translator.go`)
   - Parses validated JSON query
   - Builds WHERE clause from filters (parameterized)
   - Builds GROUP BY / SELECT for aggregations
   - Adds ORDER BY, LIMIT, OFFSET
   - Returns SQL + args

3. **Executor** (`executor.go`)
   - Receives query + tenant context
   - Validates → translates → executes
   - Returns: `{ data: [], meta: { total_rows, execution_time, has_more } }`

### API Endpoints

```
POST /api/v1/query/assets
POST /api/v1/query/findings
GET  /api/v1/query/saved
GET  /api/v1/query/saved/{name}
POST /api/v1/query/saved  (future: user-created queries)
```

**Request/Response:**

```json
// Request
POST /api/v1/query/findings
{
  "filters": [
    {"field": "severity", "operator": "in", "value": ["critical", "high"]}
  ],
  "limit": 50
}

// Response
{
  "data": [
    {
      "id": "123",
      "severity": "critical",
      "asset_name": "web-server-01",
      "cve": "CVE-2024-1234",
      "epss_score": 0.95,
      "is_kev": true,
      "last_observed_at": "2025-01-12T10:00:00Z"
    }
  ],
  "meta": {
    "total_rows": 142,
    "execution_time_ms": 45,
    "has_more": true
  }
}
```

### Error Handling Improvements

**1. Request ID Tracing**
```go
// api/internal/middleware/request_id.go
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqID := uuid.New().String()
        w.Header().Set("X-Request-ID", reqID)
        ctx := context.WithValue(r.Context(), "request_id", reqID)
        log.Logger = log.With().Str("request_id", reqID).Logger()
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**2. Structured Error Response**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid query parameter",
    "details": {
      "field": "severity",
      "issue": "must be one of: critical, high, medium, low"
    },
    "request_id": "abc-123",
    "timestamp": "2025-01-12T10:30:00Z"
  }
}
```

**3. Query Debug Logging**
- Log SQL generated from JSON queries
- Include execution time, row count
- Warn if slow query (> 1s)

**4. Health Endpoints**
- `/healthz` - basic API health
- `/healthz/db` - database connectivity
- `/healthz/queries` - query executor health

---

## Frontend Design

### Component Structure

**1. `useQuery` Hook** - Core query execution
```typescript
const { data, loading, error, meta } = useQuery({
  entity: 'findings' | 'assets',
  query: QueryObject,
  enabled?: boolean,
  refetchInterval?: number
});
```

**2. `QueryBuilder` Component** - Visual filter builder
- Dynamic form based on entity schema
- Add/remove filter groups
- Operator dropdowns (context-aware: "in" shows multi-select for severity)
- Live validation
- Generates valid JSON query

**3. `QueryResultsTable` Component** - Reusable results display
- Sortable columns (click header → adds sort, re-executes)
- Pagination (limit/offset)
- Row click → detail drawer
- Export button (CSV/JSON)
- Column visibility toggle

**4. `QueryWidget` Component** - Dashboard widget wrapper
- Executes query on mount + refresh interval
- Renders visualization based on query type:
  - Aggregation without group_by → stat card
  - Aggregation with group_by → chart or table
  - No aggregation → mini table preview
- On click → navigates to query page with filters pre-loaded

**5. Detail Drawers**
- Slide-out from right
- Shows full record details
- For findings: identifiers, CVE data (NVD/EPSS/KEV), evidence, status history
- For assets: all identifiers, findings summary, seen dates

### Query Pages

**`/findings/query` and `/assets/query`**

Layout:
- Left sidebar: QueryBuilder filters
- Main area: QueryResultsTable
- URL encodes query state (shareable links)

Flow:
1. User builds query with filters
2. Results update automatically (debounced)
3. Click row → detail drawer slides in
4. Click column header → sort
5. Export results → CSV download

### Dashboard Configuration

Dashboards are declarative:

```typescript
// ui/src/dashboards/default.ts
export const defaultDashboard = {
  title: "Overview",
  widgets: [
    {
      id: "severity-breakdown",
      title: "Open Findings by Severity",
      type: "chart",
      query: {
        filters: [
          { field: "effective_status", operator: "eq", value: "open" }
        ],
        aggregations: [
          { type: "group_by", field: "severity" },
          { type: "count" }
        ]
      },
      onClick: {
        navigate: "/findings/query",
        copyFilters: ["severity"]
      }
    },
    {
      id: "top-assets",
      title: "Top Exposed Assets",
      type: "table",
      query: {
        filters: [
          { field: "effective_status", operator: "eq", value: "open" }
        ],
        aggregations: [
          { type: "group_by", field: "asset_name" },
          { type: "count", field: "id" },
          { type: "max", field: "severity" }
        ],
        sort: [{ field: "count", order: "desc" }],
        limit: 10
      },
      onClick: {
        navigate: "/assets/query",
        copyFilters: ["asset_name"]
      }
    },
    {
      id: "intel-coverage",
      title: "Threat Intel Coverage",
      type: "stat",
      query: {
        aggregations: [
          { type: "count" },
          {
            type: "sum",
            field: "case when cve is not null then 1 else 0 end",
            as: "with_cve"
          }
        ]
      }
    }
  ]
};
```

### Error Handling UX

**1. Toast Notifications**
- Non-fatal errors show toast (bottom-right)
- Includes: message, retry button, copy details button

**2. Error Boundary**
- App-level boundary catches component crashes
- Shows friendly error page with request ID
- "Go to dashboard" and "Reload" buttons

**3. Query Error Display**
- Inline errors in QueryBuilder (red border + message)
- Shows exactly which filter is invalid
- Link to docs for valid fields/operators

---

## Database Changes

### Saved Queries Table

```sql
CREATE TABLE saved_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    entity_type VARCHAR(50) NOT NULL, -- 'assets' or 'findings'
    query_json JSONB NOT NULL,
    is_system BOOLEAN DEFAULT FALSE, -- true for dashboard widget queries
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_saved_queries_tenant ON saved_queries(tenant_id);
CREATE INDEX idx_saved_queries_system ON saved_queries(is_system) WHERE is_system = true;
```

### Query Performance Indexes

Add if not already present:

```sql
-- Findings query indexes
CREATE INDEX idx_findings_tenant_severity ON finding_instances(tenant_id, effective_status, severity_id);
CREATE INDEX idx_findings_tenant_asset ON finding_instances(tenant_id, asset_id);
CREATE INDEX idx_findings_tenant_last_observed ON finding_instances(tenant_id, last_observed_at DESC);

-- Asset query indexes
CREATE INDEX idx_assets_tenant_active ON assets(tenant_id, is_active, last_seen_at DESC);
CREATE INDEX idx_asset_identifiers_lookup ON asset_identifiers(tenant_id, id_type, id_value);
```

---

## Implementation Sequence

### Phase 1: Backend Query Framework
- [ ] Query validator (whitelist fields/operators)
- [ ] Query to SQL translator
- [ ] Query executor service
- [ ] Query endpoints (`/api/v1/query/{entity}`)
- [ ] Request ID middleware
- [ ] Structured error responses
- [ ] Tests for query validation and SQL generation
- [ ] Database migration: saved_queries table

### Phase 2: Frontend Query Infrastructure
- [ ] `useQuery` hook with caching (React Query)
- [ ] `QueryBuilder` component
- [ ] `QueryResultsTable` component
- [ ] `QueryWidget` component
- [ ] Error toast system
- [ ] App-level error boundary
- [ ] Detail drawer components

### Phase 3: Query Pages
- [ ] `/findings/query` page (builder + results + detail drawer)
- [ ] `/assets/query` page (builder + results + detail drawer)
- [ ] URL encoding/decoding of query state
- [ ] Export to CSV/JSON functionality

### Phase 4: Dashboard Migration
- [ ] Create default dashboard config
- [ ] Migrate dashboard to use QueryWidget components
- [ ] Add severity breakdown widget
- [ ] Add top exposed assets widget
- [ ] Add Intel coverage widget
- [ ] Implement widget drilldowns to query pages
- [ ] Seed system queries into saved_queries table

### Phase 5: Observability Polish
- [ ] Health check endpoints
- [ ] Query debug logging
- [ ] Request ID display in error UI
- [ ] Slow query warnings
- [ ] Health status indicator in dashboard

### Phase 6: Cleanup
- [ ] Remove old `/v1/dashboard` handler
- [ ] Remove old `/v1/assets` and `/v1/findings` handlers
- [ ] Remove old dashboard component code
- [ ] Update API documentation
- [ ] Update README with new query endpoints

---

## Testing Strategy

### Unit Tests
- Query validator rejects invalid fields/operators
- SQL translator generates correct SQL for various queries
- Aggregation queries produce expected GROUP BY clauses
- Filter operators generate correct WHERE clauses

### Integration Tests
- End-to-end query execution with test database
- Dashboard widget queries return expected aggregations
- Request ID appears in logs and response headers
- Error responses include request_id and structured details

### E2E Tests
- Build query → execute → verify results
- Dashboard widget click → navigate to query page with filters
- Export results → verify CSV format
- Error boundary catches component error

---

## Future Considerations

### Post-MVP (Out of Scope for This Work)
- **Time-series/trend data**: Defer until OpenSearch implementation
- **Saved queries (user-created)**: Allow users to save custom queries
- **Advanced visualizations**: Charts, heatmaps, scatter plots
- **Query sharing**: Share queries via URL between users
- **Query history**: Track recent queries per user

### OpenSearch Migration Path
- Query framework already abstracts SQL generation
- Can swap SQL translator for OpenSearch query DSL translator
- Frontend QueryBuilder remains unchanged
- Dashboard configs work with either backend

---

## Success Criteria

- [ ] Dashboard shows severity breakdown and top assets
- [ ] Query pages allow full filtering/sorting/pagination
- [ ] Drilldown from dashboard widget → query page with pre-populated filters
- [ ] All errors include request_id for debugging
- [ ] Query execution time logged for observability
- [ ] Tests cover query validation, SQL generation, and endpoint behavior
- [ ] Old endpoints removed and codebase simplified
- [ ] Documentation updated with query framework usage
