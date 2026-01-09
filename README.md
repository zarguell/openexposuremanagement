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

## Development

### Prerequisites

- Docker and Docker Compose
- Go 1.21+
- Node.js 20+
- Make

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
