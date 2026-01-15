# Testing Patterns

**Analysis Date:** 2026-01-15

## Test Framework

**Runner - Go:**
- Standard Go `testing` package
- github.com/stretchr/testify v1.11.1 for assertions
- Config: `api/go.mod` (test dependencies)

**Runner - TypeScript:**
- Vitest 4.0.16
- Config: `ui/vitest.config.ts`
- Environment: jsdom (DOM testing)

**Assertion Library:**
- Go: testify/assert and testify/require
- TypeScript: Vitest built-in expect (toBe, toEqual, toThrow, etc.)

**Run Commands:**
```bash
make test                # Run all Go tests
make test-unit          # Run Go unit tests (-short flag)
make test-integration   # Run Go integration tests (requires DB)
make test-verbose       # Run Go tests with verbose output
cd ui && npm test        # Run all UI tests once
cd ui && npm test -- --watch  # Watch mode for UI tests
```

## Test File Organization

**Location:**
- Go: Co-located with source files (`{source}_test.go`)
- TypeScript: Co-located with source files (`{source}.test.{ts,tsx}`)
- Go integration tests: Suffix `_integration_test.go`
- No separate test directories

**Naming:**
- Go unit tests: `{function}_test.go` or `{package}_test.go`
- Go integration tests: `{package}_integration_test.go`
- TypeScript tests: `{source}.test.{ts,tsx}`

**Structure:**
```
api/
├── internal/
│   ├── auth/
│   │   ├── apikey.go
│   │   └── apikey_test.go
│   ├── repository/
│   │   ├── asset.go
│   │   └── asset_test.go
ui/src/
├── api/
│   ├── client.ts
│   └── client.test.ts
├── components/
│   ├── QueryBuilder.tsx
│   └── QueryBuilder.test.tsx
```

## Test Structure

**Go Suite Organization:**
```go
func TestAssetRepository_GetByID(t *testing.T) {
    t.Run("returns asset when it exists", func(t *testing.T) {
        // arrange
        db := setupTestDB(t)
        tenantID := database.CreateTestTenant(t, db)
        repo := NewAssetRepository(db)

        // act
        asset, err := repo.GetByID(context.Background(), tenantID, id)

        // assert
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if asset == nil {
            t.Fatal("expected asset, got nil")
        }
    })

    t.Run("returns error when asset doesn't exist", func(t *testing.T) {
        // test code
    })
}
```

**Go Patterns:**
- Use `t.Run()` for sub-tests (table-driven tests)
- `setupTestDB(t *testing.T)` helper for test database
- `database.CreateTestTenant(t, db)` helper for test tenant
- Use table-driven tests for multiple scenarios

**TypeScript Suite Organization:**
```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('QueryBuilder', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    mockOnChange.mockClear();
  });

  it('renders without crashing', () => {
    render(<QueryBuilder entity="findings" query={{}} onChange={mockOnChange} />);
    expect(screen.getByText('Add Filter')).toBeTruthy();
  });

  it('adds a new filter when clicked', () => {
    // test code
  });
});
```

**TypeScript Patterns:**
- Use `describe()` for test suites, `it()` for individual tests
- `beforeEach()` for setup, `afterEach()` for cleanup
- Use React Testing Library for component testing
- Mock fetch with `vi.fn()` for API tests

## Mocking

**Framework - Go:**
- Standard Go testing (no built-in mocking)
- Use interfaces and test doubles for mocking
- Test database setup via `database.SetupTestDB(t)`

**Framework - TypeScript:**
- Vitest built-in mocking (`vi`)
- Module mocking via `vi.mock()`
- Mock fetch with `globalThis.fetch = vi.fn()`

**Patterns - Go:**
```go
// Use test database for repository tests
db := database.SetupTestDB(t)
repo := NewAssetRepository(db)

// Use interfaces for service tests
type mockAPI struct{}
func (m *mockAPI) fetchData() error { return nil }
```

