# Testing Guide

## Quick Start

### One-time Setup

```bash
# 1. Start the development environment
make dev-bg

# 2. Set up the test database (one-time)
make db-test-setup

# 3. Run tests
make test-integration
```

That's it! The test configuration is now built into the code.

## Running Tests

### Run all tests (Go + UI)
```bash
make test
```

### Run only unit tests (fast, no database)
```bash
make test-unit
```

### Run integration tests (requires test database)
```bash
make test-integration
```

### Run with verbose output
```bash
make test-verbose
```

### Run specific package tests
```bash
cd api
go test ./internal/ingest -v
```

### Run specific test
```bash
cd api
go test ./internal/ingest -run TestUpsertAsset -v
```

## Test Database Management

### Create/recreate test database
```bash
make db-test-setup    # Create if not exists
make db-test-reset    # Drop and recreate
```

### Open psql shell to test database
```bash
make db-test-shell
```

## Configuration

Test database connection is configured in `api/internal/database/testutil.go`:

```go
// Default matches docker-compose configuration
testDBURL := "postgres://oem:password@localhost:5432/oem_test?sslmode=disable"
```

You can override with environment variable:
```bash
export TEST_DATABASE_URL="postgres://user:pass@host:port/dbname"
go test ./...
```

## Test Cleanup

Tests automatically clean up after themselves:
- Each test gets a fresh database state
- Tables are truncated after each test
- Test tenants are created with unique names using timestamps

## Writing Tests

### Using the test database

```go
func TestMyFeature(t *testing.T) {
    db := database.SetupTestDB(t)
    if db == nil {
        return // Test was skipped (DB not available)
    }

    ctx := context.Background()
    tenantID := database.CreateTestTenant(t, db)

    // Your test code here...
}
```

### Test structure

```go
t.Run("descriptive test name", func(t *testing.T) {
    // Setup
    tenantID := database.CreateTestTenant(t, db)

    // Execute
    result, err := MyFunction(ctx, db, tenantID)

    // Assert
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
})
```

## Troubleshooting

### Tests skip with "cannot connect to test database"
```bash
# Start docker compose
make dev-bg

# Create test database
make db-test-setup
```

### Port already in use
```bash
# Check what's using port 5432
lsof -i :5432

# Stop docker compose if running
make dev-down
```

### Test database schema out of date
```bash
# Reset test database (will copy latest schema)
make db-test-reset
```

### Tests fail with "relation does not exist"
```bash
# Run migrations on main database
cd api
make migrate-up

# Then reset test database
make db-test-reset
```

## CI/CD Integration

In CI, run:
```bash
make dev-bg        # Start services
make db-test-setup # Create test DB
make test          # Run all tests
make dev-down      # Cleanup
```

## Test Coverage

```bash
cd api
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Target: ≥80% coverage for new code
