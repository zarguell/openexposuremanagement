# Codebase Structure

**Analysis Date:** 2026-01-15

## Directory Layout

```
openexposuremanagement/
├── api/                          # Go API backend service
│   ├── cmd/                       # Entry points
│   │   ├── server/                # Main API server
│   │   └── test-oql/              # OQL query testing CLI
│   ├── internal/                  # Private application code
│   │   ├── api/                   # API errors and types
│   │   ├── auth/                  # Authentication & authorization
│   │   ├── config/                # Configuration management
│   │   ├── database/              # Database connection & utilities
│   │   ├── handlers/              # HTTP request handlers
│   │   ├── ingest/                # Ingestion business logic
│   │   ├── intel/                 # Threat intelligence sync
│   │   ├── middleware/            # HTTP middleware
│   │   ├── repository/            # Data access layer
│   │   ├── server/                # HTTP server setup & routing
│   │   ├── services/              # Business logic services
│   │   │   └── query/             # Query framework
│   │   │       └── oql/           # OQL parser/translator
│   │   └── software/              # Software CPE normalization
│   ├── migrations/                # Database migrations (empty)
│   ├── docs/                      # Generated API docs (Swagger)
│   ├── go.mod                     # Go module definition
│   ├── go.sum                     # Go dependencies
│   └── Dockerfile                 # Container image for API
│
├── ui/                            # React SPA frontend
│   ├── src/                       # Source code
│   │   ├── api/                   # API client
│   │   ├── auth/                  # Authentication utilities
│   │   ├── components/            # Reusable React components
│   │   ├── contexts/              # React contexts
│   │   ├── hooks/                 # Custom React hooks
│   │   ├── pages/                 # Page components
│   │   ├── test/                  # Test utilities
│   │   ├── types/                 # TypeScript types
│   │   ├── utils/                 # Utility functions
│   │   ├── App.tsx                # Root component
│   │   └── main.tsx               # Entry point
│   ├── dist/                      # Build output
│   ├── index.html                 # HTML template
│   ├── package.json               # NPM dependencies
│   ├── tsconfig.json              # TypeScript config
│   ├── vite.config.ts             # Vite build config
│   └── Dockerfile                 # Container image for UI
│
├── db/                            # Database schemas (root-level)
│   └── migrations/                # SQL migration files
│
├── docs/                          # Project documentation
│   ├── .memory.md                 # Documentation memory
│   ├── architecture.md            # MVP spec
│   ├── operations.md              # Operational guide
│   ├── tasks.md                   # Implementation roadmap
│   ├── oql.md                     # Query language spec
│   └── plans/                     # Implementation plans
│
├── sample-data/                   # Sample payloads
├── scripts/                       # Utility scripts
├── .devcontainer/                 # Dev container config
├── .claude/                       # AI agent settings
├── AGENTS.md                      # Agent workflow (TDD)
├── TESTING.md                     # Testing guide
├── DEPLOYMENT.md                  # Deployment instructions
├── README.md                      # Project overview
├── Makefile                       # Development commands
├── setup.sh                       # Automated setup
├── demo.sh                        # Quick demo launcher
└── docker-compose.yml             # Service orchestration
```

## Directory Purposes

