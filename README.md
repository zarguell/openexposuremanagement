# Open Exposure Management (OEM)

A self-hosted vulnerability exposure management platform that ingests infrastructure findings, unifies assets, and enriches findings with EPSS and CISA KEV threat intelligence.

## MVP Goal

Demonstrate a working platform that can:
- Ingest vulnerability findings from VM scanners (Tenable/Qualys-like)
- Unify assets across scans using deterministic matching
- Browse and search assets and findings
- Enrich findings with EPSS scores and CISA KEV data
- Support CVE-level suppression proposal and approval

## Quick Start (Demo)

### 🚀 Fast Demo Mode (No Authentication)

For quick demonstrations, you can run the application without authentication:

```bash
# Start both frontend and backend in demo mode
./demo.sh

# Or manually:
# Terminal 1 - Backend (with demo mode)
cd api && DEMO_MODE=true go run ./cmd/server

# Terminal 2 - Frontend
cd ui && npm run dev

# Or with Docker Compose:
docker compose up -d
# Frontend: http://localhost:80
# Backend API: http://localhost:8080
# PgAdmin: http://localhost:5050 (admin@example.com / admin)
```

**⚠️ Security Warning**: Demo mode disables authentication entirely. This is NOT secure for production use!

### 🔐 Full Production Setup

For production use with proper authentication:

```bash
# Build and start all services
docker compose up --build

# In another terminal, run database migrations
make migrate-up

# Seed sample data (optional)
make seed

# Run smoke tests
make demo-smoke
```

Configure OIDC authentication by setting these environment variables:
- `VITE_OIDC_ISSUER` - Your OIDC provider issuer URL
- `VITE_OIDC_CLIENT_ID` - Your OIDC client ID

### 🧪 Testing & Validation

```bash
# Run all tests
make test

# Run smoke test (end-to-end validation)
make demo-smoke

# Seed sample data
make seed
```

## Development

### Prerequisites

- Docker and Docker Compose
- Go 1.21+
- Node.js 20+
- Make
- golang-migrate (for database migrations): `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

### Directory Structure

```
.
├── api/              # Go API service
│   └── cmd/server/   # Main entry point
├── ui/               # React SPA (Vite)
├── db/               # Database migrations and schemas
│   └── migrations/   # SQL migrations
├── docs/             # Architecture and runbooks
└── docker-compose.yml
```

### Common Commands

```bash
# Development
make dev              # Start development environment
make test             # Run all tests (Go + UI)
make lint             # Run linters

# Database
make migrate-up       # Apply all pending migrations
make migrate-down     # Rollback last migration
make migrate-create   # Create new migration (name=...)

# Utilities
make seed             # Seed sample data
make demo-smoke       # Run end-to-end smoke test
make help             # Show all available commands
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for detailed architecture, API endpoints, and data model.

## Tasks & Milestones

See [docs/tasks.md](docs/tasks.md) for the implementation roadmap.

## License

MIT

## Security

⚠️ **Demo Mode Warning**: The default configuration includes a demo API key for development only. Never use demo keys in production.
