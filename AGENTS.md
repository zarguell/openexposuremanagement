## Project Structure
- Specs: @docs/architecture.md, @docs/tasks.md
- Source: api/ (Go), ui/ (Vite React), db/ (migrations/sql), docs/ (runbooks/specs)
- Compose: docker-compose.yml (single-machine demo), .env.example (no secrets)
- Entry points: api/cmd/server, ui/src/main.tsx, db/migrations

## Key Context
- Architecture: @docs/architecture.md (MVP spec, endpoints, data model)
- Tasks: @docs/tasks.md (milestones + validation commands)
- Patterns: @docs/.context.md (API shapes, conventions)
- History/Lessons: @docs/.memory.md (decisions, mistakes log)

## Commands (scoped)
make help
docker compose up --build
make test
cd api && go test ./...
cd ui && npm run build
make migrate-up && make migrate-down

## Workflow
1. Read @docs/architecture.md, then pick the next unchecked item in @docs/tasks.md
2. Write/adjust tests first (Go: *_test.go; UI: component tests as needed)
3. Implement smallest change to satisfy acceptance criteria
4. Run the task’s validation command(s) and fix until green
5. Check off the task and record notable decisions in @docs/.memory.md

## Testing Requirements
- Go: `go test ./...` must pass; add unit/integration tests for handlers and DB queries
- Coverage target: ≥ 80% for new Go packages touched
- No direct DB access in handlers; go through a repository/sqlc layer

## Do / Don’t
- Do: env-driven config, tenant scoping everywhere, deterministic asset matching with “why matched”
- Don’t: hard-code ports/URLs, store SPA tokens in localStorage, bypass authN/authZ in demo mode without an explicit dev flag
