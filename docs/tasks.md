## Milestone: Repo & dev baseline
One-time setup to make the single-machine demo reproducible with a tight feedback loop; reset context after this milestone.

- [x] Task: Create repo skeleton & conventions
  - **Description:** Initialize folder structure (`api/`, `ui/`, `db/`, `docs/`) and establish naming, linting, and commit conventions.
  - **Acceptance criteria:** Repo has consistent layout; `make help` (or equivalent) lists common commands; basic README exists.
  - **Validation command:** `make help`
  - **Dependencies:** None
  - **Estimated tokens:** 2200

- [x] Task: Add docker-compose baseline (api, postgres, ui)
  - **Description:** Create `docker-compose.yml` for local single-host demo with network wiring, env vars, and volumes.
  - **Acceptance criteria:** `docker compose up` brings up Postgres, API, and UI containers; UI can reach API via configured base URL.
  - **Validation command:** `docker compose up --build`
  - **Dependencies:** Create repo skeleton & conventions
  - **Estimated tokens:** 2600

- [x] Task: Add Makefile/dev scripts
  - **Description:** Add `make` targets for `dev`, `db`, `migrate`, `test`, `lint`, `seed`.
  - **Acceptance criteria:** Targets run without manual steps; commands are documented in `README.md`.
  - **Validation command:** `make test`
  - **Dependencies:** Add docker-compose baseline (api, postgres, ui)
  - **Estimated tokens:** 1800

- [x] Task: Milestone refactor & duplication pass
  - **Description:** Review scaffolding for duplicated config, hard-coded ports, and inconsistent env naming.
  - **Acceptance criteria:** Minimal duplication; `docker compose up --build` still works; lint passes for any added scripts.
  - **Validation command:** `docker compose up --build`
  - **Dependencies:** All tasks in “Repo & dev baseline”
  - **Estimated tokens:** 1500

***

## Milestone: Database schema & migrations
Define Postgres schema, indexes, and materialized views to support Postgres-only “search & dashboards”; reset context after this milestone.

- [x] Task: Add migrations framework (golang-migrate)
  - **Description:** Wire `golang-migrate` into the repo and compose environment; add migration runner command.
  - **Acceptance criteria:** Can apply/rollback migrations against local Postgres; migration history table present.
  - **Validation command:** `make migrate-up && make migrate-down`
  - **Dependencies:** Repo & dev baseline → Add docker-compose baseline (api, postgres, ui)
  - **Estimated tokens:** 2400

- [x] Task: Create core tenancy/RBAC tables
  - **Description:** Implement tables: `tenants`, `users`, `roles`, `user_roles`, `api_keys` with constraints and indexes.
  - **Acceptance criteria:** Foreign keys enforced; unique constraints on tenant/name where appropriate; `revoked_at` supported for keys.
  - **Validation command:** `make migrate-up && psql "$DATABASE_URL" -c "\dt"`
  - **Dependencies:** Add migrations framework (golang-migrate)
  - **Estimated tokens:** 2600

- [x] Task: Create asset tables & identifier indexes
  - **Description:** Implement `assets`, `asset_identifiers` plus indexes for lookup by `(tenant_id, id_type, id_value)` and recency.
  - **Acceptance criteria:** Identifiers support multiple sources and time windows; indexes exist to support matching algorithm.
  - **Validation command:** `make migrate-up && psql "$DATABASE_URL" -c "\d assets" && psql "$DATABASE_URL" -c "\d asset_identifiers"`
  - **Dependencies:** Create core tenancy/RBAC tables
  - **Estimated tokens:** 2800

- [x] Task: Create findings tables & indexes
  - **Description:** Implement `finding_definitions`, `finding_definition_aliases`, `finding_instances` with indexes for tenant filters (`effective_status`, `definition_uid`, `last_observed_at`).
  - **Acceptance criteria:** Upsert-friendly keys; aliases allow multiple CVEs per definition; instances reference assets and definitions.
  - **Validation command:** `make migrate-up && psql "$DATABASE_URL" -c "\d finding_instances"`
  - **Dependencies:** Create asset tables & identifier indexes
  - **Estimated tokens:** 2900

