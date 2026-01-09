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
if command -v docker-compose &> /dev/null; then
    docker-compose exec -T api migrate -path db/migrations -database "$DATABASE_URL" up
else
    docker compose exec -T api migrate -path db/migrations -database "$DATABASE_URL" up
fi

print_status "Database migrations completed"

# Seed sample data
echo ""
echo "🌱 Seeding sample data..."
if [ -f "scripts/seed-data.go" ]; then
    cd scripts && go run seed-data.go -api-url="http://localhost:8080" -demo -verbose
    cd ..
    print_status "Sample data seeded"
else
    print_warning "Seed script not found, skipping data seeding"
fi

# Test key API endpoints
echo ""
echo "🧪 Testing API endpoints..."

# Test dashboard
if curl -s -f "http://localhost:8080/api/v1/dashboard" > /dev/null 2>&1; then
    print_status "Dashboard endpoint responding"
else
    print_error "Dashboard endpoint failed"
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

# Check if assets were created
ASSET_COUNT=$(curl -s "http://localhost:8080/api/v1/assets" | jq '.pagination.total // 0' 2>/dev/null || echo "0")
if [ "$ASSET_COUNT" -gt 0 ]; then
    print_status "Assets data validated ($ASSET_COUNT assets found)"
else
    print_warning "No assets found - data seeding may have failed"
fi

# Check if findings were created
FINDING_COUNT=$(curl -s "http://localhost:8080/api/v1/findings" | jq '.total // 0' 2>/dev/null || echo "0")
if [ "$FINDING_COUNT" -gt 0 ]; then
    print_status "Findings data validated ($FINDING_COUNT findings found)"
else
    print_warning "No findings found - data seeding may have failed"
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