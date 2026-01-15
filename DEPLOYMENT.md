# Deployment Configuration

This document explains how to configure the API base URL for different deployment scenarios.

## Environment Variables

### UI Configuration

The UI uses the `API_BASE_URL` environment variable to proxy API requests to the backend.

- **Variable**: `API_BASE_URL`
- **Default**: `http://localhost:8080`
- **Used by**: Vite dev server proxy configuration

### API Configuration

The API uses the following environment variables:

- **API_PORT**: Port for the API server (default: `8080`)
- **DATABASE_URL**: PostgreSQL connection string
- **DEMO_MODE**: Enable demo mode without authentication
- **SWAGGER_ENABLED**: Enable/disable Swagger UI (default: `true`)
- **OIDC_ISSUER**: OIDC provider URL
- **OIDC_CLIENT_ID**: OIDC client ID

## Deployment Scenarios

### 1. Full Docker Compose (Default)

When running the full stack with `docker-compose up`:

```yaml
# docker-compose.yml (automatically configured)
ui:
  environment:
    API_BASE_URL: http://api:8080  # Uses docker service name
```

**How it works**: 
- UI container communicates with API container using docker service names
- No port mapping needed for API (only internal communication)
- Both services run on `oem-network` docker network

### 2. Local UI Development + Docker Backend

When developing the UI locally (`npm run dev`) while running the backend in docker:

```bash
# In docker-compose.yml, change the UI environment:
ui:
  environment:
    API_BASE_URL: http://localhost:8080  # Use localhost
```

Or create a local `.env` file in the `ui/` directory:

```bash
# ui/.env
API_BASE_URL=http://localhost:8080
```

**How it works**:
- UI runs on your host machine at `http://localhost:3000`
- Vite proxies `/api` and `/swagger` requests to `http://localhost:8080` (API container)
- API port `8080` is mapped to host in docker-compose.yml

### 3. Full Local Development

When running both UI and API locally (without docker):

```bash
# Terminal 1: Start API
cd api
go run ./cmd/server

# Terminal 2: Start UI
cd ui
npm run dev
```

No configuration needed - the defaults work:
- API: `http://localhost:8080`
- UI: `http://localhost:3000` with proxy to `http://localhost:8080`

### 4. Production Deployment

For production, configure the API base URL to point to your production backend:

```bash
# Build UI with production API URL
cd ui
export API_BASE_URL=https://api.example.com
npm run build
```

Or configure in your CI/CD pipeline:

```yaml
# Example GitHub Actions
- name: Build UI
  run: |
    cd ui
    export API_BASE_URL=https://api.example.com
    npm run build
  env:
    API_BASE_URL: https://api.example.com
```

## Port Configuration

To change the default ports:

### Change API Port (default: 8080)

```yaml
# docker-compose.yml
api:
  environment:
    API_PORT: 9000  # Change to desired port
  ports:
    - "9000:9000"  # Update port mapping
```

Then update UI configuration:

```yaml
# docker-compose.yml
ui:
  environment:
    API_BASE_URL: http://api:9000  # Update port
```

### Change UI Port (default: 80 for production, 3000 for dev)

**Production** (docker):
```yaml
# docker-compose.yml
ui:
  ports:
    - "8080:80"  # Change host port to 8080
```

**Development** (local):
```typescript
// vite.config.ts
export default defineConfig({
  server: {
    port: 3001,  // Change to desired port
  },
})
```

## Swagger UI Access

Swagger UI is available at:

- **Local development**: `http://localhost:3000/swagger/`
- **Docker compose**: `http://localhost/swagger/` (or your configured UI port)
- **Direct API access**: `http://localhost:8080/swagger/`

To disable Swagger UI, set `SWAGGER_ENABLED=false` in the API environment:

```yaml
# docker-compose.yml
api:
  environment:
    SWAGGER_ENABLED: "false"
```

## Troubleshooting

### UI can't reach API

1. Check if API is running: `curl http://localhost:8080/healthz`
2. Check `API_BASE_URL` in UI configuration
3. Check docker logs: `docker-compose logs api`
4. Verify network connectivity (if using docker service names)

### Swagger UI shows 404

1. Verify `SWAGGER_ENABLED` is `true` (default)
2. Check API logs for swagger registration message
3. Try accessing swagger directly: `curl http://localhost:8080/swagger/`

### Port conflicts

If port 8080 is already in use:

```bash
# macOS
lsof -i :8080

# Linux
netstat -tulpn | grep 8080

# Then change the port in docker-compose.yml or run the service on a different port
```