- ~~Task: Create suppression workflow tables~~ (MOVED TO POST-MVP)
  - **Description:** ~~Implement `suppressions`, `suppression_reviews`, `tenant_policy_state` with constraints for state transitions and revision tracking.~~
  - **Acceptance criteria:** ~~`tenant_policy_state.policy_revision` exists and is incrementable; suppression audit trail stored.~~
  - **Note:** Suppressions workflow is out of scope for MVP. Schema reserved for future implementation.
  - **Validation command:** `make migrate-up && psql "$DATABASE_URL" -c "\d suppressions"`
  - **Dependencies:** Create findings tables & indexes
  - **Estimated tokens:** 2800

- [x] Task: Create threat intel cache tables (with NVD + EPSS + KEV fields)
  - **Description:** Implement `intel_cve` and `intel_sync_runs` tables for EPSS/KEV caching and “last updated” display.
  - **Acceptance criteria:** `intel_cve.cve` is primary key; `intel_sync_runs` records status and errors.
  - **Validation command:** `make migrate-up && psql "$DATABASE_URL" -c "\d intel_cve"`
  - **Dependencies:** Create findings tables & indexes
  - **Estimated tokens:** 2400

- [x] Task: Add dashboard materialized views + concurrent refresh readiness
  - **Description:** Define MVP materialized views (counts by effective status, open findings, assets active) and create required unique indexes to allow `REFRESH MATERIALIZED VIEW CONCURRENTLY` where used.[1]
  - **Acceptance criteria:** Views populate; refresh commands succeed; if `CONCURRENTLY` used, a qualifying unique index exists.[1]
  - **Validation command:** `psql "$DATABASE_URL" -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_dashboard_counts;"`
  - **Dependencies:** Create asset tables & identifier indexes; Create findings tables & indexes
  - **Estimated tokens:** 3000

- [x] Task: Milestone refactor & duplication pass
  - **Description:** Normalize naming, constraints, and index patterns across migrations; remove redundant indexes.
  - **Acceptance criteria:** `make migrate-up` succeeds from empty DB; schema is consistent; no unused/duplicate indexes.
  - **Validation command:** `make migrate-up`
  - **Dependencies:** All tasks in "Database schema & migrations"
  - **Estimated tokens:** 1800

***

## Milestone: Go API foundation (authN/authZ + config)
Implement API service skeleton, OIDC JWT validation, and RBAC gates for endpoints; reset context after this milestone.

- [x] Task: Initialize Go API module & health endpoints
  - **Description:** Create Go module, router, structured logging, config loader, and `/healthz` endpoint.
  - **Acceptance criteria:** API starts locally; `/healthz` returns 200; config via env works.
  - **Validation command:** `go test ./... && curl -sf http://localhost:8080/healthz`
  - **Dependencies:** Repo & dev baseline → Add docker-compose baseline (api, postgres, ui)
  - **Estimated tokens:** 2500

- [x] Task: Implement DB layer (sqlc or minimal repository)
  - **Description:** Add Postgres connection pooling, migrations hook, and typed queries for core entities.
  - **Acceptance criteria:** API connects to DB; query functions covered by unit tests (mock or test DB).
  - **Validation command:** `go test ./...`
  - **Dependencies:** Database schema & migrations → Add migrations framework (golang-migrate)
  - **Estimated tokens:** 3000

- [x] Task: Implement OIDC JWT verification for SPA bearer tokens
  - **Description:** Validate `Authorization: Bearer` tokens using issuer JWKS, map claims to user identity, and enforce tenant context.
  - **Acceptance criteria:** Requests without token are rejected; valid token yields user context; supports SPA Auth Code + PKCE usage pattern.[2]
  - **Validation command:** `go test ./...`
  - **Dependencies:** Initialize Go API module & health endpoints
  - **Estimated tokens:** 3000

- [x] Task: Implement RBAC middleware (admin/analyst/viewer)
  - **Description:** Enforce role checks per endpoint; add `GET /me` returning user, tenant, roles.
  - **Acceptance criteria:** Role-protected endpoints return 403 without required role; `/me` returns expected shape.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement OIDC JWT verification for SPA bearer tokens
  - **Estimated tokens:** 2600

