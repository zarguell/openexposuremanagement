# Codebase Concerns

**Analysis Date:** 2026-01-15

## Tech Debt

**Authentication & Authorization - NOT PRODUCTION READY:**
- Issue: JWKS fetching not implemented in `api/internal/auth/jwt.go` (line 68, 199)
- Why: Feature stubbed for MVP demo
- Impact: Cannot verify JWT signatures, accepts any JWT in demo mode
- Fix approach: Implement `fetchJWKS()` from issuer URL, cache JWKS, verify signatures

**API Key Validation - INCOMPLETE:**
- Issue: API key validation not implemented in `api/internal/auth/apikey.go`
- Why: TODO comments for repository lookup and scope checking
- Impact: All API key checks bypassed, accepts hardcoded demo key only
- Fix approach: Implement keyHash database lookup, scope validation, source binding enforcement

**RBAC Middleware - NOT IMPLEMENTED:**
- Issue: Role-based access control does nothing in `api/internal/auth/rbac.go`
- Why: TODO for user context extraction
- Impact: All role checks bypassed, no authorization enforcement
- Fix approach: Implement role checking middleware, extract user context from JWT

**Suppression Logic - MISSING:**
- Issue: Finding suppression not implemented in `api/internal/ingest/finding.go` (line 54)
- Why: TODO comment for future task
- Impact: Cannot suppress findings from effective status computation
- Fix approach: Implement suppression lookup during ingestion and status recompute

**Ingestion Validation - INCOMPLETE:**
- Issue: API key scope and source binding not enforced in `api/internal/handlers/ingest.go`
- Why: TODO comments for scope and source binding checks
- Impact: Ingestion endpoint not properly secured
- Fix approach: Complete API key validation middleware, enforce scopes

## Known Bugs

**Test Build Failure:**
- Symptoms: `api/internal/repository/software_test.go` fails to compile
- Trigger: Running `go test ./...`
- Files: `api/internal/repository/software_test.go` (lines 100-102)
- Workaround: None, blocks test execution
- Root cause: Type mismatch - cannot use string constant as *string pointer
- Fix: Correct type annotations in test

**Test Panic:**
- Symptoms: `api/internal/server/server_test.go` panics with nil pointer dereference
- Trigger: Running `TestQueryEndpointsValidation`
- File: `api/internal/server/server_test.go`
- Workaround: Skip this test
- Root cause: Database not initialized in test setup
- Fix: Add database setup to test

**Missing Integration Tests:**
- Symptoms: 55+ tests skipped with `t.Skip("TODO: Implement with test database")`
- Trigger: Running full test suite
- Files: Multiple test files (`api/internal/repository/*_test.go`, `api/internal/handlers/*_test.go`)
- Workaround: None (blocks test coverage goals)
- Root cause: Test database setup incomplete
- Fix: Complete test database utilities, remove TODOs

## Security Considerations

**Demo Mode Bypasses All Security:**
- Risk: DEMO_MODE environment variable bypasses authentication and authorization
- Files: `api/internal/handlers/middleware.go` (lines 16-29, 47-51), `api/internal/auth/jwt.go` (lines 54-60), `api/internal/auth/apikey.go` (line 29)
- Current mitigation: Demo mode warnings in logs, but no enforcement
- Recommendations: Remove demo mode from production builds, require explicit environment flag, add production guard

**Hardcoded Secrets in Examples:**
- Risk: Default passwords in `.env.example` can be used inadvertently
- Files: `.env.example` (line 2, 15)
- Current mitigation: None (examples checked into repo)
- Recommendations: Replace with placeholder strings like "your-password-here", remove demo key from example

**Incomplete JWT Implementation:**
- Risk: JWT signature verification not functional
- File: `api/internal/auth/jwt.go` (line 70)
- Current mitigation: Demo mode accepts any JWT (security hole)
- Recommendations: Implement JWKS fetching and signature verification before production use

**API Key Security:**
- Risk: API key validation incomplete, no scope enforcement
- File: `api/internal/auth/apikey.go` (line 29, 24)
- Current mitigation: Accepts hardcoded "demo-key-for-development-only"
- Recommendations: Implement database lookup, hash comparison, scope validation, source binding

**Hardcoded Tenant ID:**
- Risk: All users assigned to TenantID 1
- File: `api/internal/auth/jwt.go` (line 173)
- Current mitigation: None
- Recommendations: Extract tenant_id from JWT claims, implement per-tenant data isolation

**Test Database Password:**
- Risk: Hardcoded test database password in source
- File: `api/internal/database/testutil.go` (line 25)
- Current mitigation: Test-only code, but password visible in repo
- Recommendations: Use environment variable for test database password

## Performance Bottlenecks

