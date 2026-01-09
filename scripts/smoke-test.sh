#!/bin/bash

# OEM End-to-End Demo Smoke Test
# This script brings up the full stack, seeds data, and validates functionality

set -e

echo "🚀 Starting OEM Demo Smoke Test"
echo "================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if docker is available
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi

# Check if docker compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not available"
    exit 1
fi

print_status "Prerequisites check passed"

# Start the stack
echo ""
echo "🏗️  Starting OEM stack..."
if command -v docker-compose &> /dev/null; then
    docker-compose up -d --build
else
    docker compose up -d --build
fi

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 10

# Function to check if service is ready
check_service() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            print_status "$service_name is ready"
            return 0
        fi
        echo "Waiting for $service_name... (attempt $attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done

    print_error "$service_name failed to start"
    return 1
}

# Check backend health
if ! check_service "http://localhost:8080/healthz" "Backend"; then
    print_error "Backend health check failed"
    exit 1
fi

# Check database readiness
if ! check_service "http://localhost:8080/readyz" "Database"; then
    print_error "Database readiness check failed"
    exit 1
fi

# Run database migrations
echo ""
echo "🗄️  Running database migrations..."

# Set DATABASE_URL for migrations (connects to postgres container exposed on host)
export DATABASE_URL="${DATABASE_URL:-postgres://oem:password@localhost:5432/oem?sslmode=disable}"

# Check if migrate is available on host
if command -v migrate &> /dev/null; then
    echo "Using DATABASE_URL: $DATABASE_URL"
    migrate -path db/migrations -database "$DATABASE_URL" up
    print_status "Database migrations completed"
else
    print_error "migrate tool not found on host."
    echo ""
    echo "Please install golang-migrate:"
    echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    echo ""
    echo "Or run migrations manually:"
    echo "  export DATABASE_URL='postgres://oem:password@localhost:5432/oem?sslmode=disable'"
    echo "  make migrate-up"
    echo ""
    exit 1
fi

# Seed sample data
echo ""
echo "🌱 Seeding sample data..."
if [ -f "scripts/seed-data.go" ]; then
    # Wait a bit more for API to be fully ready
    sleep 5
    go run scripts/seed-data.go -api-url="http://localhost:8080" -data-dir="sample-data" -demo
    print_status "Sample data seeded"
else
    print_warning "Seed script not found, skipping data seeding"
fi

# Refresh materialized views
echo ""
echo "🔄 Refreshing dashboard materialized views..."
if curl -s -f -X POST "http://localhost:8080/api/v1/dashboard/refresh" > /dev/null 2>&1; then
    print_status "Materialized views refreshed"
else
    print_error "Failed to refresh materialized views"
    echo "Response:"
    curl -v -X POST "http://localhost:8080/api/v1/dashboard/refresh" 2>&1 | head -20
    exit 1
fi

# Test key API endpoints
echo ""
echo "🧪 Testing API endpoints..."

# Test dashboard
DASHBOARD_RESPONSE=$(curl -s "http://localhost:8080/api/v1/dashboard" 2>/dev/null)
if [ $? -eq 0 ] && echo "$DASHBOARD_RESPONSE" | jq -e '.assets' > /dev/null 2>&1; then
    print_status "Dashboard endpoint responding"

    # Check if we have data
    ASSET_COUNT=$(echo "$DASHBOARD_RESPONSE" | jq -r '.assets.total_assets // 0')
    FINDING_COUNT=$(echo "$DASHBOARD_RESPONSE" | jq -r '.findings.open_count // 0')

    if [ "$ASSET_COUNT" -gt 0 ]; then
        print_status "Dashboard shows $ASSET_COUNT assets"
    else
        print_warning "Dashboard shows 0 assets - data may not be loaded"
    fi

    if [ "$FINDING_COUNT" -gt 0 ]; then
        print_status "Dashboard shows $FINDING_COUNT open findings"
    else
        print_warning "Dashboard shows 0 findings - materialized views may need refresh"
    fi
