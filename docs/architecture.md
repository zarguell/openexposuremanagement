# Open Exposure Management (OEM) — MVP Spec (single-machine demo)

## Goal (MVP "aha moment")
Demonstrate a working, self-hosted platform that can ingest infrastructure vulnerability findings, unify assets enough to browse/search, and enrich findings with NVD metadata + EPSS + CISA KEV (with "last intel updated" visible).

## Non-goals (explicitly out of scope for MVP)
- OpenSearch / external search cluster (Postgres-only).
- Pull-mode connectors/workers (push ingestion only).
- SEO / SSR / React Server Components.
- Full historical reporting (current-state only; keep schema "stretch-ready").
- Ticketing / Jira/ServiceNow, scheduling, alerting, AI agents.
- **Suppressions workflow** (proposal/approval/revoke) - moved to post-MVP.
- Cross-scanner suppression (too complex for MVP baseline).

---

## Architecture (single-machine)
### Components
- **Go API**: serves REST API, authN/authZ checks, ingestion, queries, TI sync jobs.
- **PostgreSQL**: system of record + query engine for MVP search and aggregations.
- **React SPA** (Vite): dashboard + asset/finding views.
- **Optional (demo/dev only)**: docker-compose to run API + DB + UI on one host.

### Why no OpenSearch
For MVP, use Postgres indexes/materialized views for “dashboard-style” analytics, refreshed every few minutes; Postgres supports `REFRESH MATERIALIZED VIEW CONCURRENTLY` to avoid blocking reads if a unique index exists. [web:55][web:59]

---

## Authentication & authorization (MVP)
### SPA login
- OIDC Authorization Code flow + **PKCE** (public client; no client secret). [web:16][web:61]
- SPA calls API with bearer access token in `Authorization: Bearer ...`.

### Token storage (Option A)
- Store access token **in memory** (not localStorage) to reduce exposure in case of XSS; accept that hard refresh triggers re-auth redirect, typically without re-entering credentials if the IdP session is still valid. [web:32][web:16]

### Roles (simple)
- `admin`: manage tenants, users/roles, configure asset matching, trigger intel refresh.
- `analyst`: view data, search findings.
- `viewer`: read-only.

### API keys (service-to-service)
- Support scoped API keys for ingestion (e.g., `ingest:vm`) and optionally bind them to a fixed `source`.
- **Reject** ingestion payloads whose `source` mismatches the API key’s bound source.
- Global “demo key” allowed only in dev mode (explicit config flag).

---

## Data model (MVP logical)
### Tenancy & RBAC
- `tenants(id, name, created_at)`
- `users(id, tenant_id, email, display_name, status)`
- `roles(id, tenant_id, name)` — MVP fixed roles: admin/analyst/viewer
- `user_roles(user_id, role_id)`
- `api_keys(id, tenant_id, name, key_hash, scopes_json, bound_source nullable, created_at, revoked_at)`

### Assets (CMDB-lite, DHCP-aware)
Core tables:
- `assets(id, tenant_id, canonical_name, first_seen_at, last_seen_at, owner_team_id nullable, is_active bool)`
- `asset_identifiers(id, tenant_id, asset_id, id_type, id_value, first_seen_at, last_seen_at, source)`
  - `id_type` MVP: `external_id:*` (namespaced), `hostname_norm`, `shortname_norm`, `ipv4`

Asset locator guidance:
- Implement “primary locators” with a priority order; Cisco VM guidance emphasizes putting more static locators above IP when IPs are dynamic. [web:25]
- Because DHCP is common and MAC/serial are hard, use hostname-first + conditional IP matching (below). [web:25][web:41]

### Findings (VM only)
- `finding_definitions(definition_uid pk, source, source_definition_id, title, severity_default, references_json, created_at, updated_at)`
- `finding_definition_aliases(id, definition_uid fk, alias_type, alias_value)`
  - MVP: `alias_type='CVE'`
- `finding_instances(id pk, tenant_id, asset_id fk, definition_uid fk, scanner_status, first_observed_at, last_observed_at, evidence_json, effective_status, effective_reason, effective_revision bigint)`

### Threat intel cache (global, shared)
- `intel_cve(cve pk, description text, cvss_score numeric, cvss_vector varchar, epss_score numeric, epss_percentile numeric, is_kev bool, kev_date_added date, kev_due_date date, updated_at timestamptz)`
  - **NVD data**: `description`, `cvss_score`, `cvss_vector` for CVE details
  - **EPSS data**: `epss_score`, `epss_percentile` for exploitation prediction
  - **CISA KEV data**: `is_kev`, `kev_date_added`, `kev_due_date` for known exploited vulnerabilities
- `intel_sync_runs(id pk, started_at, finished_at, status, error_text nullable, source)`
  - Tracks sync runs for NVD, EPSS, and KEV data sources
  - UI reads latest sync run time and displays "Intel last updated at".

---

## Asset matching (MVP rules)
### Normalization (MVP yes)
- `hostname_norm`: lowercase, trim, remove trailing dot.
- `shortname_norm`: take substring before first `.` from `hostname_norm`.
- Shortname matching is **disabled by default**; admin can enable per tenant.

