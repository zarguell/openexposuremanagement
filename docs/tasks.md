# OpenExposureManagement Tasks

**🚨 CURRENT FOCUS: Unified Query Framework (Cross-Entity Correlation)**
***

## Milestone: Unified Query Framework (cross-entity correlation)
**🚨 IMMEDIATE PRIORITY - This is the next development focus**

Enable cross-entity queries to answer complex exposure questions like "assets without CrowdStrike in prod" or "internet-exposed endpoints with exploitable CVEs". This milestone implements 2-way JOIN queries (assets + software, assets + findings) with performance guardrails and a unified query builder UI.

### Use Cases
- Assets without specific software (e.g., "missing CrowdStrike")
- Assets with exploitable vulnerabilities (e.g., "critical CVEs in DMZ")
- Cross-entity correlation (e.g., "software X has vulnerability Y")
- Dashboard widgets showing correlated risk data

### Performance Guardrails
- 2-way LEFT JOIN only (no CROSS JOIN explosion)
- Hard limit: 5,000 result rows
- Query timeout: 5 seconds
- Require filters on primary entity first
- Subquery pushdown for filtering

 - [ ] Task: Design unified query API specification
   - **Description:** Define JSON schema for unified queries supporting 2-way joins between entities (assets + software, assets + findings). Specify join types, relationship definitions, and result aggregation patterns.
   - **Acceptance criteria:** Complete API spec with examples for all use cases; documented in `docs/architecture.md`; performance requirements specified (limits, timeouts).
   - **Validation command:** `markdownlint docs/architecture.md`
   - **Dependencies:** Software Inventory milestone complete
   - **Estimated tokens:** 2200

 - [ ] Task: Extend query types for join support
   - **Description:** Add `Join` type to `internal/services/query/types.go` with entity relationship definitions, join conditions (ON clause), and join type restrictions (LEFT JOIN only for MVP).
   - **Acceptance criteria:** Type definitions support 2-way joins; validation prevents circular references; tests cover valid/invalid join scenarios.
   - **Validation command:** `go test ./internal/services/query/...`
   - **Dependencies:** Design unified query API specification
   - **Estimated tokens:** 2800

 - [ ] Task: Extend query translator for JOIN SQL generation
   - **Description:** Update `internal/services/query/translator.go` to generate LEFT JOIN SQL from unified query JSON, including table aliases, join conditions, and proper WHERE clause placement.
   - **Acceptance criteria:** Generates valid parameterized SQL for 2-way joins; handles NULL checks correctly; unit tests verify SQL output.
   - **Validation command:** `go test ./internal/services/query/... -run TestTranslatorJoins`
   - **Dependencies:** Extend query types for join support
   - **Estimated tokens:** 3200

 - [ ] Task: Add performance guardrails to query executor
   - **Description:** Implement query result limits (5,000 rows), timeout enforcement (5 seconds), and query cost estimation in `internal/services/query/executor.go`. Add early termination and logging.
   - **Acceptance criteria:** Queries exceeding limits return error; timeout enforced; slow queries logged; metrics recorded for monitoring.
   - **Validation command:** `go test ./internal/services/query/... -run TestQueryGuardrails`
   - **Dependencies:** Extend query translator for JOIN SQL generation
   - **Estimated tokens:** 2600

 - [x] Task: Add database indexes for join performance
   - **Description:** Create migration adding composite indexes on foreign keys (asset_software.asset_id, finding_instances.asset_id) and covering indexes for common join patterns.
   - **Acceptance criteria:** Migration runs successfully; indexes used in query plans (verified via EXPLAIN ANALYZE); query performance acceptable.
   - **Validation command:** `make migrate-up && psql "$DATABASE_URL" -c "EXPLAIN ANALYZE SELECT ..."`
   - **Dependencies:** Extend query types for join support
   - **Estimated tokens:** 2400

 - [ ] Task: Implement POST /api/v1/query/unified endpoint
   - **Description:** Add unified query handler in `internal/handlers/query.go` accepting 2-way join queries, returning correlated results with entity metadata.
   - **Acceptance criteria:** Endpoint accepts valid unified query JSON; returns results with proper entity context; tenant scoping enforced; integration tests pass.
   - **Validation command:** `go test ./... && curl -sf http://localhost:8080/api/v1/query/unified`
   - **Dependencies:** Add performance guardrails to query executor
   - **Estimated tokens:** 3000

 - [ ] Task: Create unified query templates
   - **Description:** Implement query template library in `internal/services/query/templates.go` with pre-built queries for common scenarios (missing software, exploitable CVEs, software vulnerability correlation).
   - **Acceptance criteria:** Templates cover major use cases; templates are parameterized; can be loaded and executed via API.
   - **Validation command:** `go test ./internal/services/query/... -run TestQueryTemplates`
   - **Dependencies:** Implement POST /api/v1/query/unified endpoint
   - **Estimated tokens:** 2400

 - [ ] Task: Extend UI query builder for unified queries
   - **Description:** Add "Unified Query" mode to `ui/src/components/QueryBuilder.tsx` supporting entity selection, relationship builder, and multi-entity filter composition.
   - **Acceptance criteria:** UI supports building 2-way join queries; visual feedback on relationships; validates query before submission; real-time error messages.
   - **Validation command:** `npm run test:unit`
   - **Dependencies:** Implement POST /api/v1/query/unified endpoint
   - **Estimated tokens:** 4000

 - [ ] Task: Create unified query results display component
   - **Description:** Build `ui/src/components/UnifiedQueryResults.tsx` to display correlated results with entity grouping, drill-down to individual entities, and export capabilities.
   - **Acceptance criteria:** Results show joined data clearly; supports drill-down to asset/software/finding details; handles NULL values appropriately; export to CSV works.
   - **Validation command:** `npm run test:unit`
   - **Dependencies:** Extend UI query builder for unified queries
   - **Estimated tokens:** 3200

 - [ ] Task: Add unified query page with template library
   - **Description:** Create `ui/src/pages/UnifiedQueries.tsx` page with query template gallery, template one-click execution, and custom query builder interface.
   - **Acceptance criteria:** Page accessible from nav; templates listed with descriptions; one-click execution works; custom query builder functional.
   - **Validation command:** `npm run test:e2e`
   - **Dependencies:** Create unified query results display component; Create unified query templates
   - **Estimated tokens:** 2800

 - [ ] Task: Update default dashboard with unified query widgets
   - **Description:** Replace static dashboard widgets with unified query results (e.g., "Assets Missing Critical Software", "Assets with Exploitable CVEs"). Update `ui/src/pages/Dashboard.tsx`.
   - **Acceptance criteria:** Dashboard widgets use `/api/v1/query/unified`; widgets show correlated data; real-time refresh works; performance acceptable.
   - **Validation command:** `npm run test:e2e`
   - **Dependencies:** Add unified query page with template library
   - **Estimated tokens:** 2400

 - [ ] Task: Add unified query documentation and examples
   - **Description:** Document unified query syntax, templates, performance considerations, and common query patterns in `docs/unified-queries.md`.
   - **Acceptance criteria:** Examples cover all use cases; performance guidelines documented; limitations clearly stated; docs pass lint.
   - **Validation command:** `markdownlint docs/unified-queries.md`
   - **Dependencies:** Update default dashboard with unified query widgets
   - **Estimated tokens:** 2000

 - [ ] Task: Milestone refactor & optimization pass
   - **Description:** Review query performance, add subquery pushdown optimization, create materialized views for slow widget queries, and document known limitations.
   - **Acceptance criteria:** Dashboard widgets load in <2 seconds; slow queries identified and optimized; migration for materialized views created; docs updated.
   - **Validation command:** `make demo-smoke && ab -n 100 http://localhost:8080/api/v1/query/unified`
   - **Dependencies:** All tasks in "Unified Query Framework"
   - **Estimated tokens:** 2600



