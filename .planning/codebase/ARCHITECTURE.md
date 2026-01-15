# Architecture

**Analysis Date:** 2026-01-15

## Pattern Overview

**Overall:** Layered Monolithic Architecture with RESTful API and SPA Frontend

**Key Characteristics:**
- Modular monolith with clear separation of concerns (handlers → services → repositories)
- Multi-tenant PostgreSQL database (tenant_id on all tables)
- RESTful API with functional HTTP handlers
- React SPA with server state management (React Query)
- Deterministic asset matching with audit trail
- Postgres-only search (indexes and materialized views, no external search engine)

## Layers

**Presentation Layer (HTTP Handlers):**
- Purpose: HTTP request/response handling, authentication/authorization boundaries
- Contains: Route handlers in `api/internal/handlers/*.go`
- Depends on: Repository layer, auth middleware, services
- Used by: HTTP server (`api/internal/server/server.go`)

**Business Logic Layer:**
- Purpose: Complex business logic separate from handlers
- Contains: Ingestion logic (`api/internal/ingest/*.go`), query framework (`api/internal/services/query/`), intel sync (`api/internal/intel/*.go`)
- Depends on: Repository layer for data access
- Used by: HTTP handlers

**Data Access Layer (Repository Pattern):**
- Purpose: Encapsulate data access logic
- Contains: Repository implementations in `api/internal/repository/*.go`
- Depends on: Database connection (sqlx)
- Used by: Handlers, services, business logic

**Domain/Types Layer:**
- Purpose: Domain models and type definitions
- Contains: Ingestion types (`api/internal/ingest/types.go`), query types (`api/internal/services/query/types.go`), repository models
- Depends on: None (pure data structures)
- Used by: Handlers, services, repositories

**Cross-Cutting Concerns:**
- Purpose: Aspects affecting multiple layers
- Contains: Auth/authorization (`api/internal/auth/`), middleware (`api/internal/middleware/`), configuration (`api/internal/config/config.go`), database (`api/internal/database/database.go`)
- Used by: All layers

**Frontend Component Layer:**
- Purpose: Reusable UI components
- Contains: React components in `ui/src/components/*.tsx`
- Depends on: API client, custom hooks, types
- Used by: Page components

**Frontend Page Layer:**
- Purpose: Route-level components
- Contains: Page components in `ui/src/pages/*.tsx`
- Depends on: Components, API client, hooks
- Used by: Router in `ui/src/App.tsx`

**Frontend State Management:**
- Purpose: Server and UI state management
- Contains: Custom hooks (`ui/src/hooks/*.ts`), contexts (`ui/src/contexts/*.tsx`)
- Depends on: React Query, React Context API
- Used by: Components and pages

## Data Flow

**HTTP Request (Backend):**

1. HTTP request received
2. Server router matches route (`api/internal/server/server.go`)
3. Middleware chain executes (logging, request ID)
4. Auth middleware validates JWT/API key, attaches UserContext (`api/internal/handlers/middleware.go`)
5. Handler function invoked (`api/internal/handlers/*.go`)
6. Business logic executes (if needed) via services
7. Repository performs database operations
8. Response formatted and returned

**Ingestion Pipeline Flow:**

1. POST `/api/v1/ingest/vm/findings` received
2. API key validated (or bypassed in demo mode)
3. Payload validated
4. Identifiers normalized (hostname, IP, external IDs) - `api/internal/ingest/normalize.go`
5. Asset matched deterministically - `api/internal/ingest/matching.go`
6. Asset upserted (create or update) - `api/internal/ingest/asset.go`
7. Finding definition upserted - `api/internal/ingest/definition.go`
8. Finding instance upserted - `api/internal/ingest/finding.go`
9. Software normalized to CPE and upserted - `api/internal/ingest/software.go`
10. Stale software deleted
11. Response returned with summary

**Query Execution Flow:**

1. POST `/api/v1/query/{entity_type}` received
2. JSON query validated
3. Query validator checks syntax - `api/internal/services/query/validator.go`
4. Query translated to SQL - `api/internal/services/query/translator.go`
5. Tenant ID filter injected (for security) - `api/internal/services/query/executor.go`
6. SQL executed with timeout (5s max)
7. Results returned with metadata

**Frontend Request Flow:**

