#!/bin/bash

# OEM Docker Setup Script
# Handles database setup and service startup for Open Exposure Management

set -e  # Exit on any error

echo "🚀 Setting up Open Exposure Management with Docker Compose"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if docker and docker-compose are available
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not available. Please install Docker Compose.${NC}"
    exit 1
fi

# Use 'docker compose' (new syntax) if available, otherwise 'docker-compose'
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo -e "${BLUE}📦 Starting PostgreSQL database...${NC}"
$DOCKER_COMPOSE up -d postgres

echo -e "${BLUE}⏳ Waiting for database to be ready...${NC}"
# Wait for postgres to be healthy
timeout=60
counter=0
while [ $counter -lt $timeout ]; do
    if $DOCKER_COMPOSE exec -T postgres pg_isready -U oem -d oem &> /dev/null; then
        echo -e "${GREEN}✅ Database is ready!${NC}"
        break
    fi
    counter=$((counter + 1))
    sleep 1
done

if [ $counter -eq $timeout ]; then
    echo -e "${RED}❌ Database failed to start within ${timeout} seconds${NC}"
    exit 1
fi

echo -e "${BLUE}🗄️ Running database migrations...${NC}"
# Check if migrate is available
if command -v migrate &> /dev/null; then
    MIGRATE_CMD="migrate"
elif [ -f "$HOME/go/bin/migrate" ]; then
    MIGRATE_CMD="$HOME/go/bin/migrate"
else
    echo -e "${YELLOW}⚠️ Installing golang-migrate...${NC}"
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    MIGRATE_CMD="$HOME/go/bin/migrate"
fi

# Run migrations
if $MIGRATE_CMD -path db/migrations -database "postgres://oem:password@localhost:5432/oem?sslmode=disable" up; then
    echo -e "${GREEN}✅ Database migrations completed successfully!${NC}"
else
    echo -e "${RED}❌ Database migrations failed${NC}"
    exit 1
fi

echo -e "${BLUE}🏗️ Starting application services...${NC}"
$DOCKER_COMPOSE up -d api ui pgadmin

echo -e "${BLUE}⏳ Waiting for services to be ready...${NC}"
sleep 5

# Check if services are running
if $DOCKER_COMPOSE ps | grep -q "Up"; then
    echo ""
    echo -e "${GREEN}🎉 Open Exposure Management is now running!${NC}"
    echo ""
    echo -e "${GREEN}📱 Frontend:${NC} http://localhost:80"
    echo -e "${GREEN}🔧 Backend API:${NC} http://localhost:8080"
    echo -e "${GREEN}🗄️ Database Admin:${NC} http://localhost:5050 (admin@oem.local / admin)"
    echo ""
    echo -e "${YELLOW}⚠️ Demo Mode:${NC} Authentication is disabled for demonstration"
    echo -e "${YELLOW}🔒 Security:${NC} This is NOT secure for production use!"
    echo ""
    echo -e "${BLUE}🛑 To stop:${NC} $DOCKER_COMPOSE down"
    echo -e "${BLUE}📊 View logs:${NC} $DOCKER_COMPOSE logs -f"
else
    echo -e "${RED}❌ Some services failed to start. Check logs with: $DOCKER_COMPOSE logs${NC}"
    exit 1
fi