***

## ~~Milestone: Suppressions (proposal/approval + recompute)~~
**MOVED TO POST-MVP**

Implement CVE-level suppression proposal flow and async effective-status recompute; reset context after this milestone.

- [ ] Task: Implement suppression proposal endpoint
  - **Description:** Add `POST /suppressions/proposals` for analyst/admin to propose CVE suppressions.
  - **Acceptance criteria:** Validates `match_type=cve`; stores proposal in pending state; writes review entry.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Go API foundation → Implement RBAC middleware (admin/analyst/viewer); Database schema & migrations → Create suppression workflow tables
  - **Estimated tokens:** 2600

- [ ] Task: Implement approve/reject/revoke endpoints (admin)
  - **Description:** Add admin-only endpoints to approve, reject, and revoke suppressions with audit trail.
  - **Acceptance criteria:** State transitions enforced; `tenant_policy_state.policy_revision` increments on approve/revoke.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement suppression proposal endpoint
  - **Estimated tokens:** 2900

- [ ] Task: Implement async recompute worker (stale effective_revision)
  - **Description:** Background job recomputes effective status for `finding_instances` where `effective_revision < policy_revision` in batches.
  - **Acceptance criteria:** Recompute is idempotent; bounded batch size; visible progress/logging.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement approve/reject/revoke endpoints (admin); VM ingestion pipeline → Compute effective status on write (baseline)
  - **Estimated tokens:** 3000

