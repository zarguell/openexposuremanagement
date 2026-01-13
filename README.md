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

#### Docker Compose (Recommended)
```bash
# One command sets up everything automatically
./setup.sh

# Access points:
# Frontend: http://localhost:80
# Backend API: http://localhost:8080
# PgAdmin: http://localhost:5050 (admin@oem.local / admin)
```

#### Local Development (For development only)
```bash
# Requires local PostgreSQL
# Start both frontend and backend in demo mode
./demo.sh

# Or manually:
# Terminal 1 - Backend (requires PostgreSQL running)
cd api && DEMO_MODE=true go run ./cmd/server

# Terminal 2 - Frontend
cd ui && npm run dev
```

**⚠️ Security Warning**: Demo mode disables authentication entirely. This is NOT secure for production use!

### 🔐 Full Production Setup

For production use with proper authentication:

```bash
# Build and start all services (includes database setup)
docker compose up --build

# Run database migrations
export DATABASE_URL="postgres://oem:password@localhost:5432/oem?sslmode=disable"
~/go/bin/migrate -path db/migrations -database "$(DATABASE_URL)" up

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
- Go 1.21+ (for local development and migrations)
- Node.js 20+ (for local development)
- Make
- golang-migrate (automatically installed by setup.sh): `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

### Troubleshooting

#### Setup Issues
- **Setup script fails**: Ensure Docker and Docker Compose are installed and running
- **Port conflicts**: The setup script uses standard ports (80, 8080, 5432, 5050). Stop conflicting services or modify ports in docker-compose.yml

#### Service Issues
- **Services not starting**: Check logs with `docker compose logs`
- **PgAdmin not accessible**: Wait a moment for initialization, then visit http://localhost:5050 (admin@oem.local / admin)
- **Dashboard not loading**: Clear browser cache or try a hard refresh (Ctrl+F5)

#### Development Issues
- **Local demo script fails**: Ensure PostgreSQL is running locally, or use Docker setup instead
- **Build failures**: Ensure Go 1.21+ and Node.js 20+ are installed for local development

#### Common Commands
```bash
# Check database connectivity
docker exec -it oem-postgres psql -U oem -d oem

# View migration status
docker exec oem-postgres psql -U oem -d oem -c "\dt"

# Reset everything (CAUTION: destroys data)
docker compose down -v
docker volume rm openexposuremanagement_postgres-data openexposuremanagement_pgadmin-data

# Create backup
./scripts/backup.sh

# View service logs
docker compose logs -f
```

**For comprehensive operational guidance, see [docs/operations.md](docs/operations.md)**

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
DATABASE_URL="postgres://oem:password@localhost:5432/oem?sslmode=disable" ~/go/bin/migrate -path db/migrations -database "$(DATABASE_URL)" up    # Apply migrations
DATABASE_URL="postgres://oem:password@localhost:5432/oem?sslmode=disable" ~/go/bin/migrate -path db/migrations -database "$(DATABASE_URL)" down  # Rollback
DATABASE_URL="postgres://oem:password@localhost:5432/oem?sslmode=disable" ~/go/bin/migrate -path db/migrations -database "$(DATABASE_URL)" create -ext sql -dir db/migrations <name>  # Create migration

# Quick setup
./setup.sh            # Full Docker setup with database, migrations, and all services
./demo.sh             # Local development setup (requires local PostgreSQL)

# Utilities
make seed             # Seed sample data
make demo-smoke       # Run end-to-end smoke test
make help             # Show all available commands
```

## Documentation

- **[docs/architecture.md](docs/architecture.md)** - Detailed architecture, API endpoints, and data model
- **[docs/operations.md](docs/operations.md)** - Complete operational guide for deployment, management, and troubleshooting
- **[docs/tasks.md](docs/tasks.md)** - Implementation roadmap and task tracking
- **[TESTING.md](TESTING.md)** - Testing guide and procedures

## Architecture

See [docs/architecture.md](docs/architecture.md) for detailed architecture, API endpoints, and data model.

## Tasks & Milestones

See [docs/tasks.md](docs/tasks.md) for the implementation roadmap.

## License

MIT

## Security

⚠️ **Demo Mode Warning**: The default configuration includes a demo API key for development only. Never use demo keys in production.