1. User interaction in component
2. Component calls API client method
3. API client adds auth headers (JWT or demo mode)
4. HTTP request to backend API
5. Backend processes (see HTTP request flow)
6. Response received
7. React Query updates cache
8. Component re-renders with new data

**State Management:**
- Backend: Stateless (no persistent in-memory state)
- Database: PostgreSQL with connection pooling
- Frontend: React Query for server state, React Context for global UI state (toasts, auth)

## Key Abstractions

**Repository:**
- Purpose: Encapsulate data access logic
- Examples: `AssetRepository`, `FindingRepository`, `SoftwareRepository` in `api/internal/repository/*.go`
- Pattern: Constructor `New{Entity}Repository(db *sqlx.DB)`, methods like `GetByID`, `List`, `Create`, `Update`, `Upsert`

**Functional Handler:**
- Purpose: Dependency injection via closure parameters
- Examples: All handlers in `api/internal/handlers/*.go` (e.g., `ListAssets(db *sqlx.DB)`)
- Pattern: `func HandlerName(dependencies) http.HandlerFunc { return func(w, r) { ... } }`

**Middleware:**
- Purpose: Composable request processing
- Examples: `RequireAuth`, `RequireRole` in `api/internal/handlers/middleware.go`
- Pattern: Higher-order function wrapping next handler

**Service:**
- Purpose: Complex business logic separate from handlers
- Examples: `QueryExecutor`, `QueryTranslator`, `QueryValidator` in `api/internal/services/query/`
- Pattern: Struct-based services with methods for operations

**Deterministic Asset Matching:**
- Purpose: Explainable asset deduplication with audit trail
- Implementation: `api/internal/ingest/matching.go`
- Algorithm: External IDs (highest) → Hostname (canonical) → Shortname → IP + Hostname → No match (create new)

**Context-based Auth:**
- Purpose: User context attached to request context
- Implementation: `api/internal/auth/jwt.go`, `api/internal/auth/apikey.go`
- Pattern: Middleware parses JWT/API key, attaches `UserContext` to `context.Context`

**Custom Hook:**
- Purpose: Reusable stateful logic
- Examples: `useQuery`, `useDashboardQueries` in `ui/src/hooks/*.tsx`
- Pattern: `function use{Feature}(params) { return ... }`

## Entry Points

**Backend API Server:**
- Location: `api/cmd/server/main.go`
- Triggers: HTTP requests to configured port (default 8080)
- Responsibilities: Load config, initialize database, create server, register routes, graceful shutdown

**Backend OQL Test CLI:**
- Location: `api/cmd/test-oql/main.go`
- Triggers: Command-line execution
- Responsibilities: Test OQL queries against database

**Frontend Application:**
- Location: `ui/src/main.tsx`
- Triggers: Browser loads React application
- Responsibilities: Mount app, set up routing, provide contexts

**Frontend Root Component:**
- Location: `ui/src/App.tsx`
- Triggers: Application initialization
- Responsibilities: Define routes, layout structure

**Database Seed Script:**
- Location: `scripts/seed-data.go`
- Triggers: Manual execution
- Responsibilities: Seed database with sample data

## Error Handling

**Strategy:** Throw errors, catch at handler level, log and return standardized error responses

**Patterns:**
- Go: Log errors with zerolog, return HTTP error responses via helpers in `api/internal/api/errors.go`
- TypeScript: Log errors in API client, throw exceptions for component error boundaries
- Demo mode: Many auth errors bypassed in demo mode (security concern)

**Cross-Cutting Concerns:**

**Logging:**
- Backend: zerolog structured logging (`api/internal/database/database.go` initialization)
- Frontend: Console logging in API client (`ui/src/api/client.ts`), toast notifications for user errors

**Validation:**
- Backend: JSON payload validation in handlers, query validation in `api/internal/services/query/validator.go`
- Frontend: TypeScript types, React Query error handling

**Authentication:**
- Backend: JWT validation (`api/internal/auth/jwt.go`), API key validation (`api/internal/auth/apikey.go`), RBAC middleware (`api/internal/auth/rbac.go`)
- Frontend: OIDC PKCE flow via `oidc-client-ts`, demo mode bypass
- Demo mode: Bypasses all auth (security concern - not production ready)

---

*Architecture analysis: 2026-01-15*
*Update when major patterns change*
