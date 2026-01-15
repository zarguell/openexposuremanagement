# OQL (Open Query Language) - Design Document

## Overview

OQL is a concise, expression-based query language for the Open Exposure Management platform. It compiles to the existing unified query JSON format, providing a more intuitive alternative to JSON for both analysts building queries in the UI and automation scripts requiring programmatic access.

**Design Goals:**
- Reduce boilerplate compared to JSON queries (70% less verbose)
- Maintain SQL familiarity without full SQL complexity
- Support dot-walking for cross-entity queries
- Clear, actionable error messages
- Fast parsing (< 1ms for complex queries)

## Table of Contents

1. [Syntax Overview](#1-syntax-overview)
2. [Operators and Comparisons](#2-operators-and-comparisons)
3. [Logical Operators and Grouping](#3-logical-operators-and-grouping)
4. [Sorting, Limiting, and Offsets](#4-sorting-limiting-and-offsets)
5. [Dot-Walking and Entity Relationships](#5-dot-walking-and-entity-relationships)
6. [Grammar and Parser Implementation](#6-grammar-and-parser-implementation)
7. [API Integration and Endpoints](#7-api-integration-and-endpoints)
8. [Error Handling and Validation](#8-error-handling-and-validation)
9. [Testing Strategy](#9-testing-strategy)
10. [UI Integration and User Experience](#10-ui-integration-and-user-experience)
11. [Documentation and Examples](#11-documentation-and-examples)
12. [Implementation Phases](#12-implementation-phases)

---

## 1. Syntax Overview

OQL is an expression-based language with no SELECT/FROM boilerplate - just filter expressions with optional sorting and pagination.

**Basic Syntax:**
```
field comparisons [logical operators] [sort] [limit] [offset]
```

**Key Principles:**
- **Expression-based**: No SELECT/FROM boilerplate
- **Keywords only**: `and`, `or`, `not` (case-insensitive, no &&, ||, !)
- **Dot-walking**: Use `software.field` or `findings.field` for related entities
- **NOT for negation**: `not software.vendor = 'X'` creates NOT EXISTS subquery
- **String literals**: Single quotes required (e.g., `'CrowdStrike'`)
- **Boolean literals**: `true` and `false` (case-insensitive)
- **Numeric literals**: Plain numbers (e.g., `epss_score > 0.9`)

**Example Query:**
```
is_active = true and not software.vendor = 'CrowdStrike' limit 100 sort canonical_name asc
```

**JSON Equivalent:**
```json
{
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true},
    {"field": "software.vendor", "operator": "eq", "value": "CrowdStrike", "negate": true}
  ],
  "limit": 100,
  "sort": [{"field": "canonical_name", "order": "asc"}]
}
```

---

## 2. Operators and Comparisons

OQL supports a mix of symbolic operators for comparisons and keyword operators for patterns and sets.

### Comparison Operators
- `=` - Equal
- `!=` - Not equal
- `<` - Less than
- `>` - Greater than
- `<=` - Less than or equal
- `>=` - Greater than or equal

### Pattern Matching
- `like` - Pattern matching with `%` wildcards
  - `hostname like 'web%'` - Starts with "web"
  - `hostname like '%.local'` - Ends with ".local"
  - `hostname like '%prod%'` - Contains "prod"

### Set Membership
- `in` - Match multiple values
  - `severity in ('critical', 'high')`
  - `findings.cve in ('CVE-2021-44228', 'CVE-2021-45046')`

### Null Handling
- `is null` - Field is null
- `is not null` - Field is not null
  - `ipv4 is null` - Assets without IP address

### Examples
```
# Complex pattern matching
hostname like 'web%' and software.vendor = 'Microsoft'

# Multiple severity levels
findings.severity in ('critical', 'high') and findings.effective_status = 'open'

# Numeric comparison with EPSS
findings.epss_score > 0.9 and findings.is_kev = true

# Null checks
ipv4 is not null and is_active = true
```

---

## 3. Logical Operators and Grouping

OQL supports standard logical operators with parentheses for grouping complex expressions.

### Logical Operators
- `and` - Both conditions must be true
- `or` - At least one condition must be true
- `not` - Negates the condition (creates anti-join for dot-walking)

### Grouping
- Use parentheses `()` to group expressions
- Nested parentheses are supported
- Operator precedence: `not` > `and` > `or`

### Examples
```
# Simple AND
is_active = true and software.vendor = 'CrowdStrike'

# Simple OR
software.vendor = 'Microsoft' or software.vendor = 'Apple'

# NOT for negation (anti-join)
is_active = true and not software.vendor = 'CrowdStrike'

# Grouping for complex logic
(is_active = true and software.vendor = 'Microsoft') or (software.vendor = 'Apple' and software.product_name = 'macOS')

# Combining NOT with OR
is_active = true and not (software.vendor = 'Microsoft' or software.vendor = 'Apple')

# Multiple negations
is_active = true and not software.vendor = 'CrowdStrike' and not findings.severity = 'critical'

# Nested grouping
(environment = 'prod' and (software.vendor = 'CrowdStrike' or software.vendor = 'SentinelOne')) or (environment = 'dev' and software.vendor = 'ClamAV')
```

### Precedence Examples
```
# NOT binds first, then AND, then OR
not a = 'x' and b = 'y' or c = 'z'
# Evaluated as: ((not a = 'x') and b = 'y') or c = 'z'

# Use parentheses to override
not (a = 'x' and b = 'y' or c = 'z')
# Evaluated as: not ((a = 'x' and b = 'y') or c = 'z')
```

---

## 4. Sorting, Limiting, and Offsets

OQL uses end-of-query clauses for result pagination and sorting.

### Syntax
```
... limit <count> offset <count> sort <field> <direction>
```

### Clauses
- `limit <n>` - Maximum number of results (default: 100, max: 5000)
- `offset <n>` - Skip n results (for pagination, default: 0)
- `sort <field> <direction>` - Sort by field (direction: `asc` or `desc`)

### Order
Clauses must appear in this order at the end of the query:
1. `limit` (optional, defaults to 100)
2. `offset` (optional, defaults to 0)
3. `sort` (optional, multiple sorts allowed)

### Examples
```
# Basic limit
is_active = true limit 50

# Limit with offset (pagination)
is_active = true limit 50 offset 100

# Single sort ascending
is_active = true sort canonical_name asc

# Single sort descending
findings.severity = 'critical' sort findings.epss_score desc

# Multiple sorts
findings.severity = 'critical' sort findings.epss_score desc, canonical_name asc

# All clauses together
is_active = true and not software.vendor = 'CrowdStrike' limit 100 offset 0 sort canonical_name asc

# Sort on related entity field
software.vendor = 'Microsoft' sort software.product_name asc, software.version desc

# Reverse chronological
findings.effective_status = 'open' sort findings.last_observed_at desc
```

---

## 5. Dot-Walking and Entity Relationships

OQL uses dot notation to reference fields on related entities without explicit JOIN syntax. The backend automatically generates appropriate SQL (INNER JOIN or NOT EXISTS) based on context.

### Supported Entity Prefixes
- `software.*` - Fields from software_inventory table
- `findings.*` - Fields from findings table

### Dot-Walking Rules

1. **Positive match** → INNER JOIN (must have matching record)
   ```
   software.vendor = 'Microsoft'
   # Generates: INNER JOIN software_inventory ... WHERE software_inventory.vendor = 'Microsoft'
   ```

2. **Negated match** → NOT EXISTS subquery (anti-join)
   ```
   not software.vendor = 'CrowdStrike'
   # Generates: WHERE NOT EXISTS (SELECT 1 FROM software_inventory ... WHERE vendor = 'CrowdStrike')
   ```

3. **Multiple entity references** → Multiple INNER JOINs
   ```
   software.product_name = 'Log4j' and findings.cve = 'CVE-2021-44228'
   # Generates: INNER JOIN software_inventory ... INNER JOIN findings ...
   ```

### Examples
```
# Assets with specific software
software.vendor = 'Microsoft' and software.product_name = 'SQL Server'

# Assets without software (anti-join)
is_active = true and not software.vendor = 'CrowdStrike'

# Assets with software and CVEs
software.product_name = 'Apache Tomcat' and findings.severity = 'critical'

# Assets missing software with exploitable CVEs
not software.vendor = 'CrowdStrike' and findings.is_kev = true

# Complex correlation
software.vendor = 'Oracle' and (findings.severity = 'critical' or findings.epss_score > 0.9)
```

### Available Fields

**Software fields (`software.*`):**
- `vendor` - Software vendor
- `product_name` - Product name
- `version` - Version number
- `cpe_string` - CPE identifier
- `install_path` - Installation path
- `first_seen_at` - First discovery timestamp
- `last_seen_at` - Last seen timestamp

**Findings fields (`findings.*`):**
- `severity` - Severity level (critical, high, medium, low)
- `scanner_status` - Status from scanner
- `effective_status` - Computed status (open, fixed)
- `cve` - CVE identifier
- `epss_score` - EPSS score (0-1)
- `is_kev` - CISA Known Exploited Vulnerability flag
- `first_observed_at` - First observation timestamp
- `last_observed_at` - Last observation timestamp

---

## 6. Grammar and Parser Implementation

The OQL parser translates query strings into the existing JSON query format, following a formal grammar for consistent parsing.

### Formal Grammar (EBNF-like)
```
query       ::= expression [ sort_clause ] [ limit_clause ] [ offset_clause ]
expression  ::= or_expression
or_expression   ::= and_expression { 'or' and_expression }*
and_expression  ::= not_expression { 'and' not_expression }*
not_expression  ::= [ 'not' ] primary_expression
primary_expression
            ::= comparison_expression
             | '(' expression ')'

comparison_expression
            ::= field operator value
             | field 'in' '(' value_list ')'
             | field 'is' [ 'not' ] 'null'
             | field 'like' string

operator    ::= '=' | '!=' | '<' | '>' | '<=' | '>='
value       ::= string | number | boolean
value_list  ::= value { ',' value }*
field       ::= identifier [ '.' identifier ]
string      ::= '\'' [^\']* '\''
number      ::= [0-9]+ [ '.' [0-9]+ ]
boolean     ::= 'true' | 'false'
identifier  ::= [a-zA-Z_][a-zA-Z0-9_]*

sort_clause ::= 'sort' sort_term { ',' sort_term }*
sort_term   ::= field ( 'asc' | 'desc' )
limit_clause::= 'limit' number
offset_clause::= 'offset' number
```

### Parser Implementation Strategy

**1. Tokenizer (Lexer):**
- Split input into tokens: identifiers, operators, literals, keywords
- Handle string literals with escape sequences
- Preserve whitespace information (for error messages)

**2. Recursive Descent Parser:**
- Parse expressions following operator precedence
- Build Abstract Syntax Tree (AST) for the query
- Validate structure during parsing

**3. AST to JSON Translator:**
- Convert AST nodes to Filter objects
- Detect dot-walking patterns for entity relationships
- Handle NOT for negation (set `negate: true`)

### Example Parse Flow
```
Input:  "is_active = true and not software.vendor = 'CrowdStrike' limit 100"

Tokens: IDENTIFIER(is_active), OPERATOR(=), BOOLEAN(true),
        KEYWORD(and), KEYWORD(not), IDENTIFIER(software), DOT(.),
        IDENTIFIER(vendor), OPERATOR(=), STRING('CrowdStrike'),
        KEYWORD(limit), NUMBER(100)

AST:
├─ Filter: field="is_active", operator="eq", value=true
└─ AND
   └─ Filter: field="software.vendor", operator="eq", value="CrowdStrike", negate=true

JSON Output:
{
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true},
    {"field": "software.vendor", "operator": "eq", "value": "CrowdStrike", "negate": true}
  ],
  "limit": 100
}
```

### Go Package Structure
```
internal/services/query/
├── oql/
│   ├── tokenizer.go      # Lexer for tokenizing OQL strings
│   ├── parser.go         # Recursive descent parser
│   ├── ast.go            # AST node definitions
│   ├── translator.go     # AST → JSON query conversion
│   └── oql_test.go       # Comprehensive tests
├── types.go              # Existing query types
├── translator.go         # Existing SQL translator
└── executor.go           # Existing query executor
```

---

## 7. API Integration and Endpoints

OQL integrates alongside the existing JSON query API, offering multiple ways to execute queries.

### New Endpoints

**1. POST /api/v1/query/oql**
- Accepts OQL query string in request body
- Returns results in same format as `/api/v1/query/unified`
- Request body: `{"query": "is_active = true limit 10"}`

**2. POST /api/v1/query/oql/validate**
- Validates OQL syntax without executing
- Returns parse errors or syntax confirmation
- Request body: `{"query": "is_active = true and ..."}`
- Response: `{"valid": true, "errors": []}`

**3. POST /api/v1/query/oql/explain**
- Converts OQL to JSON format (for debugging/learning)
- Shows how OQL translates to unified query JSON
- Request body: `{"query": "..."}`
- Response: `{"unified_query": {...}, "sql": "..."}`
- Useful for UI to show "what will this query do?"

### Backend Flow
```
OQL Query String
    ↓
Tokenizer (Lexer)
    ↓
Parser (AST)
    ↓
JSON Translator
    ↓
Query Validator (existing)
    ↓
SQL Translator (existing)
    ↓
Executor (existing)
    ↓
Database
```

### Handler Integration
```go
// internal/handlers/query.go

func (h *QueryHandler) QueryOQL(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Query string `json:"query"`
    }
    json.Unmarshal(body, &req)

    // Parse OQL to JSON
    jsonQuery, err := oql.Parse(req.Query)
    if err != nil {
        // Return 400 with syntax error
    }

    // Execute using existing executor
    result, err := h.executor.Execute(..., jsonQuery)
    // Return results
}
```

---

## 8. Error Handling and Validation

OQL provides clear, actionable error messages to help users fix syntax issues quickly.

### Error Types

**1. Syntax Errors:**
- Missing operators: `is_active true` → "Expected operator after 'is_active'"
- Unmatched parentheses: `(is_active = true` → "Unclosed parenthesis, expected ')'"
- Invalid operators: `is_active ~= true` → "Unknown operator '~=', expected one of: =, !=, <, >, <=, >="
- Missing quotes: `hostname = web01` → "Unterminated string literal, expected quotes"

**2. Semantic Errors:**
- Unknown fields: `unknown_field = 'x'` → "Unknown field 'unknown_field'"
- Invalid dot-walking: `invalid.field = 'x'` → "Unknown entity prefix 'invalid'"
- Type mismatches: `is_active = 'yes'` → "Field 'is_active' expects boolean, got string"
- Invalid values: `severity = 'invalid'` → "Invalid value 'invalid' for field 'severity'"

**3. Validation Errors:**
- Empty queries: `` → "Query cannot be empty"
- Missing limit value: `limit` → "Expected number after 'limit'"
- Invalid sort direction: `sort name invalid` → "Sort direction must be 'asc' or 'desc'"

### Error Response Format
```json
{
  "error": "Syntax Error",
  "message": "Expected operator after 'is_active'",
  "position": {
    "line": 1,
    "column": 12,
    "context": "is_active true"
            ^
  },
  "suggestion": "Did you mean: is_active = true"
}
```

### Validation Pipeline
```
OQL Input
    ↓
Tokenizer → Token stream (syntax errors caught here)
    ↓
Parser → AST (structural errors caught here)
    ↓
Field Validator → Field existence check
    ↓
Type Validator → Value type check
    ↓
Query Validator → Existing validation logic
    ↓
Ready to execute
```

### Examples of Good Error Messages
```
Input: "is_active = true and (software.vendor = 'Microsoft'"
Error: Unclosed parenthesis
       Position: line 1, column 45
       Context: ...software.vendor = 'Microsoft'
                                          ^
       Suggestion: Add closing parenthesis: ...software.vendor = 'Microsoft')

Input: "hostname like web01"
Error: Unterminated string literal
       Position: line 1, column 15
       Context: hostname like web01
                      ^
       Suggestion: Use quotes: hostname like 'web01'

Input: "unknown_field = 'x'"
Error: Unknown field 'unknown_field'
       Suggestion: Did you mean: hostname, ipv4, canonical_name?

Input: "is_active = 'yes'"
Error: Type mismatch
       Field 'is_active' expects boolean, got string
       Suggestion: Use true or false: is_active = true
```

---

## 9. Testing Strategy

Comprehensive testing for OQL to ensure correctness, performance, and reliability.

### Unit Tests

**1. Tokenizer Tests (`oql/tokenizer_test.go`):**
- Valid tokens: identifiers, operators, literals, keywords
- String literals with escapes
- Whitespace handling
- Invalid characters
- Edge cases: empty input, special characters

**2. Parser Tests (`oql/parser_test.go`):**
- Simple expressions: `field = value`
- Logical operators: AND, OR, NOT
- Parentheses grouping
- Nested expressions
- Operator precedence
- Invalid syntax: missing operators, unmatched parens

**3. Translator Tests (`oql/translator_test.go`):**
- AST to JSON conversion
- Dot-walking detection
- Negate flag handling
- Sort clause translation
- Limit/offset translation
- Complex nested expressions

### Integration Tests (`oql/oql_test.go`)

End-to-end OQL → JSON → SQL → Results flow:
```go
func TestOQLIntegration(t *testing.T) {
    tests := []struct {
        name     string
        oql      string
        expected Query
    }{
        {
            name: "simple filter",
            oql:  "is_active = true limit 10",
            expected: Query{
                Filters: []Filter{{Field: "is_active", Operator: "eq", Value: true}},
                Limit:   IntPtr(10),
            },
        },
        {
            name: "dot-walking with negate",
            oql:  "is_active = true and not software.vendor = 'CrowdStrike'",
            expected: Query{
                Filters: []Filter{
                    {Field: "is_active", Operator: "eq", Value: true},
                    {Field: "software.vendor", Operator: "eq", Value: "CrowdStrike", Negate: true},
                },
            },
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Parse(tt.oql)
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Test Coverage Targets
- Tokenizer: 95%+ (simple, deterministic)
- Parser: 90%+ (recursive logic)
- Translator: 90%+ (mapping logic)
- Overall: 90%+ coverage

### Performance Tests
```go
func TestOQLParserPerformance(t *testing.T) {
    complexQuery := `
        (software.vendor = 'Microsoft' or software.vendor = 'Apple') and
        (findings.severity = 'critical' or findings.epss_score > 0.9) and
        is_active = true
        sort findings.epss_score desc limit 100
    `

    bench := benchmark.New(1000)
    for i := 0; i < 1000; i++ {
        start := time.Now()
        Parse(complexQuery)
        bench.Add(time.Since(start))
    }

    // Should parse in < 1ms
    assert.Less(t, bench.Mean(), 1*time.Millisecond)
}
```

### Fuzzing Tests
```go
func FuzzOQLParser(f *testing.F) {
    f.Add("is_active = true")
    f.Add("software.vendor = 'X'")
    f.Add("invalid!!!")

    f.Fuzz(func(t *testing.T, input string) {
        // Should never panic
        Parse(input)
    })
}
```

---

## 10. UI Integration and User Experience

The OQL language will be integrated into the UI with an editor that provides real-time validation, autocomplete, and a smooth transition between visual and text-based query building.

### UI Components

**1. OQL Query Editor (`OQLQueryEditor.tsx`):**
- Code editor with syntax highlighting (Monaco Editor or CodeMirror)
- Real-time syntax validation
- Inline error display with position markers
- "Explain" button to show JSON/SQL translation
- "Format" button to beautify query

**2. Toggle Between Modes:**
- Switch between "Visual Builder" and "OQL Editor"
- Bidirectional conversion: Visual → OQL and OQL → Visual
- Preserve query state when switching

**3. Autocomplete/Suggestions:**
- Field names with dot-walking (`software.<tab>` → vendor, product_name, ...)
- Operators (=, !=, <, >, like, in, ...)
- Values for enum fields (severity: critical, high, medium, low)
- Keywords (and, or, not, limit, sort, ...)

**4. Query Templates with OQL:**
- Show both OQL and JSON examples
- One-click insert OQL into editor
- Copy/paste OQL queries

### OQL Editor UI Layout
```
┌─────────────────────────────────────────────────────────────┐
│ [Visual Builder] [OQL Editor]                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  is_active = true                                           │
│  and not software.vendor = 'CrowdStrike'                    │
│  sort canonical_name asc                                    │
│  limit 100                                                  │
│                                                             │
│  ─────────────────────────────────────────────────────────  │
│  ✅ Valid query                                             │
│  💡 Tip: Use 'not' for anti-joins (assets without X)        │
│                                                             │
│  [Execute] [Explain] [Format] [Clear]                       │
└─────────────────────────────────────────────────────────────┘
```

### "Explain" Modal
```
OQL Query:
is_active = true and not software.vendor = 'CrowdStrike' limit 100

↓ Translates to ↓

JSON:
{
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true},
    {"field": "software.vendor", "operator": "eq", "value": "CrowdStrike", "negate": true}
  ],
  "limit": 100
}

↓ Generates ↓

SQL:
SELECT * FROM assets_extended AS assets
WHERE assets.tenant_id = $1
  AND assets.is_active = true
  AND NOT EXISTS (
    SELECT 1 FROM software_inventory
    WHERE software_inventory.asset_id = assets.id
      AND software_inventory.tenant_id = $2
      AND software_inventory.vendor = 'CrowdStrike'
  )
LIMIT 100
```

### Error Display
```
is_active = true and (software.vendor = 'Microsoft'
                      ──────────────────────────────
                      ❌ Unclosed parenthesis

                      Expected: ) after 'Microsoft'
                      Position: line 1, column 52

                      [Fix automatically] [Learn more]
```

### Learning Resources
- "OQL Syntax Guide" button → Modal with examples
- "Convert to OQL" tip when using visual builder
- Sample queries in sidebar

---

## 11. Documentation and Examples

Comprehensive documentation for users learning and using OQL.

### Documentation Structure

**1. `docs/oql.md` - Main OQL Reference**
- Quick start guide
- Complete syntax reference
- Operator reference
- Field reference with dot-walking
- Error handling guide
- Performance tips

**2. `docs/oql-examples.md` - Example Queries**
- Common use cases with OQL and JSON equivalents
- Progressive examples (simple → complex)
- Real-world scenarios

**3. Inline Code Comments**
- Parser grammar documentation
- Token types
- AST node types

### Example Query Gallery

**Assets Missing Software**
```
OQL:
is_active = true and not software.vendor = 'CrowdStrike' limit 100

JSON Equivalent:
{
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true},
    {"field": "software.vendor", "operator": "eq", "value": "CrowdStrike", "negate": true}
  ],
  "limit": 100
}
```

**Assets with Exploitable CVEs**
```
OQL:
findings.is_kev = true and findings.effective_status = 'open'
sort findings.epss_score desc limit 50

SQL Generated:
SELECT DISTINCT assets.*
FROM assets_extended AS assets
INNER JOIN findings ON findings.asset_id = assets.id
WHERE findings.tenant_id = $1
  AND findings.is_kev = true
  AND findings.effective_status = 'open'
ORDER BY findings.epss_score DESC
LIMIT 50
```

**Complex Correlation**
```
OQL:
software.product_name = 'Log4j' and findings.cve in ('CVE-2021-44228', 'CVE-2021-45046')
sort findings.last_observed_at desc limit 100
```

**Assets by Pattern Matching**
```
OQL:
hostname like 'web%' and ipv4 is not null
sort canonical_name asc limit 100
```

**Multiple Entity Join**
```
OQL:
software.vendor = 'Oracle' and findings.severity in ('critical', 'high')
not software.vendor = 'Oracle Corporation' limit 100
```

### Quick Reference Card

```
# OQL Quick Reference

## Operators
Comparison: =, !=, <, >, <=, >=
Pattern: like 'value%'
Set: in ('a', 'b', 'c')
Null: is null, is not null

## Logical Operators
and, or, not
Use ( ) for grouping

## Sort/Pagination
sort field asc|desc
limit N
offset N

## Dot-Walking
software.field - Software inventory
findings.field - Vulnerability findings

## Examples
is_active = true
not software.vendor = 'X'
hostname like 'web%'
severity in ('critical', 'high')
```

### Migration Guide (JSON → OQL)

**Before (JSON):**
```json
{
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true},
    {"field": "software.vendor", "operator": "eq", "value": "CrowdStrike", "negate": true}
  ],
  "limit": 100,
  "sort": [{"field": "canonical_name", "order": "asc"}]
}
```

**After (OQL):**
```
is_active = true and not software.vendor = 'CrowdStrike'
limit 100 sort canonical_name asc
```

**Benefits:**
- 70% less boilerplate
- More readable
- Easier to write manually

---

## 12. Implementation Phases

The OQL implementation will be delivered in phases, starting with core functionality and expanding to advanced features.

### Phase 1: Core Parser and API (Foundation)
- Tokenizer implementation
- Recursive descent parser
- AST to JSON translator
- Basic operator support (=, !=, <, >, <=, >=, like, in, is null)
- AND, OR, NOT logical operators
- Parentheses grouping
- `/api/v1/query/oql` endpoint
- Unit tests (90%+ coverage)
- **Target: 3-4 days**

### Phase 2: Advanced Query Features
- Sort clause (single and multi-field)
- Limit and offset clauses
- Dot-walking validation
- Field whitelist enforcement
- Type checking for values
- `/api/v1/query/oql/validate` endpoint
- `/api/v1/query/oql/explain` endpoint
- Integration tests with real data
- **Target: 2-3 days**

### Phase 3: Error Handling and DX
- Comprehensive error messages
- Position-aware error reporting
- Suggestions for common mistakes
- Performance optimization (parser caching)
- Fuzzing tests for robustness
- **Target: 2 days**

### Phase 4: UI Integration
- OQL query editor component (Monaco/CodeMirror)
- Real-time syntax validation
- Toggle between Visual Builder and OQL
- Autocomplete for fields, operators, keywords
- Query templates with OQL
- "Explain" modal
- Error highlighting and suggestions
- **Target: 3-4 days**

### Phase 5: Documentation and Polish
- Complete OQL reference documentation
- Example query gallery
- Migration guide (JSON → OQL)
- Quick reference card
- UI help text and tooltips
- Performance tuning and optimization
- **Target: 2 days**

### Total Timeline
**~12-15 days**

### Success Criteria
- ✅ All existing JSON queries translatable to OQL
- ✅ Parse time < 1ms for complex queries
- ✅ 90%+ test coverage
- ✅ Clear, actionable error messages
- ✅ UI supports both visual and OQL modes
- ✅ Documentation complete and verified
- ✅ No breaking changes to existing API

### Backward Compatibility
- Existing `/api/v1/query/unified` JSON endpoint unchanged
- OQL is additive only
- UI can use either format (internal conversion)

---

## Appendix: Complete OQL Examples

### Example 1: Simple Filter
```
is_active = true
```

### Example 2: Multiple Filters with AND
```
is_active = true and ipv4 is not null
```

### Example 3: Pattern Matching
```
hostname like 'web%' and is_active = true
```

### Example 4: Set Membership
```
findings.severity in ('critical', 'high') and findings.effective_status = 'open'
```

### Example 5: Dot-Walking with Negation
```
is_active = true and not software.vendor = 'CrowdStrike'
```

### Example 6: Complex Grouping
```
(is_active = true and software.vendor = 'Microsoft') or (software.vendor = 'Apple' and software.product_name = 'macOS')
```

### Example 7: Multiple Entity Joins
```
software.product_name = 'Log4j' and findings.cve = 'CVE-2021-44228'
```

### Example 8: Sorting and Pagination
```
is_active = true limit 100 offset 0 sort canonical_name asc
```

### Example 9: Numeric Comparisons
```
findings.epss_score > 0.9 and findings.is_kev = true
```

### Example 10: Full Featured Query
```
is_active = true
and not software.vendor = 'CrowdStrike'
and findings.severity in ('critical', 'high')
sort findings.epss_score desc, canonical_name asc
limit 100
```

---

## Summary

OQL provides a concise, SQL-like query language that reduces query boilerplate by ~70% while maintaining full compatibility with the existing unified query API. The expression-based syntax with dot-walking makes cross-entity queries intuitive for both analysts and automation scripts.

**Key Benefits:**
- **Concise:** 70% less verbose than JSON
- **Intuitive:** Expression-based syntax with SQL familiarity
- **Powerful:** Full support for dot-walking, joins, aggregations
- **Safe:** Clear error messages with position-aware suggestions
- **Fast:** < 1ms parse time for complex queries
- **Compatible:** Translates to existing JSON format, no breaking changes

The design is complete and ready for implementation following the phased approach outlined in Section 12.