### Matching algorithm (deterministic & explainable)
Use a rule engine with “why matched” output stored in audit.
Recommended order for on-prem DHCP-heavy preset:
1. External IDs (if present) are strongest (namespaced).
2. `hostname_norm` exact match.
3. `shortname_norm` exact match **only if enabled**.
4. IP address match is **not sufficient alone**; only match on:
   - (`ipv4` AND `hostname_norm` matches), OR
   - asset explicitly marked `static_ip=true` (optional flag) and IP is within a configurable “inactivity window”.
Qualys notes that DHCP IP dedupe should use timestamp proximity (“IP inactivity time”) to reduce bad merges. [web:41]

### Admin merge/undo
- Maintain audit log for merges and allow admin undo.
- Undo reconstitutes original asset rows and reattaches findings (implementation detail; must be deterministic).

---

## Effective status computation (on write)
### Why
Reads should be fast, so we compute effective status on write.

### Mechanism
- On ingestion (or finding upsert), compute and write:
  - `effective_status`: maps scanner status to effective status
    - `open` → `open`
    - `fixed`, `fixed_by_verification` → `fixed`
  - `effective_reason`: always `"scanner"` for MVP (no suppressions)
  - `effective_revision`: always `0` for MVP (no policy state needed)
- **Note**: Schema remains stretch-ready for future suppressions feature (post-MVP).

---

## Threat intel (TI) module (inside Go API)
### Sync behavior
- Daily scheduled job:
  - Pull **NVD CVE data** (description, CVSS score, CVSS vector) and upsert `intel_cve`.
  - Pull EPSS (point-in-time snapshot) and upsert `intel_cve` EPSS fields.
  - Pull CISA KEV catalog and upsert `intel_cve` KEV fields.
- Manual admin endpoint triggers "refresh now".
- UI shows "Intel last updated at" from `intel_sync_runs`.
- **Data sources**:
  - NVD API v2.0: https://nvd.nist.gov/developers/vulnerabilities
  - EPSS data via EPSS tracker: https://www.first.org/epss
  - CISA KEV catalog: https://www.cisa.gov/known-exploited-vulnerabilities-catalog

---

## Postgres-only search & dashboards (MVP)
### Query patterns
- Findings list: filter by `source`, severity, `effective_status`, `cve`, asset name.
- Asset list: search by canonical name / hostname.

### Indexing (examples)
- B-tree indexes: `(tenant_id, last_seen_at)`, `(tenant_id, effective_status)`, `(tenant_id, definition_uid)`.
- GIN index for JSON fields if needed (evidence/ref filters).
- Optional full-text / trigram index for hostname search.

### Dashboard aggregations
Use materialized views refreshed every few minutes; if using `CONCURRENTLY`, ensure a unique index on the materialized view. [web:55][web:59]

---

## API (MVP)
Base: `/api/v1`

### Auth
- `GET /me` — user identity + roles + tenant.

### Ingestion (push)
- `POST /ingest/vm/findings`
  - Auth: API key with `ingest:vm` OR user token with `admin/analyst`.
  - Validates payload + enforces `source` match with API key (reject on mismatch).
  - Performs:
    1. Normalize asset identifiers
    2. Match or create asset
    3. Upsert definition
    4. Upsert finding instance window (first/last observed)
    5. Attach CVE aliases
    6. Compute `effective_status` on write

### Assets
- `GET /assets?query=...`
- `GET /assets/{id}`

### Findings
- `GET /findings?query=...&include_suppressed=false`
  - Returns `scanner_status`, `effective_status`, `effective_reason`, CVE aliases, NVD metadata (description, CVSS), EPSS/KEV fields (if present).

### Threat intel (admin)
- `POST /intel/refresh` — triggers sync job
- `GET /intel/status` — last run, counts, errors

---

## Frontend (React SPA, MVP)
### Pages
- Login (OIDC PKCE).
- Dashboard:
  - total assets, open findings counts by severity
  - "Intel last updated at"
- Asset Inventory:
  - table + asset details drawer
- Findings List:
  - table with filters (asset, cve, source, severity, effective_status)
  - show NVD description, CVSS score, EPSS score + KEV flags/due date if present

### State
- React Query for server-state caching/pagination.
- Zustand for UI state.

---

## Operations (MVP)
- docker-compose: `api`, `postgres`, `ui` (nginx static hosting).
- Migrations: `golang-migrate` for Postgres.
- Backup/restore: backup docker volumes (documented as MVP baseline).

---

## MVP acceptance checklist (demo-ready)
- Ingest sample Tenable/Qualys-like payloads via API key.
- Assets appear deduped by hostname (DHCP-safe).
- Findings list searchable and filterable in UI.
- **NVD enrichment visible** (CVE descriptions, CVSS scores/vectors).
- EPSS/KEV enrichment visible, plus "last updated" timestamp.
- Dashboard shows aggregate counts and intel sync status.
