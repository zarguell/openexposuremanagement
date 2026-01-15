# Roadmap: OQL Frontend Integration

## Project Overview

**Goal**: Integrate Monaco editor for OQL queries and increase test coverage to 80%+

**Context**: OQL backend is production-ready with 74.6% test coverage. Frontend integration and additional testing needed.

**Current Position**: Phase 1 - Error Handling Test Coverage (Ready to plan)

---

## Domain Expertise

**None** - No domain-specific expertise files found. Project relies on:
- Internal OQL documentation (`docs/oql.md`)
- Implementation summary (`docs/oql-implementation-summary.md`)
- Test report (`docs/oql-test-report.md`)
- Existing codebase patterns

---

## Phase Breakdown

| Phase | Name | Goal | Research Needed | Dependencies | Status |
|-------|------|------|-----------------|--------------|--------|
| 1 | Error Handling Test Coverage | Add comprehensive error handling tests | No | None | 📋 Ready |
| 2 | Integration Test Coverage | Add end-to-end integration tests | No | Phase 1 | ⏳ Pending |
| 3 | Performance Benchmarking | Add comprehensive benchmarks | No | Phase 2 | ⏳ Pending |
| 4 | Monaco Editor Setup | Basic editor with syntax highlighting | Yes | Phase 3 | ⏳ Pending |
| 5 | Monaco Editor Autocomplete | Implement autocomplete for OQL | Yes | Phase 4 | ⏳ Pending |
| 6 | Monaco Editor Validation | Real-time syntax validation | No | Phase 5 | ⏳ Pending |
| 7 | Monaco Editor Templates | Query template library | No | Phase 6 | ⏳ Pending |
| 8 | Frontend OQL Integration | Connect editor to backend API | No | Phase 7 | ⏳ Pending |

---

## Phase Details

### Phase 1: Error Handling Test Coverage
**Goal**: Reach 80% test coverage by adding error handling tests

**Research Required**: No - internal codebase

**Deliverables**:
- Parser error path tests
- Validation error case tests
- API error response tests
- Coverage report showing 80%+

**Acceptance Criteria**:
- All error paths in OQL package tested
- `go test ./... -cover` shows ≥80% coverage
- No test failures

---

### Phase 2: Integration Test Coverage
**Goal**: End-to-end integration tests for OQL functionality

**Research Required**: No - existing patterns

**Dependencies**: Phase 1 complete

**Deliverables**:
- HTTP API integration tests
- Database integration tests with PostgreSQL
- End-to-end query execution tests

**Acceptance Criteria**:
- All API endpoints tested with real HTTP
- Database operations tested with test database
- Integration test suite passes

---

### Phase 3: Performance Benchmarking
**Goal**: Comprehensive benchmarks for all operations

**Research Required**: No - existing benchmark patterns

**Dependencies**: Phase 2 complete

**Deliverables**:
- Benchmarks for parsing operations
- Benchmarks for query execution
- Benchmarks for validation

**Acceptance Criteria**:
- All major operations benchmarked
- Performance regression protection
- Benchmark results documented

---

### Phase 4: Monaco Editor Setup
**Goal**: Basic Monaco editor with OQL syntax highlighting

**Research Required**: Yes - Monaco API, OQL language definition

**Dependencies**: Phase 3 complete

**Deliverables**:
- Monaco editor installed and configured
- OQL language definition (syntax highlighting rules)
- Basic query editor component

**Acceptance Criteria**:
- Monaco editor renders in UI
- OQL keywords properly highlighted
- Editor is interactive and functional

---

### Phase 5: Monaco Editor Autocomplete
**Goal**: Implement autocomplete for OQL fields and operators

**Research Required**: Yes - Monaco completion providers, OQL syntax

**Dependencies**: Phase 4 complete

**Deliverables**:
- Field/property autocomplete providers
- Operator/function autocomplete providers
- Context-aware suggestions

**Acceptance Criteria**:
- Autocomplete triggers on trigger characters
- Suggestions are contextually relevant
- Common OQL patterns are suggested

---

### Phase 6: Monaco Editor Validation
**Goal**: Real-time syntax validation with error display

**Research Required**: No - backend API exists

**Dependencies**: Phase 5 complete

**Deliverables**:
- Real-time syntax validation
- Error highlighting in editor
- Connection to /api/v1/query/oql/validate

**Acceptance Criteria**:
- Validation runs as user types
- Syntax errors highlighted inline
- Validation API integrated

---

### Phase 7: Monaco Editor Templates
**Goal**: Query template library for common patterns

**Research Required**: No - internal OQL examples

**Dependencies**: Phase 6 complete

**Deliverables**:
- Query template library
- Template insertion UI
- Template documentation

**Acceptance Criteria**:
- 5-10 common query templates
- Templates are easily insertable
- Templates documented in UI

---

### Phase 8: Frontend OQL Integration
**Goal**: Connect Monaco editor to backend for query execution

**Research Required**: No - existing frontend patterns

**Dependencies**: Phase 7 complete

**Deliverables**:
- Query execution flow
- Results display in data table
- Error handling for execution

**Acceptance Criteria**:
- OQL queries execute via Monaco editor
- Results displayed in table
- Execution errors properly handled

---

## Progress Tracking

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Test Coverage | 80%+ | 74.6% | 🟡 In Progress |
| Phases Complete | 8 | 0 | ⏳ Not Started |
| Active Phase | Phase 1 | Phase 1 | 📋 Ready to Plan |

---

## Dependencies

**External**:
- None - all dependencies are internal

**Internal**:
- Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5 → Phase 6 → Phase 7 → Phase 8

---

## Notes

- Test backfill prioritized before frontend work (Decision: Test backfill first)
- Monaco editor chosen over custom implementation (Decision: Monaco editor)
- Incremental delivery: Syntax highlighting → Autocomplete → Validation → Templates
- JSON query endpoints remain backward compatible

---

*Last updated: 2026-01-15*