- [x] Task: Implement API key auth for ingestion (scopes + bound source)
  - **Description:** Add API key parsing, hashing/verification, scope check (`ingest:vm`), and optional `bound_source` enforcement.
  - **Acceptance criteria:** Ingestion rejects missing/invalid/revoked keys; rejects payload where `source` mismatches bound source.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement DB layer (sqlc or minimal repository); Create core tenancy/RBAC tables
  - **Estimated tokens:** 3000

- [x] Task: Milestone refactor & duplication pass
  - **Description:** Consolidate auth context plumbing, error formats, and config parsing; reduce boilerplate handlers.
  - **Acceptance criteria:** Lint/test passes; no duplicate auth parsing logic across middlewares.
  - **Validation command:** `go test ./...`
  - **Dependencies:** All tasks in “Go API foundation (authN/authZ + config)”
  - **Estimated tokens:** 1700

***

## Milestone: VM ingestion pipeline (assets + findings)
Push ingestion endpoint that normalizes identifiers, matches/creates assets, upserts findings/aliases, and computes effective status on write; reset context after this milestone.

- [ ] Task: Define ingestion payload schema & validation
  - **Description:** Specify JSON contract for Tenable/Qualys-like findings, including `source`, asset identifiers, definition IDs, status, timestamps, evidence.
  - **Acceptance criteria:** Strict validation with clear errors; rejects unknown/invalid fields; includes unit tests with sample payloads.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Go API foundation → Implement API key auth for ingestion (scopes + bound source)
  - **Estimated tokens:** 2600

- [ ] Task: Implement identifier normalization helpers
  - **Description:** Implement `hostname_norm` and `shortname_norm` rules and ensure consistent use across ingestion and search.
  - **Acceptance criteria:** Normalization is deterministic (lowercase/trim/trailing dot removal); unit tests cover edge cases.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Define ingestion payload schema & validation
  - **Estimated tokens:** 1800

- [ ] Task: Implement deterministic asset matching (rule engine + “why matched”)
  - **Description:** Implement matching order: external IDs → hostname → optional shortname → conditional IP logic; produce match explanation for audit.
  - **Acceptance criteria:** Given test fixtures, matching selects expected asset; shortname matching off by default; IP-only never matches.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement identifier normalization helpers; Database schema & migrations → Create asset tables & identifier indexes
  - **Estimated tokens:** 3000

- [ ] Task: Implement asset upsert (create/update seen times + identifiers)
  - **Description:** When matched, update `last_seen_at`; when new, create asset + identifiers with `first_seen_at/last_seen_at`.
  - **Acceptance criteria:** Idempotent ingest; repeated payload doesn’t create duplicates; identifiers update time windows.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement deterministic asset matching (rule engine + “why matched”)
  - **Estimated tokens:** 2900

- [ ] Task: Implement definition + alias upsert (CVE aliases)
  - **Description:** Upsert `finding_definitions` and attach CVE aliases in `finding_definition_aliases`.
  - **Acceptance criteria:** Same definition from same source updates metadata; CVE aliases are deduped; unit tests verify.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Define ingestion payload schema & validation; Database schema & migrations → Create findings tables & indexes
  - **Estimated tokens:** 2400

- [ ] Task: Implement finding instance upsert (observation window)
  - **Description:** Upsert `finding_instances` per asset+definition with first/last observed tracking and evidence JSON.
  - **Acceptance criteria:** First observed only moves earlier; last observed only moves later; scanner status recorded.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement asset upsert (create/update seen times + identifiers); Implement definition + alias upsert (CVE aliases)
  - **Estimated tokens:** 2800

- [ ] Task: Compute effective status on write (baseline)
  - **Description:** Implement effective status computation using scanner status and current approved suppressions at CVE-level.
  - **Acceptance criteria:** Writes set `effective_status`, `effective_reason`, and `effective_revision` from `tenant_policy_state`.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement finding instance upsert (observation window); Database schema & migrations → Create suppression workflow tables
  - **Estimated tokens:** 3000

- [ ] Task: Implement POST /ingest/vm/findings end-to-end
  - **Description:** Wire the endpoint to run validation → match/upsert asset → upsert definition/aliases → upsert instance → compute effective status.
  - **Acceptance criteria:** End-to-end integration test ingests sample payload; enforces source binding rule on API key.
  - **Validation command:** `go test ./...`
  - **Dependencies:** All tasks in “VM ingestion pipeline (assets + findings)”
  - **Estimated tokens:** 3000