else
    print_error "Dashboard endpoint failed"
    echo "Response: $DASHBOARD_RESPONSE"
    exit 1
fi

# Test assets
if curl -s -f "http://localhost:8080/api/v1/assets" > /dev/null 2>&1; then
    print_status "Assets endpoint responding"
else
    print_error "Assets endpoint failed"
    exit 1
fi

# Test findings
if curl -s -f "http://localhost:8080/api/v1/findings" > /dev/null 2>&1; then
    print_status "Findings endpoint responding"
else
    print_error "Findings endpoint failed"
    exit 1
fi

# Test intel status
if curl -s -f "http://localhost:8080/api/v1/intel/status" > /dev/null 2>&1; then
    print_status "Intel status endpoint responding"
else
    print_error "Intel status endpoint failed"
    exit 1
fi

# Test data validation
echo ""
echo "📊 Validating seeded data..."

# Check if jq is available for JSON parsing
if command -v jq &> /dev/null; then
    # Check if assets were created
    ASSETS_RESPONSE=$(curl -s "http://localhost:8080/api/v1/assets" 2>/dev/null)
    ASSET_COUNT=$(echo "$ASSETS_RESPONSE" | jq '.pagination.total // 0' 2>/dev/null || echo "0")
    if [ "$ASSET_COUNT" -gt 0 ]; then
        print_status "Assets API validated ($ASSET_COUNT assets found)"
    else
        print_warning "Assets API shows 0 assets"
        echo "Assets response: $ASSETS_RESPONSE"
    fi

    # Check if findings were created
    FINDINGS_RESPONSE=$(curl -s "http://localhost:8080/api/v1/findings" 2>/dev/null)
    FINDING_COUNT=$(echo "$FINDINGS_RESPONSE" | jq '.total // 0' 2>/dev/null || echo "0")
    if [ "$FINDING_COUNT" -gt 0 ]; then
        print_status "Findings API validated ($FINDING_COUNT findings found)"
    else
        print_warning "Findings API shows 0 findings"
        echo "Findings response: $FINDINGS_RESPONSE"
    fi

    # Check raw database counts (if we can connect)
    echo ""
    echo "🔍 Database validation:"
    if command -v psql &> /dev/null; then
        ASSET_DB_COUNT=$(psql "postgres://oem:password@localhost:5432/oem?sslmode=disable" -t -c "SELECT COUNT(*) FROM assets WHERE tenant_id = 1;" 2>/dev/null || echo "error")
        FINDING_DB_COUNT=$(psql "postgres://oem:password@localhost:5432/oem?sslmode=disable" -t -c "SELECT COUNT(*) FROM finding_instances WHERE tenant_id = 1;" 2>/dev/null || echo "error")

        if [ "$ASSET_DB_COUNT" != "error" ] && [ "$FINDING_DB_COUNT" != "error" ]; then
            print_status "Database: $ASSET_DB_COUNT assets, $FINDING_DB_COUNT findings for tenant 1"
        else
            print_warning "Cannot connect to database for validation"
        fi
    else
        print_warning "psql not available - skipping database validation"
    fi
else
    print_warning "jq not available for JSON parsing - skipping detailed validation"
    print_warning "Install jq for better validation: apt-get install jq or brew install jq"
fi

echo ""
echo "🎉 Demo smoke test completed successfully!"
echo ""
echo "🌐 Frontend: http://localhost:3000"
echo "🔧 Backend API: http://localhost:8080"
echo "📊 PgAdmin: http://localhost:5050 (admin@oem.local / admin)"
echo ""
echo "🛑 To stop: make dev-down"

# Keep container running for manual testing
echo ""
echo "Containers will remain running for manual testing..."
echo "Press Ctrl+C to stop all services"

# Wait for interrupt
trap 'echo ""; echo "🛑 Stopping services..."; if command -v docker-compose &> /dev/null; then docker-compose down; else docker compose down; fi' INT TERM
wait