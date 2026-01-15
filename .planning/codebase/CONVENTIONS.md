# Coding Conventions

**Analysis Date:** 2026-01-15

## Naming Patterns

**Files:**
- Go: `snake_case.go` for all source files (e.g., `asset.go`, `matching.go`, `apikey.go`)
- Go tests: `{source}_test.go` co-located with source
- TypeScript: `PascalCase.ts` or `.tsx` for components/types (e.g., `QueryBuilder.tsx`, `useQuery.ts`)
- TS tests: `{source}.test.{ts,tsx}` co-located with source
- Config: kebab-case (e.g., `vite.config.ts`, `tsconfig.json`)

**Functions:**
- Go exported: PascalCase (e.g., `NewAssetRepository`, `GetByID`, `ListAssets`)
- Go unexported: camelCase (e.g., `getFindingCounts`, `hashAPIKey`)
- Go receivers: Short lowercase initials (1-2 chars) (e.g., `func (r *AssetRepository)`)
- TypeScript functions: camelCase (e.g., `getAssets`, `useQuery`)
- React components: PascalCase (e.g., `QueryBuilder`, `DataTable`)
- Custom hooks: `use{Feature}` prefix (e.g., `useQuery`, `useDashboardQueries`)

**Variables:**
- Go variables: camelCase (e.g., `db`, `assetID`, `tenantID`)
- Go constants: PascalCase for exported (e.g., `NodeTypeOperator`, `ErrInvalidInput`)
- Go errors: PascalCase starting with `Err` (e.g., `ErrUnauthorized`, `ErrInvalidToken`)
- TypeScript variables: camelCase
- TypeScript constants: UPPER_SNAKE_CASE (e.g., `SEVERITY_VALUES`, `ALLOWED_OPERATORS`)

**Types:**
- Go structs/interfaces: PascalCase (e.g., `Asset`, `FindingInstance`, `AssetRepository`)
- Go no I prefix for interfaces
- TypeScript interfaces: PascalCase (e.g., `Query`, `Filter`, `Column`)
- TypeScript type aliases: PascalCase (e.g., `EntityType`, `QueryOptions`)

## Code Style

**Formatting - Go:**
- Tool: Standard `gofmt` (no explicit config found)
- Indentation: Tabs (Go standard)
- Quotes: Double quotes for strings, backticks for multi-line/raw strings
- Semicolons: No explicit semicolons (Go auto-inserts)
- Struct tags: Backticks for JSON/DB tags (e.g., `` `db:"id" json:"id"` ``)

**Formatting - TypeScript:**
- Tool: ESLint (from `ui/package.json`), no `.prettierrc` found
- Indentation: 2 spaces
- Quotes: Single quotes in code, double quotes in JSON, double quotes for JSX attributes
- Semicolons: Explicit semicolons required
- Line length: Not explicitly configured (likely 100 or 120)

**Linting:**
- Go: golangci-lint (from `Makefile` - `make lint`)
- TypeScript: ESLint with TypeScript plugin (from `ui/package.json`)
- Run commands: `make lint` (Go), `cd ui && npm run lint` (TypeScript)

## Import Organization

**Go:**
- Order: Standard library → Third-party → Internal modules
- Grouping: Blank line between groups
- Sorting: Go fmt auto-sorts

**TypeScript:**
- Order: External packages (react, etc.) → Internal modules (@/ aliases) → Relative imports (./) → Type imports
- Grouping: Blank line between groups
- Path aliases: @/ not configured (uses relative imports)
- Sorting: Not strictly enforced

## Error Handling

**Patterns - Go:**
- Strategy: Return errors, handle at boundaries (handlers)
- Custom errors: Extend `errors.New()` with descriptive messages
- Error variables: PascalCase with `Err` prefix (e.g., `ErrUnauthorized`, `ErrInvalidToken`)
- Logging: Use zerolog structured logging before returning errors
- Pattern: `if err != nil { log.Error().Err(err).Msg("description"); return err }`

**Error Handling - TypeScript:**
- Strategy: Throw errors, catch with try/catch in async functions
- API errors: Log error in client, throw for component handling
- React Query: Error boundaries for component-level error handling
- Pattern: `try { await apiCall() } catch (err) { console.error(err); throw err; }`

**When to Throw:**
- Go: Invalid input, missing dependencies, database errors, auth failures
- TypeScript: Invalid input, API errors, unexpected states

**Logging:**
- Go: Use zerolog with context: `log.Error().Err(err).Str("user_id", userID).Msg("Failed to X")`
- TypeScript: Console logging for debugging, toast notifications for user-facing errors

## Logging

**Framework:**
- Go: zerolog v1.34.0 (structured JSON logging)
- TypeScript: console.log/console.error (dev), toast notifications (user-facing)

**Levels:**
- Go: debug, info, warn, error (standard zerolog levels)
- TypeScript: console.log (info), console.error (errors), toast (user-facing)

**Patterns:**
- Go: Structured logging with context fields: `log.Info().Str("action", "ingest").Int("count", n).Msg("Ingestion complete")`
- TypeScript: Error logging in API client (`ui/src/api/client.ts`), toast notifications via ToastContext
- Log at service boundaries (handlers), not in utilities
- Use debug level for verbose info

## Comments

**When to Comment:**
- Explain why, not what (e.g., "Retry 3 times because API has transient failures")
- Document business rules and algorithms
- Explain non-obvious logic
- Avoid obvious comments like "// increment counter"

**Go:**
- Package comments before package declaration
- Exported function comments: `// FunctionName does X...`
- Swagger annotations: `@Summary`, `@Description`, `@Router` tags
- TODO comments: `// TODO: description` (without username, using git blame)

**TypeScript:**
- JSDoc-style for component props: `interface Props { field: string; }`
- Inline comments for complex logic
- JSX comments: `{/* comment */}`

**JSDoc/TSDoc:**
- Go: Standard Go doc comments
- TypeScript: Optional for self-explanatory code, required for public APIs
- Use `@param`, `@returns` tags for complex functions

**TODO Comments:**
- Format: `// TODO: description` (Go), `// TODO: description` (TypeScript)
- No username prefix (using git blame instead)
- Link to issue if exists

## Function Design

**Size:**
- Go: Keep under 50 lines when possible, extract helpers
- TypeScript: Similar guidelines, extract components for complex UI

**Parameters:**
- Go: Max 3 parameters, use options struct for more
- Go: Destructure in parameter list: `func process(ctx context.Context, opts ProcessOptions) error`
- TypeScript: Interface/props for component parameters, object for function options

**Return Values:**
- Go: Explicit returns, error as last value: `func Get(...) (*Asset, error)`
- TypeScript: Explicit returns, return early for guard clauses
- Go: Return early for guard clauses: `if err != nil { return err }`

## Module Design

**Exports - Go:**
- Named exports for all (PascalCase = exported)
- No default exports in Go
- Internal packages (`api/internal/`) cannot be imported externally

**Exports - TypeScript:**
- Named exports preferred: `export const useQuery = ...`
- Default exports for React components: `export default QueryBuilder;`
- API client: Named exports from object: `export const apiClient = { ... }`

**Barrel Files:**
- Go: Not used (internal packages discouraged by Go convention)
- TypeScript: No barrel files (index.ts) found in src/

---

*Convention analysis: 2026-01-15*
*Update when patterns change*