- [ ] Task: Milestone refactor & duplication pass
  - **Description:** Consolidate ingestion transaction boundaries, error handling, and repeated query code; reduce cyclomatic complexity.
  - **Acceptance criteria:** `go test ./...` passes; lint/static checks pass; ingestion logic readable and modular.
  - **Validation command:** `go test ./...`
  - **Dependencies:** All tasks in “VM ingestion pipeline (assets + findings)”
  - **Estimated tokens:** 1700

***

## Milestone: Query APIs (assets/findings/dashboard)
Provide Postgres-backed browsing, filtering, and enrichment joins; reset context after this milestone.

- [ ] Task: Implement GET /assets (search by canonical/hostname)
  - **Description:** Add query parameter handling and DB search (ILIKE/trigram optional), returning paginated results.
  - **Acceptance criteria:** Search returns stable pagination; tenant scoping enforced; tests cover query parsing.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Go API foundation → Implement RBAC middleware (admin/analyst/viewer); Database schema & migrations → Create asset tables & identifier indexes
  - **Estimated tokens:** 2600

- [ ] Task: Implement GET /assets/{id} (details + identifiers + finding counts)
  - **Description:** Return asset record plus identifiers and summary counts of findings by effective status.
  - **Acceptance criteria:** 404 on missing/foreign-tenant asset; response shape matches UI needs; unit tests.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement GET /assets (search by canonical/hostname)
  - **Estimated tokens:** 2500

- [ ] Task: Implement GET /findings with filters
  - **Description:** Support filters: `source`, severity, `effective_status`, `cve`, asset name; include suppression toggles.
  - **Acceptance criteria:** Queries are indexed appropriately; returns scanner + effective status + reason + CVEs.
  - **Validation command:** `go test ./...`
  - **Dependencies:** VM ingestion pipeline → Implement POST /ingest/vm/findings end-to-end; Database schema & migrations → Create findings tables & indexes
  - **Estimated tokens:** 3000

- [ ] Task: Join threat intel fields into findings response
  - **Description:** Enrich findings by joining CVE aliases to `intel_cve` (EPSS score/percentile, KEV flags/dates).
  - **Acceptance criteria:** If intel missing, fields are null; if present, fields populate; tests with seeded intel rows.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement GET /findings with filters; Database schema & migrations → Create threat intel cache tables
  - **Estimated tokens:** 2400

- [ ] Task: Implement dashboard endpoints (counts + intel status)
  - **Description:** Add endpoints that read from materialized views plus `GET /intel/status` for “Intel last updated at”.
  - **Acceptance criteria:** Dashboard endpoint latency is low; intel status returns latest run time and error (if any).
  - **Validation command:** `go test ./...`
  - **Dependencies:** Database schema & migrations → Add dashboard materialized views + concurrent refresh readiness; Threat intel cache tables
  - **Estimated tokens:** 2700

- [ ] Task: Add materialized view refresher job
  - **Description:** Add periodic refresh (every few minutes) using `REFRESH MATERIALIZED VIEW CONCURRENTLY` where configured and valid.[1]
  - **Acceptance criteria:** Refresh runs without blocking reads; failures logged and visible via health/status endpoint.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement dashboard endpoints (counts + intel status)
  - **Estimated tokens:** 2600

- [ ] Task: Milestone refactor & duplication pass
  - **Description:** Deduplicate filter parsing, pagination structs, and SQL fragments; ensure consistent response envelopes.
  - **Acceptance criteria:** `go test ./...` passes; lint passes; no duplicated query-building logic.
  - **Validation command:** `go test ./...`
  - **Dependencies:** All tasks in “Query APIs (assets/findings/dashboard)”
  - **Estimated tokens:** 1700

***

## Milestone: Threat intel sync (NVD + EPSS + KEV)
Daily scheduled TI sync plus manual refresh and status display; reset context after this milestone.