**Patterns - TypeScript:**
```typescript
// Mock fetch globally
const mockFetch = vi.fn();
beforeEach(() => {
  globalThis.fetch = mockFetch;
  mockFetch.mockClear();
});

// Mock module
vi.mock('./external-service', () => ({
  fetchData: vi.fn()
}));
```

**What to Mock:**
- Go: External APIs, network calls, file system
- TypeScript: Fetch API, external services, React Query client

**What NOT to Mock:**
- Go: Pure functions, business logic, database queries (use test DB)
- TypeScript: Pure utilities, internal logic

## Fixtures and Factories

**Test Data - Go:**
- Inline test data in table-driven tests
- Helper functions for creating test entities
- Example: `database.CreateTestTenant(t, db)`

**Test Data - TypeScript:**
- Inline mock data in test files
- No shared fixture directory found
- Example: `const mockQuery = { filters: [] };`

**Location:**
- Go: Co-located with tests (no fixtures/ directory)
- TypeScript: Co-located with tests (no fixtures/ directory)

## Coverage

**Requirements:**
- Go: Minimum 80% coverage (from `AGENTS.md`)
- TypeScript: No explicit coverage requirement

**Configuration:**
- Go: `go test ./... -cover` for coverage
- TypeScript: Vitest supports coverage (c8), but no explicit config

**View Coverage:**
```bash
# Go coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# TypeScript coverage (if configured)
cd ui && npm run test:coverage
```

## Test Types

**Unit Tests - Go:**
- Scope: Test single function/class in isolation
- Mocking: Use test database for repositories, interfaces for services
- Speed: Fast (no network calls)
- Examples: `api/internal/auth/apikey_test.go`, `api/internal/repository/asset_test.go`

**Integration Tests - Go:**
- Scope: Test multiple modules together with real database
- Mocking: Mock only external services, use test DB
- Setup: Use `database.SetupTestDB(t)` for test database
- Examples: `api/internal/handlers/ingest_integration_test.go`, `api/internal/services/query/executor_integration_test.go`

**Unit Tests - TypeScript:**
- Scope: Test single function/hook/component
- Mocking: Mock fetch API, external dependencies
- Examples: `ui/src/api/client.test.ts`, `ui/src/utils/export.test.ts`

**Component Tests - TypeScript:**
- Scope: Test React components with user interactions
- Framework: React Testing Library
- Examples: `ui/src/components/QueryBuilder.test.tsx`, `ui/src/components/DataTable.test.tsx`

**Hook Tests - TypeScript:**
- Scope: Test custom React hooks
- Framework: @testing-library/react
- Examples: `ui/src/hooks/useQuery.test.tsx`, `ui/src/hooks/useDashboardQueries.test.tsx`

**End-to-End Tests:**
- Not currently implemented (planned for future)
- Full stack testing with real API and database

## Common Patterns

**Async Testing - TypeScript:**
```typescript
it('should handle async operation', async () => {
  const { result } = renderHook(() => useQuery('findings', {}));
  await waitFor(() => expect(result.current.isLoading).toBe(false));
  expect(result.current.data).toEqual(expectedData);
});
```

**Error Testing - Go:**
```go
t.Run("returns error on invalid input", func(t *testing.T) {
    _, err := repo.Create(context.Background(), &Asset{})
    if err == nil {
        t.Fatal("expected error, got nil")
    }
})
```

**Error Testing - TypeScript:**
```typescript
it('should throw on invalid input', async () => {
  await expect(apiClient.queryExecute('invalid', {})).rejects.toThrow();
});
```

**Table-Driven Tests - Go:**
```go
func TestNormalizeHostname(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"WebServer01", "webserver01"},
        {"  spaces  ", "spaces"},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := NormalizeHostname(tt.input)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

**Mock Fetch - TypeScript:**
```typescript
const mockFetch = vi.fn();
mockFetch.mockResolvedValue({
    ok: true,
    json: async () => ({ data: [], meta: {} })
});

beforeEach(() => {
    globalThis.fetch = mockFetch;
});
```

---

*Testing analysis: 2026-01-15*
*Update when test patterns change*