- [ ] Task: Implement suppression listing endpoints (for UI)
  - **Description:** Add endpoints to list proposals/approved suppressions and show audit trail entries.
  - **Acceptance criteria:** Tenant scoping enforced; includes state, dates, reason, and review actions.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement approve/reject/revoke endpoints (admin)
  - **Estimated tokens:** 2400

- [x] Task: Milestone refactor & duplication pass
   - **Description:** Deduplicate filter parsing, pagination structs, and SQL fragments; ensure consistent response envelopes.
   - **Acceptance criteria:** `go test ./...` passes; lint passes; no duplicated query-building logic.
   - **Validation command:** `go test ./...`
   - **Dependencies:** All tasks in “Query APIs (assets/findings/dashboard)”
   - **Estimated tokens:** 1700

***

## Milestone: Demo data, ops docs, and demo checklist
Make the demo easy to run and verify end-to-end; reset context after this milestone.

- [x] Task: Add sample ingestion payloads + seeding script
  - **Description:** Provide sample Tenable/Qualys-like JSON payloads and a script to POST them using an ingestion API key.
  - **Acceptance criteria:** Running the script populates assets and findings; repeat runs are idempotent.
  - **Validation command:** `make seed && make demo-smoke`
  - **Dependencies:** VM ingestion pipeline → Implement POST /ingest/vm/findings end-to-end
  - **Estimated tokens:** 2600

- [x] Task: Add end-to-end “demo smoke test” script
  - **Description:** Script that brings up stack, ingests data, refreshes views, and hits key endpoints (`/assets`, `/findings`, `/intel/status`).
  - **Acceptance criteria:** Script exits non-zero on failure; produces minimal readable output for demo operator.
  - **Validation command:** `make demo-smoke`
  - **Dependencies:** Add sample ingestion payloads + seeding script; Query APIs milestone
  - **Estimated tokens:** 2400

- [ ] Task: Document backup/restore (volume snapshot) and MVP runbook
   - **Description:** Write docs for starting stack, configuring IdP, creating API keys, backup/restore of docker volumes, and common troubleshooting.
   - **Acceptance criteria:** A new developer can run demo from scratch following docs; includes "demo key dev-only" warning.
   - **Validation command:** `markdownlint docs/*.md` (or `make docs-check`)
   - **Dependencies:** Repo & dev baseline → Add Makefile/dev scripts
   - **Estimated tokens:** 2000

 - [ ] Task: Milestone refactor & duplication pass
   - **Description:** Review scripts/docs for duplication and drift; ensure commands match Makefile and compose.
   - **Acceptance criteria:** `make demo-smoke` works; docs commands are consistent; lint passes.
   - **Validation command:** `make demo-smoke`
   - **Dependencies:** All tasks in "Demo data, ops docs, and demo checklist"
   - **Estimated tokens:** 1500

***

## Future Enhancements (Stretch Goals)

These features are planned for post-MVP development and are not currently prioritized.

### Network Context & Asset Attributes
- [ ] Add `internet_facing`, `environment` (prod/dev/staging), and `subnet` fields to assets table
- [ ] Create network segmentation views (DMZ, internal, internet-facing)
- [ ] Implement network-based correlation queries
- **Rationale:** Enable queries like "internet-exposed endpoints with critical vulnerabilities"

### Custom Dashboard Builder
- [ ] Dashboard CRUD API (save/load/delete per-user)
- [ ] Drag-and-drop widget layout builder
- [ ] Widget library with unified query components
- [ ] Dashboard sharing and export/import
- **Storage:** Consider NoSQL database (e.g., PostgreSQL JSONB) or file-based configs
- **Rationale:** Allow users to create custom exposure dashboards

### Advanced Query Features
- [ ] N-way JOIN support (3+ entity joins)
- [ ] Subquery pushdown optimization
- [ ] Query result caching (5-minute TTL)
- [ ] Materialized views for common join patterns
- [ ] Query cost estimation and optimization hints
- **Rationale:** Support complex correlation queries at scale

### Real-time Updates & Streaming
- [ ] WebSocket-based real-time dashboard updates
- [ ] Streaming query results for large datasets
- [ ] Change data capture (CDC) for auto-refresh
- **Rationale:** Eliminate manual refresh, show live exposure data

- [ ] Task: Milestone refactor & duplication pass
  - **Description:** Review scripts/docs for duplication and drift; ensure commands match Makefile and compose.
  - **Acceptance criteria:** `make demo-smoke` works; docs commands are consistent; lint passes.
  - **Validation command:** `make demo-smoke`
  - **Dependencies:** All tasks in “Demo data, ops docs, and demo checklist”
  - **Estimated tokens:** 1500
