# External Integrations

**Analysis Date:** 2026-01-15

## APIs & External Services

**NVD API v2.0 (National Vulnerability Database):**
- CVE vulnerability data, CVSS scores, descriptions
- SDK/Client: Custom Go HTTP client (`api/internal/intel/nvd.go`)
- Auth: None (public API)
- Rate limits: 5 second delay between requests (default)
- Base URL: https://services.nvd.nist.gov/rest/json/cves/2.0

**CISA KEV Catalog (Known Exploited Vulnerabilities):**
- Known exploited vulnerabilities
- SDK/Client: Custom Go HTTP client (`api/internal/intel/kev.go`)
- Auth: None (public JSON feed)
- Base URL: https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json
- Timeout: 2 minutes

**EPSS API (Exploit Prediction Scoring System):**
- Exploit prediction scores for CVEs
- SDK/Client: Custom Go HTTP client with CSV parsing (`api/internal/intel/epss.go`)
- Auth: None (public CSV)
- Base URL: https://epss.cyentia.com/epss_scores-current.csv.gz
- Format: CSV (gzipped) streaming
- Timeout: 5 minutes

## Data Storage

**Databases:**
- PostgreSQL 16 - Primary data store
  - Connection: via DATABASE_URL env var
  - Client: github.com/lib/pq v1.10.9 (native PostgreSQL driver)
  - ORM/Helper: github.com/jmoiron/sqlx v1.4.0 (type-safe SQL wrapper)
  - Migrations: db/migrations/*.sql via golang-migrate v4.19.1
  - Connection pooling: 25 max open, 5 idle, 5 min lifetime
  - Multi-tenant: All tables have tenant_id column

**File Storage:**
- None currently (no file uploads in MVP)

**Caching:**
- React Query for frontend server state (in-memory cache)
- No Redis or external cache (planned for future)

## Authentication & Identity

**Auth Provider:**
- OIDC PKCE - SPA authentication with third-party IdP
  - Implementation: oidc-client-ts v3.4.1
  - Frontend: `ui/src/auth/authConfig.ts`, `ui/src/auth/AuthContext.tsx`, `ui/src/auth/Login.tsx`
  - Backend: `api/internal/auth/jwt.go` (JWT validation)
  - Token storage: httpOnly cookies (planned), currently localStorage/sessionStorage
  - Supported providers: Any OIDC-compliant (Okta, Auth0, Keycloak, etc.)

**Demo Mode (Bypass Authentication):**
- Purpose: Development/demo mode without external IdP
  - Implementation: Mock user context in both UI and API
  - Files: `ui/src/auth/AuthContext.tsx`, `api/cmd/server/main.go`, `api/internal/handlers/middleware.go`
  - Environment: DEMO_MODE=true
  - Security: Bypasses all auth checks (NOT PRODUCTION READY)

**API Keys (Service-to-Service):**
- Purpose: Ingestion endpoint authentication for scanners
  - Implementation: `api/internal/auth/apikey.go`
  - Status: NOT FULLY IMPLEMENTED (TODOs present)
  - Scopes: `ingest:vm` for VM findings ingestion
  - Source binding: Optional, restricts API key to specific source

## Monitoring & Observability

**Error Tracking:**
- None (no Sentry or similar service)

**Analytics:**
- None (no product analytics)

**Logs:**
- Backend: zerolog structured JSON logging (`api/internal/database/database.go`)
  - Output: stdout/stderr (captured by Docker)
  - Levels: debug, info, warn, error
- Frontend: console logging (dev only)
  - No centralized log aggregation

**API Documentation:**
- Swagger/OpenAPI - Interactive API documentation
  - Implementation: swaggo/http-swagger v1.3.4, swaggo/swag v1.16.6
  - Access: http://localhost:8080/swagger/
  - Annotations: Go comments in handlers (e.g., `// @Summary Ingest VM findings`)

**Database Admin:**
- pgAdmin 8.2 - Database administration UI
  - Image: dpage/pgadmin4:8.2
  - Access: http://localhost:5050
  - Config: `docker-compose.yml` (lines 82-98)

## CI/CD & Deployment

**Hosting:**
- Vite dev server (development): http://localhost:5173
- Production: Not configured (planned: nginx/Alpine via Docker)
  - UI Dockerfile: Multi-stage build (node:20-alpine → nginx:alpine)
  - API Dockerfile: Go build → alpine base

**CI Pipeline:**
- None configured (no GitHub Actions, GitLab CI, etc.)
  - Manual testing via `make test`
  - Manual linting via `make lint`

## Environment Configuration

**Development:**
- Required env vars (backend): DATABASE_URL, API_PORT, OIDC_ISSUER, OIDC_CLIENT_ID, OIDC_SCOPES
- Optional env vars (backend): DEMO_MODE, DEMO_API_KEY, ENVIRONMENT, LOG_LEVEL
- Required env vars (frontend): API_BASE_URL, VITE_OIDC_ISSUER, VITE_OIDC_CLIENT_ID
- Optional env vars (frontend): VITE_ENV
- Secrets location: `.env` (gitignored), `ui/.env` (gitignored)
- Example configs: `.env.example`, `ui/.env.example`
- Mock/stub services: Stripe test mode (N/A), local PostgreSQL via Docker Compose

**Staging:**
- Not configured (planned)
- Would use separate PostgreSQL database
- Would use staging OIDC provider

**Production:**
- Secrets management: Environment variables (Docker/Kubernetes)
- Database: Production PostgreSQL instance
- Demo mode: Must be disabled (DEMO_MODE=false)
- No failover/redundancy configured

## Webhooks & Callbacks

**Incoming:**
- None currently (planned: scanner webhooks for real-time ingestion)

**Outgoing:**
- None currently (planned: notification webhooks for critical findings)

## Internal APIs

**Backend API:**
- Framework: net/http (Go standard library)
- Base URL: http://localhost:8080/api/v1
- Authentication: Bearer JWT or API key (demo mode bypass)
- Router: http.ServeMux with custom middleware
- Documentation: Swagger/OpenAPI

**Key Endpoints:**
- Health: `/healthz`, `/healthz/live`, `/healthz/ready`
- Auth: `/api/v1/me`
- Ingestion: `/api/v1/ingest/vm/findings`
- Assets: `/api/v1/assets`, `/api/v1/assets/{id}`, `/api/v1/assets/{id}/software`
- Findings: `/api/v1/findings`
- Software: `/api/v1/software`, `/api/v1/software/{id}`
- Dashboard: `/api/v1/dashboard`
- Intel: `/api/v1/intel/status`, `/api/v1/intel/refresh`
- Query: `/api/v1/query/findings`, `/api/v1/query/assets`, `/api/v1/query/software_inventory`, `/api/v1/query/unified`, `/api/v1/query/oql`

**Frontend API Client:**
- Implementation: Custom fetch wrapper with auth injection
- Location: `ui/src/api/client.ts`
- Auth: Automatic JWT injection (or demo mode bypass)
- Error handling: Standardized error responses
- Key methods: `getAssets()`, `getAsset()`, `getFindings()`, `getSoftware()`, `queryExecute()`, `queryUnified()`, `getIntelStatus()`, `refreshIntel()`

**Vite Dev Server Proxy:**
- Purpose: Proxy API requests during development
- Configuration: `ui/vite.config.ts`
- Proxies: `/api/*`, `/swagger/*` → `API_BASE_URL`

---

*Integration audit: 2026-01-15*
*Update when adding/removing external services*