**SQL String Manipulation in Query Executor:**
- Problem: Tenant filter injected via string manipulation in `api/internal/services/query/executor.go` (lines 84-126)
- Measurement: Not measured, but multiple `strings.Index()` and `strings.ReplaceAll()` calls per query
- Cause: Manual WHERE/AND clause injection for tenant filtering
- Improvement path: Use SQL builder library (e.g., squirrel) or refactor to parameter-based injection

**No Connection Pool Timeout Configuration:**
- Problem: Database connection pool uses Go defaults (no explicit timeout)
- File: `api/internal/database/database.go`
- Measurement: No explicit timeout set
- Cause: Connection pool setup incomplete
- Improvement path: Configure SetConnMaxLifetime, SetConnMaxIdleTime in sqlx

## Fragile Areas

**Demo Mode Logic:**
- Why fragile: Duplicated across 4 files, bypasses all security
- Files: `api/internal/handlers/middleware.go`, `api/internal/auth/jwt.go`, `api/internal/auth/apikey.go`, `api/internal/auth/rbac.go`
- Common failures: Accidental deployment with demo mode enabled, security bypass in production
- Safe modification: Centralize demo mode check in config, add production guard, document security implications
- Test coverage: Auth tests bypassed in demo mode, gaps in coverage

**Query Executor Tenant Injection:**
- Why fragile: String-based SQL manipulation complex and error-prone
- File: `api/internal/services/query/executor.go` (lines 76-126)
- Common failures: SQL syntax errors, incorrect parameter ordering, missed edge cases
- Safe modification: Refactor to use SQL builder library, add comprehensive tests for tenant injection
- Test coverage: Limited test coverage for edge cases

**OQL Parser/Translator Chain:**
- Why fragile: Complex recursive descent parser with 3-stage chain (tokenize → parse → translate)
- Files: `api/internal/services/query/oql/parser.go` (464 lines), `api/internal/services/query/oql/translator.go` (337 lines)
- Common failures: Unhandled syntax errors, incorrect SQL generation for complex queries
- Safe modification: Add comprehensive parser test cases, document supported grammar clearly
- Test coverage: 484-line test file, but may miss edge cases

## Scaling Limits

**PostgreSQL Connection Pool:**
- Current capacity: 25 max open connections (configured in `api/internal/database/database.go`)
- Limit: Database connection exhaustion at high concurrency
- Symptoms at limit: Connection pool errors, request timeouts
- Scaling path: Increase pool size, add connection pool metrics, implement horizontal scaling

**Query Execution Timeout:**
- Current capacity: 5 second timeout for all queries (`api/internal/services/query/executor.go` line 17)
- Limit: Long-running queries terminated
- Symptoms at limit: Query failures for complex reports
- Scaling path: Per-endpoint timeout configuration, query optimization, async query execution

**React Query Cache:**
- Current capacity: In-memory cache per browser tab
- Limit: No shared cache, re-fetch on page reload
- Symptoms at limit: Repeated API calls, poor performance
- Scaling path: Add persistent cache (Redis), implement cache invalidation strategy

## Dependencies at Risk

**lib/pq (PostgreSQL Driver):**
- Risk: v1.10.9 is older, less actively maintained
- Impact: May have compatibility issues with newer PostgreSQL versions
- Migration plan: Consider migrating to pgx (more modern, better performance)

**Missing Dependencies:**
- Risk: No logging aggregation service (Sentry, DataDog)
- Impact: Production errors not centrally tracked
- Migration plan: Add error tracking service for production monitoring

## Missing Critical Features

**Suppression Logic:**
- Problem: Cannot suppress findings from effective status computation
- Current workaround: N/A (feature not available)
- Blocks: Users cannot ignore false positives or temporary issues
- Implementation complexity: Medium (need suppression UI, policies, status recompute)

**JWKS Fetching:**
- Problem: Cannot verify JWT signatures from OIDC provider
- Current workaround: Demo mode bypasses verification
- Blocks: Production deployment with real OIDC
- Implementation complexity: Medium (HTTP client, caching, signature verification)

**API Key Validation:**
- Problem: Ingestion endpoint accepts any API key (demo mode only)
- Current workaround: N/A (security hole)
- Blocks: Secure scanner ingestion
- Implementation complexity: Medium (database lookup, hash comparison, scope validation)

**User Context Extraction:**
- Problem: Cannot determine current user for authorization
- Current workaround: Demo mode hardcodes admin user
- Blocks: RBAC enforcement, user-specific queries
- Implementation complexity: Low (extract from JWT claims)

## Test Coverage Gaps

**Low Coverage Packages (<50%):**
- What's not tested: `api/cmd/server` (0.0%), `api/cmd/test-oql` (0.0%), `api/internal/database` (3.2%), `api/internal/auth` (35.6%), `api/internal/handlers` (37.4%), `api/internal/intel` (35.4%), `api/internal/config` (64.0%)
- Risk: Critical auth, handlers, and database code under-tested
- Priority: High (auth, handlers), Medium (database, intel, config)
- Difficulty to test: Integration test infrastructure incomplete, TODOs blocking test implementation

