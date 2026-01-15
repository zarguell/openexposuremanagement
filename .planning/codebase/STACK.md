# Technology Stack

**Analysis Date:** 2026-01-15

## Languages

**Primary:**
- Go 1.24.0 - All backend API code
- TypeScript 5.2.2 - All frontend code
- SQL - Database schema and migrations

**Secondary:**
- JavaScript - Frontend build config, package.json scripts
- Shell - Utility scripts (setup.sh, demo.sh)

## Runtime

**Environment:**
- Go 1.24.0 - Backend API server (`api/go.mod`)
- Node.js 20 - Frontend build and dev server (`ui/Dockerfile`)
- PostgreSQL 16 - Primary database (`docker-compose.yml`)
- Alpine Linux - Container base images

**Package Manager:**
- Go Modules - `api/go.mod`, `api/go.sum`
- npm - `ui/package.json`, `ui/package-lock.json`

## Frameworks

**Core - Backend:**
- net/http - Standard library HTTP server (`api/internal/server/server.go`)
- sqlx - Database helper (`api/go.mod`)
- gorilla/mux - HTTP router (implied, check Go files)

**Core - Frontend:**
- React 18.2.0 - UI framework (`ui/package.json`)
- React Router DOM 6.20.0 - Client-side routing (`ui/package.json`)
- TanStack React Query 5.90.16 - Data fetching and caching (`ui/package.json`)
- Vite 7.3.1 - Build tool and dev server (`ui/vite.config.ts`)
- oidc-client-ts 3.4.1 - OIDC PKCE authentication (`ui/package.json`)

**Testing - Backend:**
- testify v1.11.1 - Testing framework (`api/go.mod`)

**Testing - Frontend:**
- Vitest 4.0.16 - Test runner (`ui/package.json`)
- @testing-library/react 16.3.1 - Component testing (`ui/package.json`)
- @testing-library/jest-dom 6.9.1 - DOM matchers (`ui/package.json`)

**Build/Dev:**
- TypeScript 5.2.2 - Compiler (`ui/tsconfig.json`)
- golang-migrate v4.19.1 - Database migrations (`api/go.mod`)
- swagger/http-swagger - API documentation (`api/go.mod`)

## Key Dependencies

**Critical - Backend:**
- github.com/jmoiron/sqlx v1.4.0 - Database operations (`api/go.mod`)
- github.com/lib/pq v1.10.9 - PostgreSQL driver (`api/go.mod`)
- github.com/golang-jwt/jwt/v5 v5.3.0 - JWT validation (`api/go.mod`)
- github.com/rs/zerolog v1.34.0 - Structured logging (`api/go.mod`)
- github.com/golang-migrate/migrate/v4 v4.19.1 - Database migrations (`api/go.mod`)
- github.com/stretchr/testify v1.11.1 - Test assertions (`api/go.mod`)

**Critical - Frontend:**
- @tanstack/react-query 5.90.16 - Server state management (`ui/package.json`)
- oidc-client-ts 3.4.1 - OIDC authentication (`ui/package.json`)

**Infrastructure:**
- docker/compose - Service orchestration (`docker-compose.yml`)
- PostgreSQL 16 - Database (`docker-compose.yml`)

## Configuration

**Environment:**
- Backend: `.env` file with DATABASE_URL, API_PORT, OIDC_ISSUER, OIDC_CLIENT_ID, OIDC_SCOPES, DEMO_MODE, DEMO_API_KEY, ENVIRONMENT, LOG_LEVEL
- Frontend: `ui/.env` with API_BASE_URL, VITE_OIDC_ISSUER, VITE_OIDC_CLIENT_ID, VITE_ENV
- Example configs: `.env.example`, `ui/.env.example`

**Build:**
- Go: `api/go.mod`, `api/go.sum` (Go modules)
- TypeScript: `ui/tsconfig.json` (compiler options)
- Vite: `ui/vite.config.ts` (build config, API proxy)
- Vitest: `ui/vitest.config.ts` (test config)
- Docker: `api/Dockerfile`, `ui/Dockerfile` (container images)
- Docker Compose: `docker-compose.yml` (multi-service setup)
- Makefile: `Makefile` (build, test, run commands)

## Platform Requirements

**Development:**
- Go 1.24.0 or later
- Node.js 20 or later
- Docker and Docker Compose (for local database)
- PostgreSQL 16 (via Docker or local)

**Production:**
- Linux (Alpine via Docker)
- PostgreSQL 16 database
- Environment variables for configuration
- Containerized deployment (Docker images provided)

---

*Stack analysis: 2026-01-15*
*Update after major dependency changes*