- [ ] Task: Implement NVD fetch + upsert
  - **Description:** Fetch NVD CVE data (description, CVSS score, CVSS vector) via NVD API v2.0 and upsert into `intel_cve` with `updated_at`.
  - **Acceptance criteria:** Handles NVD API rate limits; idempotent upsert; records sync run; resilient to partial failures.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Database schema & migrations → Create threat intel cache tables; Go API foundation → Implement DB layer (sqlc or minimal repository)
  - **Estimated tokens:** 3000

- [ ] Task: Implement EPSS fetch + upsert
  - **Description:** Fetch EPSS snapshot data and upsert EPSS fields into `intel_cve` with `updated_at`.
  - **Acceptance criteria:** Handles paging/large files; idempotent upsert; records sync run.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement NVD fetch + upsert
  - **Estimated tokens:** 2800

- [ ] Task: Implement KEV fetch + upsert
  - **Description:** Fetch CISA KEV catalog and upsert KEV fields into `intel_cve`; record sync run.
  - **Acceptance criteria:** Idempotent; updates `is_kev`, dates; resilient to partial failures.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement EPSS fetch + upsert
  - **Estimated tokens:** 2600

- [ ] Task: Implement scheduled job + admin refresh endpoint
  - **Description:** Add daily scheduler and `POST /intel/refresh` to trigger sync; avoid overlapping runs.
  - **Acceptance criteria:** Manual refresh returns job accepted/completed; scheduler can be disabled in dev/test.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement KEV fetch + upsert
  - **Estimated tokens:** 2600

- [ ] Task: Implement GET /intel/status
  - **Description:** Return latest sync run time/status and counts/errors for UI "Intel last updated at".
  - **Acceptance criteria:** Always returns a coherent status even if no run yet; tested.
  - **Validation command:** `go test ./...`
  - **Dependencies:** Implement scheduled job + admin refresh endpoint
  - **Estimated tokens:** 2000

- [ ] Task: Milestone refactor & duplication pass
  - **Description:** Deduplicate HTTP fetching, parsing, and upsert logic; centralize sync run recording.
  - **Acceptance criteria:** Tests pass; lint passes; sync code shared and readable.
  - **Validation command:** `go test ./...`
  - **Dependencies:** All tasks in "Threat intel sync (NVD + EPSS + KEV)"
  - **Estimated tokens:** 1700

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

- [ ] Task: Milestone refactor & duplication pass
  - **Description:** Consolidate state machine logic and audit writing; reduce duplicated transaction code.
  - **Acceptance criteria:** Tests pass; lint passes; state transitions clearly centralized.
  - **Validation command:** `go test ./...`
  - **Dependencies:** All tasks in "Suppressions (proposal/approval + recompute)"
  - **Estimated tokens:** 1600

***

## Milestone: React SPA (login + core pages)
Deliver the demo UI: OIDC PKCE login, dashboard, assets, findings with NVD enrichment; reset context after this milestone.

- [ ] Task: Bootstrap Vite React app with routing and env config
  - **Description:** Set up Vite React SPA, router, API base URL config, and error boundary.
  - **Acceptance criteria:** App starts; routes render; config is environment-driven.
  - **Validation command:** `cd ui && npm test` (or `npm run build`)
  - **Dependencies:** Repo & dev baseline → Add docker-compose baseline (api, postgres, ui)
  - **Estimated tokens:** 2400

- [ ] Task: Implement OIDC Auth Code + PKCE login flow
  - **Description:** Add OIDC client in SPA using Authorization Code + PKCE for a public client (no client secret) and bearer token usage.[2]
  - **Acceptance criteria:** Login redirects to IdP and back; access token stored in memory; refresh triggers re-auth redirect.
  - **Validation command:** `cd ui && npm run build`
  - **Dependencies:** Bootstrap Vite React app with routing and env config
  - **Estimated tokens:** 3000

- [ ] Task: Add React Query + API client with auth header
  - **Description:** Implement typed API client, React Query providers, and automatic `Authorization: Bearer` injection.
  - **Acceptance criteria:** Queries retry sensibly; 401 triggers re-login; errors are shown.
  - **Validation command:** `cd ui && npm run build`
  - **Dependencies:** Implement OIDC Auth Code + PKCE login flow
  - **Estimated tokens:** 2600