**Skipped Integration Tests:**
- What's not tested: 55+ tests skipped with `t.Skip("TODO: Implement with test database")`
- Risk: Test suite incomplete, coverage goals unmet
- Priority: High
- Difficulty to test: Medium (need test database setup, but utilities exist)

**Frontend Test Coverage:**
- What's not tested: No explicit coverage requirement, may have gaps
- Risk: Untested UI components may fail in production
- Priority: Medium
- Difficulty to test: Low (Vitest configured, but coverage tracking not enabled)

## Code Complexity

**Large Files (>200 lines):**
- Backend: `api/internal/services/query/oql/translator_test.go` (484 lines), `api/internal/services/query/oql/parser.go` (464 lines), `api/internal/services/query/oql/oql.go` (458 lines), `api/internal/services/query/oql/translator.go` (337 lines), `api/internal/services/query/oql/tokenizer.go` (212 lines)
- Frontend: `ui/src/components/QueryBuilder.tsx` (432 lines), `ui/src/pages/Dashboard.tsx` (356 lines), `ui/src/components/UnifiedQueryBuilder.tsx` (324 lines), `ui/src/pages/Findings.tsx` (302 lines)
- Risk: Hard to understand, difficult to test, likely contains multiple responsibilities
- Fix approach: Extract sub-components and helper functions, split large files by feature

**TypeScript `any` Usage:**
- What's not typed correctly: 38 instances of `any` type in TS/TSX files
- Files: `ui/src/components/QueryWidget.tsx`, `ui/src/components/QueryResultsTable.tsx`, `ui/src/components/UnifiedQueryBuilder.tsx`, `ui/src/pages/Dashboard.tsx`, `ui/src/pages/Assets.tsx`, `ui/src/pages/Findings.tsx`, and others
- Risk: Type safety lost, potential runtime errors, harder refactoring
- Fix approach: Replace `any` with proper types or generics, add type definitions for API responses

## Duplicate Code Patterns

**Repeated Error Logging:**
- Issue: Same error logging pattern repeated 30+ times across handlers
- Pattern: `log.Error().Err(err).Msg("Failed to...")`
- Files: `api/internal/handlers/query.go` (8 instances), `api/internal/handlers/dashboard.go` (4 instances), `api/internal/handlers/ingest.go` (5 instances), and more
- Fix approach: Create helper function for standardized error logging with context

**Repeated Repository Query Patterns:**
- Issue: Similar SELECT patterns duplicated across repository files
- Pattern: Long SELECT field lists repeated 5+ times
- Files: `api/internal/repository/finding.go`, `api/internal/repository/definition.go`, `api/internal/repository/asset.go`
- Fix approach: Use struct field list constant or reflection-based mapper

**Repeated Demo Mode Checks:**
- Issue: Demo mode bypass logic duplicated in 4 files
- Files: `api/internal/handlers/middleware.go`, `api/internal/auth/jwt.go`, `api/internal/auth/apikey.go`, `api/internal/auth/rbac.go`
- Fix approach: Centralize demo mode check in config, use single function

## Documentation Gaps

**Missing Documentation:**
- What's missing: CONTRIBUTING.md, DEVELOPMENT.md at project root
- Impact: New contributors lack onboarding guidance
- Fix approach: Add contributor and development guides based on `AGENTS.md` workflow

**Complex Function Documentation:**
- What's missing: Detailed comments for complex translation and tenant injection logic
- Files: `api/internal/services/query/translator.go` (lines 87-231), `api/internal/services/query/executor.go` (lines 76-126)
- Impact: Hard to understand complex algorithms
- Fix approach: Add detailed inline comments, examples, and architectural documentation

**Environment Variable Documentation:**
- What's missing: No descriptions for 23 environment variables in `.env.example`
- Impact: Developers don't know what each variable does
- Fix approach: Add comments to `.env.example` or create environment variables documentation

## Positive Findings

**Good Practices:**
- SQL injection protection: All database queries use parameterized queries with `$1, $2` placeholders
- Test coverage in critical areas: `api/internal/api` (83.3%), `api/internal/ingest` (71.4%), `api/internal/software` (97.0%), `api/internal/middleware` (100.0%)
- Comprehensive logging: Structured logging with zerolog throughout backend
- Clear architecture: Layered separation (handlers → services → repositories)
- Context propagation: Proper use of `context.Context` throughout Go codebase
- Resource cleanup: Consistent `defer Close()` and `defer cancel()` patterns
- API documentation: Swagger annotations present in handlers
- TypeScript compilation: No TS compilation errors
- TDD workflow: Strict test-driven development enforced in `AGENTS.md`

---

*Concerns audit: 2026-01-15*
*Update as issues are fixed or new ones discovered*
