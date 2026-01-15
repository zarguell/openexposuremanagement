# OQL Frontend Integration

## What This Is

Open Exposure Management platform with OQL (Open Query Language) backend implemented and production-ready. Next phase focuses on increasing test coverage to 80%+ and integrating Monaco editor for intuitive frontend query building.

## Core Value

Users can write complex vulnerability queries in SQL-like syntax with 70% less boilerplate, validated in real-time, and executed with sub-millisecond performance.

## Requirements

### Validated

- ✓ OQL backend implementation — MVP (2025-01-15)
- ✓ Parser with 16 AST node types — `api/internal/services/query/oql/ast.go`
- ✓ Tokenizer with position tracking — `api/internal/services/query/oql/tokenizer.go`
- ✓ Recursive descent parser — `api/internal/services/query/oql/parser.go`
- ✓ AST to JSON translator — `api/internal/services/query/oql/translator.go`
- ✓ POST /api/v1/query/oql — Execute OQL queries
- ✓ POST /api/v1/query/oql/validate — Validate syntax
- ✓ POST /api/v1/query/oql/explain — Show translation
- ✓ 140+ tests passing (74.6% coverage) — Unit + integration
- ✓ Performance benchmarks (< 4 μs) — Sub-millisecond parsing
- ✓ Comprehensive documentation (843+ lines) — `docs/oql.md`
- ✓ API Swagger annotations — Interactive docs

### Active

- [ ] Increase OQL test coverage to 80%+
  - Add error handling test paths
  - Add integration tests with real HTTP
  - Add performance benchmarks for all operations
- [ ] Monaco editor integration
  - OQL syntax highlighting
  - Autocomplete for fields and operators
  - Real-time validation with error display
  - Query templates (pre-built examples)
- [ ] Frontend query execution with OQL
  - Replace/supplement JSON query builder with Monaco
  - Execute OQL via /api/v1/query/oql
  - Display query results in table

### Out of Scope

- JSON query deprecation — Keep backward compatible for now
- Advanced OQL features — GROUP BY, aggregations, SELECT clause (MVP only)
- N-way joins — 3+ entity joins (MVP only)
- Database schema changes — No new tables/columns for OQL
- Query result caching — Use React Query caching (backend already cached)

## Context

**Technical Environment:**
- Go 1.24.0 backend with sqlx and PostgreSQL 16
- React 18 + TypeScript 5.2 frontend with Vite
- React Query for server state management
- PostgreSQL-only search (no OpenSearch for MVP)

**Existing OQL Work:**
- Implemented by previous agent with production-ready status
- See `docs/oql-implementation-summary.md` for complete details
- See `docs/oql-test-report.md` for test coverage analysis
- See `docs/oql-performance-report.md` for performance data
- API endpoints already registered and Swagger-documented

**Known Issues to Address:**
- Test coverage at 74.6% needs to reach 80%+ target
- Some integration tests pending from Task 12
- Software repository test has compilation error (unrelated to OQL)
- No frontend integration for OQL yet (only JSON query builder exists)

**User Research Themes:**
- OQL significantly reduces query boilerplate (~70%)
- SQL-like syntax is intuitive for security practitioners
- Real-time validation prevents invalid queries from reaching database
- Dot-walking enables powerful cross-entity queries without complex JOINs

## Constraints

- **Technology**: Go 1.24.0, React 18, TypeScript 5.2, PostgreSQL 16 — Existing stack, maintain compatibility
- **Testing**: 80% minimum coverage for OQL package — Project TDD requirement
- **Performance**: < 5 ms query execution time — Current benchmarks show < 4 μs parsing
- **API Compatibility**: Must not break existing JSON query endpoints — Backward compatible
- **Documentation**: All new features documented with Swagger — Consistent with existing patterns

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Test backfill before frontend | Ensure backend quality foundation before UI integration | — Pending |
| Monaco editor over custom editor | Proven, feature-rich, industry standard | — Pending |
| Syntax highlighting first, then autocomplete | Incremental frontend delivery, faster validation | — Pending |
| Keep JSON queries backward compatible | Avoid breaking existing users, allow gradual migration | — Pending |

---
*Last updated: 2026-01-15 after initialization*
