# Open Exposure Management - Operations Guide

Complete operational manual for deploying, managing, and troubleshooting the OEM platform.

## Table of Contents

1. [Reference Architecture](#reference-architecture)
2. [Initial Deployment](#initial-deployment)
3. [Service Management](#service-management)
4. [Authentication Configuration](#authentication-configuration)
5. [Backup & Restore](#backup--restore)
6. [Troubleshooting](#troubleshooting)
7. [Development Workflow](#development-workflow)

---

## Reference Architecture

### Services

The OEM platform runs as a set of interconnected Docker containers on a single host:

**oem-postgres** (PostgreSQL 16)
- **Port**: 5432
- **Purpose**: System of record for all data
- **Credentials**: `oem/password` (default - change for production)
- **Volume**: `postgres-data` persists all database data
- **Stores**: tenants, users, assets, findings, threat intel cache

**oem-api** (Go 1.21+)
- **Port**: 8080
- **Purpose**: REST API service
- **Handles**: ingestion, queries, auth, intel sync
- **Environment**: `DEMO_MODE=true` disables authentication
- **Depends on**: postgres health check

**oem-ui** (React SPA + nginx)
- **Port**: 80
- **Purpose**: Frontend web interface
- **Serves**: dashboard, asset inventory, findings list
- **Communicates**: with API via `/api/v1`
- **Static files**: built into container

**oem-pgadmin** (pgAdmin 4)
- **Port**: 5050
- **Purpose**: Database management UI
- **Access**: `admin@oem.local` / `admin`
- **Volume**: `pgadmin-data` for saved sessions

### Network Architecture

All services communicate via `oem-network` bridge network:
- API connects to postgres using internal DNS (`postgres:5432`)
- UI reaches API via environment-configured base URL
- No external network access required for inter-service communication

### Data Flow

1. **Ingestion**: External scanners → API (port 8080) → postgres
2. **Queries**: UI (port 80) → API → postgres
3. **Intel Sync**: API scheduled jobs → NVD/EPSS/KEV APIs → postgres cache

### Storage Locations

- **Database data**: `postgres-data` Docker volume
- **pgAdmin sessions**: `pgadmin-data` Docker volume
- **Backups**: Created to `./backups/` directory on host

---

## Initial Deployment

### Prerequisites Checklist

Before deploying, ensure you have:
- Docker 20.10+ and Docker Compose v2+ installed
- At least 4GB RAM available for containers
- Ports 80, 8080, 5432, and 5050 available
- Git to clone the repository

### One-Command Deployment

The `setup.sh` script automates everything:

```bash
git clone <repository-url>
cd openexposuremanagement
./setup.sh
```

This script:
1. Checks Docker/Docker Compose availability
2. Creates required directories (backups/, logs/)
3. Starts all services with `docker compose up --build`
4. Applies database migrations
5. Seeds sample data (optional)
6. Runs health checks

### Manual Deployment

If you prefer manual setup:

```bash
# Create backup directory
mkdir -p backups logs

# Start services
docker compose up --build -d

# Wait for postgres to be healthy
docker exec oem-postgres pg_isready -U oem

# Apply migrations
cd api
export DATABASE_URL="postgres://oem:password@localhost:5432/oem?sslmode=disable"
go run ./cmd/server migrate up
cd ..

# Verify deployment
curl -f http://localhost:8080/healthz
```

### Verification Steps

After deployment completes, verify each service:

```bash
# 1. Check all containers are healthy
docker compose ps

# Expected output: All services show "healthy" or "running"
# NAME            STATUS
# oem-postgres    Up (healthy)
# oem-api         Up (healthy)
# oem-ui          Up
# oem-pgadmin     Up

# 2. Verify API health endpoint
curl -f http://localhost:8080/healthz || echo "API health check failed"

# Expected: {"status":"ok"}

# 3. Check database connectivity
docker exec oem-postgres pg_isready -U oem

# Expected: /var/run/postgresql:5432 - accepting connections

# 4. Verify database tables exist
docker exec oem-postgres psql -U oem -d oem -c "\dt"

# Expected: List of tables including tenants, users, assets, findings, etc.
```

### Access Points

After successful deployment:

- **Frontend**: http://localhost:80 (demo mode auto-logs in)
- **API**: http://localhost:8080
- **PgAdmin**: http://localhost:5050

**PgAdmin Setup:**
1. Log in with `admin@oem.local` / `admin`
2. Add server connection:
   - Host: `postgres`
   - Port: `5432`
   - Database: `oem`
   - Username: `oem`
   - Password: `password`

### Run Smoke Tests

Validate the deployment with end-to-end tests:

```bash
make demo-smoke
```

This validates:
- Ingestion endpoint accepts sample payload
- Assets appear in database
- Findings are queryable
- Intel status endpoint responds

---

## Service Management

### Starting and Stopping Services

```bash
# Start all services in foreground (logs stream to terminal)
make dev

# Start all services in background
make dev-bg

# Stop all services
make dev-down

# Restart specific service
docker compose restart api

# View service status
docker compose ps
```

### Health Monitoring

Each service exposes health information:

```bash
# API health check
curl http://localhost:8080/healthz

# Expected: {"status":"ok"}

# Database readiness
docker exec oem-postgres pg_isready -U oem

# Expected: accepting connections

# Container health status
docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Health}}"
```

### Log Management

View logs for troubleshooting:

```bash
# All services, last 100 lines
docker compose logs --tail=100

# Follow logs in real-time
docker compose logs -f

# Specific service
docker compose logs -f api
docker compose logs -f postgres

# Last 50 lines of API, then exit
docker compose logs --tail=50 api
```

**Persisting logs to disk:**

```bash
# Create logs directory
mkdir -p logs

# Export logs to file
docker compose logs > logs/docker-$(date +%Y%m%d-%H%M%S).log

# Export specific service
docker compose logs api > logs/api-$(date +%Y%m%d-%H%M%S).log
```

### Common Operational Commands

```bash
# Enter API container for debugging
docker exec -it oem-api sh

# Connect to PostgreSQL directly
docker exec -it oem-postgres psql -U oem -d oem

# Rebuild specific service after code changes
docker compose up --build -d api

# Check resource usage
docker stats

# Clean up old images
docker image prune -a

# View service configuration
docker compose config
```

### Service-Specific Monitoring

**API Service:**
- Monitor `/healthz` endpoint for 200 responses
- Check logs for ingestion errors or auth failures
- CPU/memory usage via `docker stats oem-api`

**PostgreSQL:**
- Use pgAdmin or connect directly for query inspection
- Monitor connection counts: `docker exec oem-postgres psql -U oem -c "SELECT count(*) FROM pg_stat_activity;"`
- Check slow queries via logs

**UI:**
- Check browser console for JavaScript errors
- Verify API connectivity via browser Network tab
- Confirm static assets load (should return 200)

### Automated Health Checks

The system includes materialized view refresh jobs every 5 minutes. Monitor via:

```bash
curl http://localhost:8080/api/v1/intel/status
```

Expected response includes last sync time and status.

---

## Authentication Configuration

### ⚠️ Security Warning

**Demo mode disables authentication entirely. Never use demo mode in production or internet-facing deployments.**

### Demo Mode (Development/Testing Only)

Demo mode is enabled by default in `docker-compose.yml`:

```yaml
environment:
  DEMO_MODE: "true"
  DEMO_API_KEY: "demo-key-for-development-only"
```

**How Demo Mode Works:**
- No authentication required for any endpoint
- Demo API key (`demo-key-for-development-only`) pre-configured
- UI bypasses login screen entirely
- All requests succeed without auth checks

**When to Use:**
- ✅ Local development and testing
- ✅ Internal demonstrations on isolated networks
- ✅ CI/CD pipelines with no external access

**When NOT to Use:**
- ❌ Production environments
- ❌ Any deployment accessible from the internet
- ❌ Systems handling real vulnerability data

### Production Setup with OIDC

#### Step 1: Configure OIDC Provider

Set up an OIDC application with your identity provider (Okta, Auth0, Azure AD, Keycloak):

- **Redirect URI**: `http://localhost:80/callback` (adjust for your domain)
- **Grant Type**: Authorization Code with PKCE
- **Scopes**: `openid`, `profile`, `email`

#### Step 2: Update docker-compose.yml

Replace demo mode configuration with OIDC settings:

```yaml
environment:
  DEMO_MODE: "false"  # Disable demo mode
  # Add OIDC configuration
  VITE_OIDC_ISSUER: "https://your-idp.example.com"
  VITE_OIDC_CLIENT_ID: "your-client-id"
```

#### Step 3: Configure User Roles

After first login, users will be created but need role assignment:

```bash
# Connect to database
docker exec -it oem-postgres psql -U oem -d oem

# Assign admin role to user
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.name = 'admin' AND r.tenant_id = u.tenant_id
WHERE u.email = 'user@example.com';
```

**Available roles:**
- `admin`: Manage tenants, users, roles, configuration
- `analyst`: View data, search findings, create suppression proposals
- `viewer`: Read-only access

### API Key Management

API keys are used by scanners to push findings via the ingestion endpoint.

#### Creating Ingestion API Keys

**Via database SQL:**

```sql
-- Connect to database
docker exec -it oem-postgres psql -U oem -d oem

-- Create API key
INSERT INTO api_keys (tenant_id, name, key_hash, scopes_json, source_binding)
VALUES (
  1,  -- tenant_id (adjust for your tenant)
  'tenable-scanner-prod',
  'scrambled-sha256-hash',  -- Use: echo -n "your-key" | sha256sum
  '["ingest:vm"]',
  'tenable'  -- optional: bind to specific source
);
```

**Generating secure API keys:**

```bash
# Generate random 32-byte key
openssl rand -hex 32

# Example output: a1b2c3d4e5f6... (use this as your API key)
# Hash it for database storage:
echo -n "a1b2c3d4e5f6..." | sha256sum
```

#### API Key Security Best Practices

- Keys are stored as SHA-256 hashes in the database
- Include `ingest:vm` scope for VM finding ingestion
- Optional `source_binding` enforces payload source matches key
- Set `revoked_at` timestamp to disable keys (don't delete - preserves audit trail)
- Rotate keys regularly in production

#### Testing API Keys

```bash
# Test ingestion endpoint
curl -X POST http://localhost:8080/api/v1/ingest/vm/findings \
  -H "X-API-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d @sample-payload.json

# Expected: {"status":"accepted","findings_processed":5}
```

#### Troubleshooting Authentication

```bash
# Verify auth is enabled/disabled
docker compose logs api | grep -i "demo mode"

# Check user creation
docker exec -it oem-postgres psql -U oem -d oem \
  -c "SELECT email, status FROM users;"

# Verify API key exists
docker exec -it oem-postgres psql -U oem -d oem \
  -c "SELECT name, scopes_json, source_binding, revoked_at FROM api_keys;"

# Test API key validity
curl -H "X-API-Key: demo-key-for-development-only" \
  http://localhost:8080/api/v1/assets
```

---

## Backup & Restore

### Backup Strategy Overview

OEM supports **two complementary backup approaches**:

1. **Docker Volume Backups**: Fast, complete snapshots of entire data directory
2. **PostgreSQL Logical Backups**: SQL dumps, portable across PostgreSQL versions

**Recommendation**: Use both for redundancy. Volume backups for quick restores, SQL dumps for long-term archival and portability.

### Docker Volume Backups

#### Creating Volume Snapshots

```bash
# Ensure services are stopped for consistent backup
docker compose stop postgres api

# Create backup directory
mkdir -p backups/volumes

# Backup postgres volume
docker run --rm -v openexposuremanagement_postgres-data:/data \
  -v $(pwd)/backups/volumes:/backup \
  alpine tar czf /backup/postgres-$(date +%Y%m%d-%H%M%S).tar.gz -C /data .

# Backup pgadmin volume (optional - contains saved queries/sessions)
docker run --rm -v openexposuremanagement_pgadmin-data:/data \
  -v $(pwd)/backups/volumes:/backup \
  alpine tar czf /backup/pgadmin-$(date +%Y%m%d-%H%M%S).tar.gz -C /data .

# Restart services
docker compose start postgres api
```

#### Restoring from Volume Backup

```bash
# Stop services
docker compose down

# Remove existing volumes (WARNING: destroys current data)
docker volume rm openexposuremanagement_postgres-data
docker volume rm openexposuremanagement_pgadmin-data

# Recreate volumes
docker volume create openexposuremanagement_postgres-data
docker volume create openexposuremanagement_pgadmin-data

# Extract backup (replace filename with actual backup)
docker run --rm -v openexposuremanagement_postgres-data:/data \
  -v $(pwd)/backups/volumes:/backup \
  alpine tar xzf /backup/postgres-20250112-143022.tar.gz -C /data

# Start services
docker compose up -d
```

### PostgreSQL Logical Backups

#### Creating SQL Dumps

```bash
# Create backup directory
mkdir -p backups/sql

# Dump entire database in custom format (compressed)
docker exec oem-postgres pg_dump -U oem -h localhost \
  -F c -f /tmp/oem-backup.dump oem

# Copy dump to host
docker cp oem-postgres:/tmp/oem-backup.dump \
  backups/sql/oem-$(date +%Y%m%d-%H%M%S).dump

# Clean up temporary file
docker exec oem-postgres rm /tmp/oem-backup.dump
```

**Backup format options:**
- `-F c`: Custom format (compressed, includes TOC, portable) - **Recommended**
- `-F d`: Directory format (parallel restore capable)
- `-F p`: Plain text SQL (human-readable, portable)

#### Restoring from SQL Dump

```bash
# Copy dump to container
docker cp backups/sql/oem-20250112-143022.dump oem-postgres:/tmp/restore.dump

# Stop API to prevent connections during restore
docker compose stop api

# Drop and recreate database
docker exec oem-postgres psql -U oem -c "DROP DATABASE IF EXISTS oem;"
docker exec oem-postgres psql -U oem -c "CREATE DATABASE oem;"

# Restore from dump (parallel jobs for speed)
docker exec oem-postgres pg_restore -U oem -d oem -j 4 /tmp/restore.dump

# Clean up and restart
docker exec oem-postgres rm /tmp/restore.dump
docker compose start api
```

### Automated Backup Script

Create `scripts/backup.sh`:

```bash
#!/bin/bash
set -e

BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p "$BACKUP_DIR/volumes"
mkdir -p "$BACKUP_DIR/sql"

echo "[$DATE] Starting backup..."

# Volume backup
echo "[$DATE] Creating volume snapshots..."
docker compose stop postgres api
docker run --rm -v openexposuremanagement_postgres-data:/data \
  -v $(pwd)/"$BACKUP_DIR/volumes":/backup \
  alpine tar czf /backup/postgres-$DATE.tar.gz -C /data .
docker compose start postgres api

# SQL dump
echo "[$DATE] Creating SQL dump..."
docker exec oem-postgres pg_dump -U oem -h localhost \
  -F c -f /tmp/oem.dump oem
docker cp oem-postgres:/tmp/oem.dump \
  "$BACKUP_DIR/sql/oem-$DATE.dump"
docker exec oem-postgres rm /tmp/oem.dump

echo "[$DATE] Backup complete: $BACKUP_DIR"
ls -lh "$BACKUP_DIR/volumes" "$BACKUP_DIR/sql"
```

Make it executable:
```bash
chmod +x scripts/backup.sh
```

**Schedule with cron** (Linux/Mac):

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * cd /path/to/openexposuremanagement && ./scripts/backup.sh >> logs/backup.log 2>&1
```

### Verification Procedures

**Always verify backups by test-restore:**

```bash
# Test restore to temporary database
docker exec oem-postgres psql -U oem -c "CREATE DATABASE oem_restore_test;"

# Restore backup to test database
docker cp backups/sql/oem-20250112-143022.dump oem-postgres:/tmp/test-restore.dump
docker exec oem-postgres pg_restore -U oem -d oem_restore_test /tmp/test-restore.dump

# Verify tables exist and have data
docker exec oem-postgres psql -U oem -d oem_restore_test \
  -c "SELECT COUNT(*) FROM assets;"

# Expected: Number of assets in backup

# Clean up
docker exec oem-postgres psql -U oem -c "DROP DATABASE oem_restore_test;"
docker exec oem-postgres rm /tmp/test-restore.dump

echo "Backup verification complete"
```

### Backup Retention

**Recommended retention schedule:**
- Daily backups: Keep 7 days
- Weekly backups: Keep 4 weeks
- Monthly backups: Keep 12 months

**Automated cleanup:**

```bash
# Cleanup old backups (keep last 7 days)
find backups/ -name "*.tar.gz" -mtime +7 -delete
find backups/ -name "*.dump" -mtime +7 -delete

# Add to cron for automatic cleanup
0 3 * * * find /path/to/openexposuremanagement/backups/ -name "*.tar.gz" -mtime +7 -delete
0 3 * * * find /path/to/openexposuremanagement/backups/ -name "*.dump" -mtime +7 -delete
```

---

## Troubleshooting

### Service Startup Problems

#### Problem: Services fail to start with "port already in use"

**Symptoms:**
```
Error: bind: address already in use
```

**Diagnosis:**
```bash
# Check what's using the port
lsof -i :80   # UI
lsof -i :8080 # API
lsof -i :5432 # PostgreSQL
```

**Solutions:**
- Stop conflicting service (nginx, apache, local postgres)
- Or change port in `docker-compose.yml`:
  ```yaml
  services:
    ui:
      ports:
        - "8081:80"  # Use port 8081 instead of 80
  ```

#### Problem: PostgreSQL container restarts repeatedly

**Symptoms:**
```bash
docker compose ps
# oem-postgres   Restarting (1) 5 seconds ago
```

**Diagnosis:**
```bash
# Check logs
docker compose logs postgres | tail -50
```

**Common causes:**
- Volume corruption
- Permission issues
- Insufficient disk space

**Solution: Reset volume (WARNING: destroys data)**
```bash
docker compose down
docker volume rm openexposuremanagement_postgres-data
docker compose up -d
```

#### Problem: API container crashes on startup

**Symptoms:**
```
oem-api exited with code 1
```

**Diagnosis:**
```bash
# Check logs for errors
docker compose logs api

# Look for database connection errors
docker compose logs api | grep -i "error"
```

**Solution:**
```bash
# Verify postgres is healthy
docker exec oem-postgres pg_isready -U oem

# Check migrations applied
docker exec oem-postgres psql -U oem -d oem -c "\dt"

# If tables missing, apply migrations
cd api && go run ./cmd/server migrate up && cd ..
```

### Performance Issues

#### Problem: Slow dashboard queries

**Symptoms:**
- Dashboard takes >5 seconds to load
- Browser shows "pending" on API calls

**Diagnosis:**
```bash
# Check materialized view refresh status
docker exec oem-postgres psql -U oem -d oem \
  -c "SELECT schemaname, matviewname, last_refresh FROM pg_matviews;"
```

**Solution:**
```bash
# Manually refresh materialized views
docker exec oem-postgres psql -U oem -d oem \
  -c "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_dashboard_counts;"

# Check if scheduled refresh job is running
curl http://localhost:8080/api/v1/intel/status
```

#### Problem: High memory usage

**Symptoms:**
```bash
docker stats
# oem-postgres: 2GB+ memory usage
```

**Solution:** Tune PostgreSQL memory in `docker-compose.yml`:
```yaml
services:
  postgres:
    command: postgres -c shared_buffers=256MB -c effective_cache_size=1GB
```

### Authentication Issues

#### Problem: API returns 401/403 errors

**Symptoms:**
```bash
curl http://localhost:8080/api/v1/assets
# {"error":"unauthorized"}
```

**Diagnosis:**
```bash
# Verify demo mode status
docker compose logs api | grep "DEMO_MODE"

# If demo mode should be enabled but isn't:
docker compose down
# Edit docker-compose.yml to set DEMO_MODE="true"
docker compose up -d
```

**For production (OIDC):**
```bash
# Check API key exists and isn't revoked
docker exec oem-postgres psql -U oem -d oem \
  -c "SELECT name, scopes_json, revoked_at FROM api_keys;"

# Test with demo key (if demo mode enabled)
curl -H "X-API-Key: demo-key-for-development-only" \
  http://localhost:8080/api/v1/assets
```

#### Problem: OIDC login redirects incorrectly

**Symptoms:**
- Browser redirects to wrong URL
- Login loop (redirects back to login)

**Diagnosis:**
```bash
# Verify issuer URL in API logs
docker compose logs api | grep "issuer"

# Check UI environment variables
docker compose config | grep VITE_OIDC
```

**Solution:**
- Verify OIDC issuer URL is correct in `docker-compose.yml`
- Check redirect URI matches in OIDC provider configuration
- Browser console may show CORS or redirect errors (F12)

### Data Issues

#### Problem: Assets not appearing after ingestion

**Symptoms:**
```bash
curl -X POST /api/v1/ingest/vm/findings -d @payload.json
# Returns "accepted" but assets count is 0
```

**Diagnosis:**
```bash
# Check if data was ingested
docker exec oem-postgres psql -U oem -d oem \
  -c "SELECT COUNT(*) FROM assets;"

# Check ingestion logs for errors
docker compose logs api | grep -A 5 "ingest"

# Verify API key scope
docker exec oem-postgres psql -U oem -d oem \
  -c "SELECT name, scopes_json FROM api_keys;"
```

**Solution:**
- Ensure API key has `ingest:vm` scope
- Check payload format matches expected schema
- Verify `source` field matches API key's `source_binding` if set

#### Problem: Intel not enriching findings

**Symptoms:**
- Findings show no CVE descriptions
- EPSS/KEV fields are null

**Diagnosis:**
```bash
# Check last sync run
curl http://localhost:8080/api/v1/intel/status

# Check intel_cve table has data
docker exec oem-postgres psql -U oem -d oem \
  -c "SELECT COUNT(*) FROM intel_cve;"
```

**Solution:**
```bash
# Manual refresh
curl -X POST http://localhost:8080/api/v1/intel/refresh

# Check sync logs for errors
docker compose logs api | grep -i "intel.*sync"
```

### Database Connection Issues

#### Problem: "connection refused" errors

**Symptoms:**
```
Error: connect: connection refused
```

**Diagnosis:**
```bash
# Verify postgres container is running
docker compose ps postgres

# Check health status
docker inspect oem-postgres | jq '.[0].State.Health.Status'
```

**Solution:**
```bash
# Restart if unhealthy
docker compose restart postgres

# Wait for health check
docker exec oem-postgres pg_isready -U oem
```

#### Problem: "too many connections"

**Symptoms:**
```
FATAL: remaining connection slots are reserved
```

**Diagnosis:**
```bash
# Check connection count
docker exec oem-postgres psql -U oem -d oem \
  -c "SELECT count(*) FROM pg_stat_activity;"

# Check max connections setting
docker exec oem-postgres psql -U oem -d oem \
  -c "SHOW max_connections;"
```

**Solution:**
```bash
# Kill long-running queries if needed
docker exec oem-postgres psql -U oem -d oem \
  -c "SELECT pid, query, state FROM pg_stat_activity WHERE state != 'idle' AND query != '';"

# Terminate specific connection (use PID from above)
docker exec oem-postgres psql -U oem -d oem \
  -c "SELECT pg_terminate_backend(12345);"
```

### Getting Help

When troubleshooting fails, collect diagnostic information:

**1. Export logs and status:**
```bash
mkdir -p logs/debug

docker compose logs > logs/debug/docker-$(date +%Y%m%d-%H%M%S).log
docker compose ps > logs/debug/status-$(date +%Y%m%d-%H%M%S).txt
docker stats --no-stream > logs/debug/metrics-$(date +%Y%m%d-%H%M%S).txt
```

**2. Verify environment:**
```bash
docker --version
docker compose version
docker compose config
```

**3. Check documentation:**
- `README.md` - Quick start and common issues
- `TESTING.md` - Test database issues
- `docs/architecture.md` - Architecture reference

**4. Search existing issues:**
- Check GitHub Issues for similar problems
- Review error messages in logs

---

## Development Workflow

### Local Development Setup

For developers contributing to OEM, you can run services locally instead of Docker:

**Prerequisites:**
- Go 1.21+
- Node.js 20+
- PostgreSQL 16+ running locally on port 5432

#### Backend Development (Go)

```bash
# Set database URL
export DATABASE_URL="postgres://oem:password@localhost:5432/oem?sslmode=disable"

# Run migrations
cd api
make migrate-up

# Run API server in demo mode
DEMO_MODE=true go run ./cmd/server

# Run tests
go test ./...

# Run specific package with verbose output
go test -v ./internal/ingest

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

#### Frontend Development (React)

```bash
# Install dependencies
cd ui
npm install

# Start dev server (Vite with hot reload)
npm run dev

# Build for production
npm run build

# Run tests
npm test

# Type checking
npm run type-check
```

### Making Configuration Changes

#### Adding Environment Variables

1. Add variable to `docker-compose.yml`:
```yaml
services:
  api:
    environment:
      NEW_SETTING: "value"
```

2. Access in Go code:
```go
value := os.Getenv("NEW_SETTING")
```

3. Rebuild and restart:
```bash
docker compose up --build -d api
```

#### Updating Database Schema

```bash
# Create new migration
make migrate-create name=add_new_column

# Edit generated migration file
vim db/migrations/YYYYMMDDHHMMSS_add_new_column.up.sql
vim db/migrations/YYYYMMDDHHMMSS_add_new_column.down.sql

# Apply migration
make migrate-up

# Verify
docker exec oem-postgres psql -U oem -d oem -c "\d table_name"
```

### Testing Workflow

**Before committing changes:**

```bash
# Run all tests
make test

# Run integration tests only
make test-integration

# Run with verbose output
make test-verbose

# Run linters
make lint
```

**Integration test setup:**

```bash
# Start services in background
make dev-bg

# Create test database
make db-test-setup

# Run integration tests
make test-integration

# Cleanup
make dev-down
```

### Debugging Tips

#### API Debugging

```bash
# Enable debug logging
# Edit docker-compose.yml: LOG_LEVEL=debug
docker compose restart api

# Follow logs
docker compose logs -f api

# Attach debugger (if using IDE):
# Set breakpoints in ./api directory and run cmd/server

# Enable query logging in PostgreSQL
docker exec oem-postgres psql -U oem -c "ALTER SYSTEM SET log_statement = 'all';"
docker compose restart postgres
```

#### Frontend Debugging

```bash
# Vite dev server includes hot module reload
npm run dev

# Browser DevTools:
# - React DevTools extension for component inspection
# - Network tab for API calls
# - Console for JavaScript errors

# Test production build locally
npm run build
npx serve dist
```

### Code Organization

```
api/
├── cmd/server/         # Application entry point
├── internal/           # Private application code
│   ├── auth/          # Authentication & authorization
│   ├── database/      # Database layer & test utilities
│   ├── ingest/        # VM finding ingestion
│   └── api/           # HTTP handlers & middleware
└── pkg/               # Public libraries (if any)

ui/
├── src/
│   ├── components/    # React components
│   ├── pages/         # Page components
│   ├── api/           # API client functions
│   └── hooks/         # Custom React hooks

db/
└── migrations/        # SQL migrations
```

### Common Development Tasks

#### Adding a new API endpoint

1. Define handler in `api/internal/api/`
2. Add route in router setup
3. Add authentication middleware if needed
4. Write tests in `*_test.go`
5. Update TypeScript types in `ui/src/api/`

#### Adding database query

1. Create migration if schema changes needed
2. Add query function to appropriate repository
3. Write test with test database
4. Update API to use new query

#### Modifying UI components

1. Update component in `ui/src/components/`
2. Update types if data shape changes
3. Test with both dev server and production build
4. Check for TypeScript errors: `npm run type-check`

### Git Workflow Conventions

```bash
# Feature branch workflow
git checkout -b feature/add-new-endpoint

# Make changes...
make test  # Verify tests pass

# Commit with conventional commits
git commit -m "feat: add new endpoint for X"

# Commit message prefixes:
# feat:     New feature
# fix:      Bug fix
# docs:     Documentation changes
# refactor: Code refactoring
# test:     Test changes
# chore:    Maintenance tasks
```

### Performance Testing

```bash
# Load test ingestion endpoint
# Install: go install github.com/rakyll/hey@latest
hey -n 1000 -c 10 -H "X-API-Key: demo-key-for-development-only" \
  -m POST -H "Content-Type: application/json" \
  -D payload.json http://localhost:8080/api/v1/ingest/vm/findings

# Monitor database during load
docker exec oem-postgres psql -U oem -d oem \
  -c "SELECT count(*) FROM pg_stat_activity;"
```

---

## Quick Reference

### Essential Commands

```bash
# Start all services
make dev-bg

# Stop all services
make dev-down

# Run tests
make test

# Create backup
./scripts/backup.sh

# View logs
docker compose logs -f

# Database shell
docker exec -it oem-postgres psql -U oem -d oem

# API health check
curl http://localhost:8080/healthz
```

### Important URLs

- Frontend: http://localhost:80
- API: http://localhost:8080
- PgAdmin: http://localhost:5050
- API Health: http://localhost:8080/healthz
- Intel Status: http://localhost:8080/api/v1/intel/status

### File Locations

- Configuration: `docker-compose.yml`
- Migrations: `db/migrations/`
- Backups: `backups/`
- Logs: `logs/`

### Default Credentials

- PostgreSQL: `oem` / `password`
- PgAdmin: `admin@oem.local` / `admin`
- Demo API Key: `demo-key-for-development-only`

**⚠️ Change these for production deployments!**

---

For more information, see:
- [README.md](../README.md) - Quick start guide
- [TESTING.md](../TESTING.md) - Testing guide
- [docs/architecture.md](architecture.md) - Architecture reference
- [docs/tasks.md](tasks.md) - Implementation roadmap