**api/**
- Purpose: Go backend API service
- Contains: Source code, build artifacts, Dockerfile
- Key files: `api/cmd/server/main.go`, `api/go.mod`
- Subdirectories: `cmd/` (entry points), `internal/` (private code), `migrations/` (empty), `docs/` (Swagger)

**api/internal/**
- Purpose: Private application code (Go convention)
- Contains: All business logic, handlers, repositories
- Key files: All Go source files
- Subdirectories: `api/` (errors), `auth/` (auth/jwt/rbac), `config/`, `database/`, `handlers/`, `ingest/`, `intel/`, `middleware/`, `repository/`, `server/`, `services/`, `software/`

**api/internal/handlers/**
- Purpose: HTTP request handlers
- Contains: Route handlers for all endpoints
- Key files: `assets.go`, `findings.go`, `ingest.go`, `dashboard.go`, `query.go`, `software.go`, `middleware.go`

**api/internal/repository/**
- Purpose: Data access layer
- Contains: Repository implementations for all entities
- Key files: `asset.go`, `finding.go`, `definition.go`, `software.go`, `user.go`, `apikey.go`, `tenant.go`, `policy.go`, `intel.go`, `dashboard.go`

**api/internal/ingest/**
- Purpose: Ingestion business logic
- Contains: Asset matching, normalization, upsert logic
- Key files: `asset.go`, `finding.go`, `definition.go`, `software.go`, `matching.go`, `normalize.go`

**api/internal/services/query/**
- Purpose: Query framework service
- Contains: Query validation, translation, execution, OQL parser
- Key files: `executor.go`, `translator.go`, `validator.go`, `oql/` (parser, tokenizer, translator, AST)

**api/internal/auth/**
- Purpose: Authentication and authorization
- Contains: JWT validation, API key validation, RBAC
- Key files: `jwt.go`, `apikey.go`, `rbac.go`

**api/internal/intel/**
- Purpose: Threat intelligence sync
- Contains: NVD, EPSS, KEV sync logic
- Key files: `nvd.go`, `epss.go`, `kev.go`, `sync.go`

**ui/**
- Purpose: React SPA frontend
- Contains: Source code, build output, dependencies
- Key files: `ui/package.json`, `ui/vite.config.ts`, `ui/Dockerfile`

**ui/src/**
- Purpose: Frontend source code
- Contains: Components, pages, API client, hooks
- Key files: `ui/src/App.tsx`, `ui/src/main.tsx`
- Subdirectories: `api/`, `auth/`, `components/`, `contexts/`, `hooks/`, `pages/`, `test/`, `types/`, `utils/`

**ui/src/components/**
- Purpose: Reusable React components
- Contains: UI components for tables, query builders, drawers
- Key files: `QueryBuilder.tsx`, `QueryResultsTable.tsx`, `DataTable.tsx`, `AssetDrawer.tsx`, `FindingDrawer.tsx`, `SoftwareDrawer.tsx`

**ui/src/pages/**
- Purpose: Page-level components
- Contains: Route components for each page
- Key files: `Dashboard.tsx`, `Assets.tsx`, `Findings.tsx`, `Software.tsx`, `UnifiedQueries.tsx`, `AssetsQuery.tsx`, `FindingsQuery.tsx`

**ui/src/api/**
- Purpose: API client
- Contains: HTTP client with auth injection
- Key files: `client.ts`

**ui/src/hooks/**
- Purpose: Custom React hooks
- Contains: Server state management via React Query
- Key files: `useQuery.ts`, `useDashboardQueries.ts`

**ui/src/contexts/**
- Purpose: React contexts
- Contains: Global state providers
- Key files: `ToastContext.tsx`

**ui/src/auth/**
- Purpose: Authentication utilities
- Contains: OIDC configuration, auth context
- Key files: `authConfig.ts`, `AuthContext.tsx`, `Login.tsx`, `AuthCallback.tsx`, `ProtectedRoute.tsx`

**db/migrations/**
- Purpose: Database schema migrations
- Contains: SQL migration files
- Key files: 17 migration files from `000001_create_tenants_table.up.sql` to `000017_add_unified_query_indexes.up.sql`

**docs/**
- Purpose: Project documentation
- Contains: Architecture, operations, tasks, OQL spec
- Key files: `architecture.md`, `operations.md`, `tasks.md`, `oql.md`

**sample-data/**
- Purpose: Sample ingestion payloads
- Contains: JSON payloads from Tenable, Qualys
- Key files: `tenable-scan-1.json`, `qualys-scan-1.json`, `tenable-scan-software.json`

**scripts/**
- Purpose: Utility scripts
- Contains: Database seeding, backups, smoke tests
- Key files: `seed-data.go`, `backup.sh`, `smoke-test.sh`

## Key File Locations

**Entry Points:**
- `api/cmd/server/main.go` - Backend API server
- `api/cmd/test-oql/main.go` - OQL query testing CLI
- `ui/src/main.tsx` - Frontend entry point
- `ui/src/App.tsx` - Frontend root component

**Configuration:**
- `api/internal/config/config.go` - Backend config loader
- `api/go.mod` - Go dependencies
- `ui/package.json` - NPM dependencies
- `ui/tsconfig.json` - TypeScript config
- `ui/vite.config.ts` - Vite build config
- `.env`, `ui/.env` - Environment variables
- `.env.example`, `ui/.env.example` - Environment templates

**Core Logic:**
- `api/internal/handlers/*.go` - HTTP handlers
- `api/internal/repository/*.go` - Data access
- `api/internal/ingest/*.go` - Ingestion logic
- `api/internal/services/query/*.go` - Query framework
- `api/internal/intel/*.go` - Threat intel sync
- `ui/src/api/client.ts` - API client
- `ui/src/hooks/*.ts` - Custom hooks

**Testing:**
- `api/**/*_test.go` - Go tests (co-located)
- `ui/src/**/*.test.{ts,tsx}` - React tests (co-located)
- `ui/src/test/setup.ts` - Test setup

**Documentation:**
- `docs/architecture.md` - MVP spec
- `docs/tasks.md` - Implementation roadmap
- `docs/oql.md` - Query language spec
- `AGENTS.md` - Agent workflow
- `TESTING.md` - Testing guide
- `DEPLOYMENT.md` - Deployment instructions

## Naming Conventions

**Files:**
- Go: `snake_case.go` (e.g., `asset.go`, `matching.go`)
- Go tests: `{source}_test.go` (e.g., `asset_test.go`)
- TypeScript: `PascalCase.ts` or `.tsx` (e.g., `QueryBuilder.tsx`)
- TS tests: `{source}.test.{ts,tsx}`
- Config: kebab-case (e.g., `vite.config.ts`)

**Directories:**
- All directories: `kebab-case` (e.g., `api/internal/services/query/`)
- Plural names for collections (e.g., `handlers/`, `components/`, `pages/`)

**Special Patterns:**
- Go tests: Co-located with source (`{file}_test.go`)
- TS tests: Co-located with source (`{file}.test.tsx`)
- Entry points: `cmd/{name}/main.go`
- Migrations: Numbered with `.up.sql` suffix

## Where to Add New Code

**New Feature (Backend):**
- Primary code: `api/internal/handlers/{feature}.go`
- Business logic: `api/internal/services/{feature}/` or `api/internal/{feature}/`
- Data access: `api/internal/repository/{entity}.go`
- Tests: Co-located `{file}_test.go`

**New Feature (Frontend):**
- Primary code: `ui/src/pages/{FeaturePage}.tsx`
- Components: `ui/src/components/{ComponentName}.tsx`
- API client: `ui/src/api/client.ts` (add methods)
- Hooks: `ui/src/hooks/use{Feature}.ts`
- Tests: Co-located `{file}.test.{ts,tsx}`

**New Repository (Backend):**
- Implementation: `api/internal/repository/{entity}.go`
- Tests: `api/internal/repository/{entity}_test.go`
- Migration: `db/migrations/` (if schema change)

**New Component (Frontend):**
- Implementation: `ui/src/components/{ComponentName}.tsx`
- Types: `ui/src/types/` (if needed)
- Tests: `ui/src/components/{ComponentName}.test.tsx`

**New API Endpoint:**
- Handler: `api/internal/handlers/{resource}.go`
- Route registration: `api/internal/server/server.go`
- Repository: `api/internal/repository/{resource}.go` (if needed)
- Tests: `api/internal/handlers/{resource}_test.go`

**New Query Type:**
- OQL support: `api/internal/services/query/oql/` (parser, translator, AST)
- Handler: `api/internal/handlers/query.go`
- Tests: `api/internal/services/query/*_test.go`

## Special Directories

**api/internal/**
- Purpose: Private application code (not importable by external packages)
- Source: Go convention (`internal/` directory is non-importable)
- Committed: Yes

**api/docs/**
- Purpose: Generated API documentation (Swagger)
- Source: Auto-generated from Go comments using swag
- Committed: Yes (checked into repo)

**ui/dist/**
- Purpose: Build output (Vite production build)
- Source: Auto-generated by `npm run build`
- Committed: No (in `.gitignore`)

**ui/node_modules/**
- Purpose: NPM dependencies
- Source: Auto-generated by `npm install`
- Committed: No (in `.gitignore`)

**db/migrations/**
- Purpose: Database schema migrations
- Source: Handwritten SQL files
- Committed: Yes

**.devcontainer/**
- Purpose: VS Code dev container configuration
- Source: VS Code dev containers
- Committed: Yes

---

*Structure analysis: 2026-01-15*
*Update when directory structure changes*