- [ ] Task: Build Dashboard page (counts + intel timestamp)
  - **Description:** Create dashboard widgets for total assets, open findings counts by severity, and "Intel last updated at".
  - **Acceptance criteria:** Data loads from API endpoints; empty state handled; timestamp displayed clearly.
  - **Validation command:** `cd ui && npm run build`
  - **Dependencies:** Query APIs → Implement dashboard endpoints (counts + intel status)
  - **Estimated tokens:** 2600

- [ ] Task: Build Asset Inventory page (table + details drawer)
  - **Description:** Implement searchable table and asset details drawer via `/assets` and `/assets/{id}`.
  - **Acceptance criteria:** Search works; selecting row opens drawer; pagination handled.
  - **Validation command:** `cd ui && npm run build`
  - **Dependencies:** Query APIs → Implement GET /assets (search by canonical/hostname); Implement GET /assets/{id}
  - **Estimated tokens:** 3000

- [ ] Task: Build Findings List page (filters + NVD + intel fields)
  - **Description:** Implement table with filters (asset, cve, source, severity, effective_status) and display NVD description, CVSS score, EPSS/KEV fields.
  - **Acceptance criteria:** Filters map to query params; performance acceptable for demo dataset; NVD data displayed prominently.
  - **Validation command:** `cd ui && npm run build`
  - **Dependencies:** Query APIs → Implement GET /findings with filters; Join threat intel fields into findings response
  - **Estimated tokens:** 3200

- ~~Task: (Stretch) Build Suppressions page~~ (MOVED TO POST-MVP)
  - **Description:** ~~Add UI to propose CVE suppression and admin approve/reject/revoke; show state changes.~~
  - **Note:** Suppressions are out of MVP scope.
  - **Validation command:** `cd ui && npm run build`
  - **Dependencies:** Suppressions milestone → Implement suppression listing endpoints (for UI); Async recompute worker
  - **Estimated tokens:** 3000

- [ ] Task: Milestone refactor & duplication pass
  - **Description:** Deduplicate table/filter components and API hooks; ensure consistent loading/empty/error states.
  - **Acceptance criteria:** `npm run build` passes; lint passes; no repeated query logic.
  - **Validation command:** `cd ui && npm run build`
  - **Dependencies:** All tasks in "React SPA (login + core pages)"
  - **Estimated tokens:** 1700

***

## Milestone: Demo data, ops docs, and demo checklist
Make the demo easy to run and verify end-to-end; reset context after this milestone.

- [ ] Task: Add sample ingestion payloads + seeding script
  - **Description:** Provide sample Tenable/Qualys-like JSON payloads and a script to POST them using an ingestion API key.
  - **Acceptance criteria:** Running the script populates assets and findings; repeat runs are idempotent.
  - **Validation command:** `make seed && make demo-smoke`
  - **Dependencies:** VM ingestion pipeline → Implement POST /ingest/vm/findings end-to-end
  - **Estimated tokens:** 2600

- [ ] Task: Add end-to-end “demo smoke test” script
  - **Description:** Script that brings up stack, ingests data, refreshes views, and hits key endpoints (`/assets`, `/findings`, `/intel/status`).
  - **Acceptance criteria:** Script exits non-zero on failure; produces minimal readable output for demo operator.
  - **Validation command:** `make demo-smoke`
  - **Dependencies:** Add sample ingestion payloads + seeding script; Query APIs milestone
  - **Estimated tokens:** 2400

- [ ] Task: Document backup/restore (volume snapshot) and MVP runbook
  - **Description:** Write docs for starting stack, configuring IdP, creating API keys, backup/restore of docker volumes, and common troubleshooting.
  - **Acceptance criteria:** A new developer can run demo from scratch following docs; includes “demo key dev-only” warning.
  - **Validation command:** `markdownlint docs/*.md` (or `make docs-check`)
  - **Dependencies:** Repo & dev baseline → Add Makefile/dev scripts
  - **Estimated tokens:** 2000

- [ ] Task: Milestone refactor & duplication pass
  - **Description:** Review scripts/docs for duplication and drift; ensure commands match Makefile and compose.
  - **Acceptance criteria:** `make demo-smoke` works; docs commands are consistent; lint passes.
  - **Validation command:** `make demo-smoke`
  - **Dependencies:** All tasks in “Demo data, ops docs, and demo checklist”
  - **Estimated tokens:** 1500
