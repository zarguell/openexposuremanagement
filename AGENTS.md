# Open Exposure Management - Agent Development Workflow

## Project Philosophy: Test-Driven Development (TDD)

**We practice strict TDD: write tests BEFORE implementing features.**

### Core TDD Rules

1. **Red-Green-Refactor Cycle**
   - 🔴 **RED**: Write a failing test first
   - 🟢 **GREEN**: Write minimal code to make the test pass
   - ♻️ **REFACTOR**: Clean up the code while keeping tests green

2. **No Code Without Tests**
   - Never write production code without a failing test first
   - Tests define the requirements and acceptance criteria
   - If you can't write a test for it, don't build it yet

3. **Test Coverage Requirements**
   - **Go**: Minimum 80% coverage for new code
   - Unit tests for all public functions/methods
   - Integration tests for critical paths (ingestion, matching, auth)
   - Table-driven tests for multiple scenarios

4. **Testing Priorities**
   - Test behavior, not implementation
   - Test edge cases and error conditions
   - Test with realistic fixtures and sample data
   - Tests should be fast, independent, and repeatable

### Development Workflow

For any task (from docs/tasks.md), follow this sequence:

```bash
# 1. Write the test FIRST
# Create test file or add to existing
# Focus on one behavior/requirement

# 2. Run the test - it should FAIL (red)
go test ./...

# 3. Write minimal implementation to pass
# Don't worry about perfection, just make it pass

# 4. Run test again - should PASS (green)
go test ./...

# 5. Refactor if needed
# Improve code while tests stay green

# 6. Commit
git add . && git commit -m "feat: implement X with tests"
```

### Example: Adding a New Function

**BAD** (writing code first):
```go
// normalize.go
func NormalizeHostname(name string) string {
    return strings.ToLower(strings.TrimSpace(name))
}
```

**GOOD** (TDD approach):
```go
// normalize_test.go - WRITE THIS FIRST
func TestNormalizeHostname(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"WebServer01", "webserver01"},
        {"  spaces  ", "spaces"},
        {"MixedCase.EXAMPLE.com", "mixedcase.example.com"},
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

// NOW run: go test ./... - it FAILS (red)

// normalize.go - WRITE THIS AFTER TEST FAILS
func NormalizeHostname(name string) string {
    return strings.ToLower(strings.TrimSpace(name))
}

// run: go test ./... - it PASSES (green)
```

### Code Review Checklist

Before committing or marking a task complete:

- [ ] All tests were written BEFORE the implementation
- [ ] `go test ./...` passes
- [ ] `go test ./... -cover` shows ≥80% coverage
- [ ] No obvious code duplication
- [ ] Error cases are tested
- [ ] Edge cases are covered

### Project Context

#### Specs & Requirements
- Architecture: @docs/architecture.md (MVP spec, endpoints, data model)
- Tasks: @docs/tasks.md (milestones + validation commands)
- Patterns: @docs/.context.md (API shapes, conventions)

#### Key Constraints
- **Database**: Postgres-only (no OpenSearch for MVP)
- **Search**: Use indexes and materialized views
- **Auth**: OIDC PKCE for SPA, API keys for ingestion
- **Tenancy**: All data tenant-scoped (except intel cache)
- **Asset Matching**: Deterministic with audit trail ("why matched")

#### Architecture Highlights
- **API**: Go 1.21+ with sqlx, zerolog, gorilla/mux
- **Database**: Postgres 16 with golang-migrate
- **UI**: React + Vite + TypeScript (for later milestone)
- **Deployment**: Docker Compose for single-machine demo

#### Critical Paths (Must Have Tests)
1. **Ingestion Pipeline** (VM findings → assets → findings)
   - Payload validation
   - Identifier normalization
   - Asset matching algorithm
   - Upsert operations (assets, definitions, findings)

2. **Authentication & Authorization**
   - API key validation
   - JWT token parsing (OIDC)
   - Role-based access control

3. **Effective Status Computation**
   - Scanner status → effective status
   - Suppression application
   - Recompute on policy change

#### Testing Strategy

**Unit Tests** (fast, isolated)
- Repository methods (with test DB or mock)
- Normalization functions
- Matching logic
- Validation functions

**Integration Tests** (slower, with DB)
- Full ingestion flow
- Asset matching with real data
- Auth middleware with test tokens

**End-to-End Tests** (slowest, full stack)
- POST /ingest/vm/findings → DB verification
- GET /assets, /findings → response validation
- Auth flow → protected endpoint access

### Common Patterns

**Table-Driven Tests** (for multiple scenarios):
```go
func TestNormalizeHostname(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"simple", "WebServer", "webserver"},
        {"with spaces", "  Server  ", "server"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := NormalizeHostname(tt.input)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

**Error Testing**:
```go
func TestValidatePayload(t *testing.T) {
    t.Run("missing source returns error", func(t *testing.T) {
        payload := &VMFindingsPayload{Source: ""}
        err := payload.Validate()
        if err == nil {
            t.Fatal("expected error, got nil")
        }
    })
}
```

### Commands

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/ingest -v

# Run specific test
go test ./internal/ingest -run TestNormalizeHostname

# View coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Important: Backfilling Tests

For existing code (Milestones 1-3), we need to backfill tests:
1. Start with critical paths (auth, repository)
2. Write tests that document current behavior
3. Refactor to improve if tests reveal issues
4. Aim for 80%+ coverage before moving to new features

## Agent Instructions

When working on this project:

1. **Always start with a test** - never write production code first
2. **Follow the TDD cycle** - red → green → refactor
3. **Check coverage** - ensure ≥80% before committing
4. **Use table-driven tests** for multiple scenarios
5. **Test errors and edge cases** - not just happy paths
6. **Keep tests independent** - no shared state between tests
7. **Make tests fast** - use mocks or test DB where appropriate

Remember: **Tests are documentation.** They show how the code SHOULD behave.
