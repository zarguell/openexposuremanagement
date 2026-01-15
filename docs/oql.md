# OQL (Open Query Language)

## Overview

**OQL (Open Query Language)** is a concise, SQL-like query language for the Open Exposure Management platform. It provides a more intuitive and compact way to write queries compared to JSON, reducing boilerplate by approximately 70%.

### Key Benefits

- **Concise Syntax**: Write complex queries in a single line of text
- **SQL-Like**: Familiar syntax for anyone who knows SQL
- **Dot-Walking**: Easy cross-entity queries using dot notation
- **Anti-Joins**: Simple syntax for "assets without X" queries
- **Type Safety**: Automatic validation with helpful error messages

## Query Structure

OQL queries are expression-based and consist of filters, logical operators, sorting, and pagination:

```
[condition] [AND|OR condition] ... [order by field asc|desc] [limit N] [offset M]
```

### Basic Example

```sql
is_active = true AND software.vendor = "Microsoft" order by canonical_name limit 10
```

## Syntax Reference

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `=` | Equal to | `is_active = true` |
| `!=` or `<>` | Not equal to | `severity != "low"` |
| `>` | Greater than | `epss_score > 0.9` |
| `>=` | Greater than or equal | `cvss_score >= 7.0` |
| `<` | Less than | `last_seen_at < "2024-01-01"` |
| `<=` | Less than or equal | `epss_percentile <= 0.5` |
| `like` | Pattern matching | `hostname like "web%"` |
| `in` | In list | `severity in ("critical", "high")` |
| `is null` | Is null | `owner_team_id is null` |
| `is not null` | Is not null | `canonical_name is not null` |

### Logical Operators

| Operator | Description | Example | Precedence |
|----------|-------------|---------|------------|
| `NOT` | Logical negation | `NOT is_active` | Highest |
| `AND` | Logical and | `is_active = true AND software.vendor = "Microsoft"` | Medium |
| `OR` | Logical or | `severity = "critical" OR severity = "high"` | Lowest |

**Note**: Operator precedence follows SQL standards: NOT > AND > OR. Use parentheses to control evaluation order.

### Dot-Walking (Cross-Entity Queries)

OQL supports dot notation for querying related entities:

```sql
-- Query assets with Microsoft software
software.vendor = "Microsoft"

-- Query assets with critical findings
findings.severity = "critical"

-- Query assets with Log4j vulnerabilities
software.cpe_23 like "cpe:2.3:a:apache:log4j:%"

-- Deep dot-walking
findings.definition.cve = "CVE-2021-44228"
```

### Anti-Joins (NOT Queries)

Use the `NOT` operator to find assets **without** a specific property:

```sql
-- Find assets without CrowdStrike
NOT software.vendor = "CrowdStrike"

-- Find assets without critical vulnerabilities
NOT findings.severity = "critical"

-- Combine with other conditions
is_active = true AND NOT software.vendor = "CrowdStrike"
```

**Note**: Anti-join queries generate `NOT EXISTS` subqueries for optimal performance.

### Sorting

Sort results by one or more fields:

```sql
-- Single field sort
order by canonical_name asc

-- Descending order
order by last_seen_at desc

-- Multiple sorts (only last one is used in MVP)
order by canonical_name asc, last_seen_at desc
```

### Pagination

Control result set size:

```sql
-- Limit results
limit 100

-- Skip first N results
offset 50

-- Combine both
limit 50 offset 100
```

## Data Types

### Literals

| Type | Examples |
|------|----------|
| String | `"Microsoft"`, `"webserver01"`, `"CVE-2024-1234"` |
| Number | `10`, `3.14`, `0.95` |
| Boolean | `true`, `false` |
| Null | `null` (used with `is null`/`is not null`) |

### Field Types

Different fields support different operators:

- **Text fields** (`canonical_name`, `hostname`, `vendor`): `=`, `!=`, `like`, `in`
- **Numeric fields** (`cvss_score`, `epss_score`): `=`, `!=`, `>`, `>=`, `<`, `<=`
- **Boolean fields** (`is_active`): `=`, `!=`
- **Date fields** (`last_seen_at`, `first_observed_at`): `=`, `!=`, `>`, `>=`, `<`, `<=`
- **Array fields** (via dot-walking): `=`, `!=`, `in`

## Complex Queries

### Grouping with Parentheses

Control evaluation order with parentheses:

```sql
-- Find critical findings that are either exploitable or in DMZ
(severity = "critical" AND epss_score > 0.9) OR environment = "DMZ"

-- Find assets with software from Microsoft OR Oracle, but only if active
is_active = true AND (software.vendor = "Microsoft" OR software.vendor = "Oracle")
```

### Combining Conditions

```sql
-- Complex real-world example
is_active = true
AND NOT software.vendor = "CrowdStrike"
AND (findings.severity = "critical" OR findings.severity = "high")
AND findings.epss_score > 0.8
order by findings.epss_score desc
limit 20
```

### Pattern Matching with LIKE

Use `%` as a wildcard in LIKE patterns:

```sql
-- Hostnames starting with "web"
hostname like "web%"

-- Hostnames ending with ".local"
hostname like "%.local"

-- Hostnames containing "server"
hostname like "%server%"

-- CPE patterns
software.cpe_23 like "cpe:2.3:a:microsoft:%"
```

## Performance Considerations

### Query Optimization

1. **Use specific filters**: More filters = faster queries
   ```sql
   -- Good: Specific filters
   is_active = true AND software.vendor = "Microsoft"

   -- Slower: Broad query
   is_active = true
   ```

2. **Avoid leading wildcards in LIKE**:
   ```sql
   -- Slower: Full scan
   hostname like "%server%"

   -- Faster: Index-friendly
   hostname like "server%"
   ```

3. **Use LIMIT for large result sets**:
   ```sql
   -- Good: Limited results
   is_active = true limit 100

   -- Slower: Unbounded results
   is_active = true
   ```

### Performance Guardrails

The API enforces these limits:

- **Max results**: 5,000 rows per query
- **Query timeout**: 5 seconds
- **Max dot-walk depth**: 3 levels

## Error Messages

OQL provides helpful error messages with line and column information:

```
Error: unexpected token 'AND' at line 1:20
  is_active = true AND AND software.vendor = "Microsoft"
                     ^
Expected: identifier, string, number, or keyword
```

## Comparison: OQL vs JSON

### JSON Query (Verbose)

```json
{
  "filters": [
    {
      "field": "is_active",
      "operator": "eq",
      "value": true
    },
    {
      "field": "software.vendor",
      "operator": "eq",
      "value": "Microsoft"
    }
  ],
  "sort": [
    {
      "field": "canonical_name",
      "order": "asc"
    }
  ],
  "limit": 10
}
```

### OQL Query (Concise)

```sql
is_active = true AND software.vendor = "Microsoft" order by canonical_name limit 10
```

**Reduction**: ~70% less boilerplate

## API Endpoints

### Execute OQL Query

```http
POST /api/v1/query/oql
Content-Type: application/json

{
  "query": "is_active = true AND software.vendor = \"Microsoft\" limit 10"
}
```

### Validate OQL Syntax

```http
POST /api/v1/query/oql/validate
Content-Type: application/json

{
  "query": "is_active = true AND software.vendor = \"Microsoft\""
}
```

Returns:
```json
{
  "valid": true,
  "errors": []
}
```

### Explain OQL Translation

```http
POST /api/v1/query/oql/explain
Content-Type: application/json

{
  "query": "is_active = true limit 10"
}
```

Returns:
```json
{
  "unified_query": {
    "primary_entity": "assets",
    "filters": [...],
    "limit": 10
  },
  "sql": "SELECT ... FROM assets ... LIMIT 10",
  "args": [...]
}
```

## Best Practices

1. **Start simple**: Build complex queries incrementally
2. **Use validation**: Call `/validate` endpoint before executing
3. **Add filters gradually**: Test each filter as you add it
4. **Use EXPLAIN**: Understand how your query translates to SQL
5. **Set reasonable limits**: Start with small LIMIT values
6. **Quote strings properly**: Always use double quotes for strings
7. **Check for typos**: Error messages show exact line/column of issues

## Limitations (MVP)

- **No SELECT clause**: Always returns all fields for the primary entity
- **No JOIN syntax**: Use dot-walking for cross-entity queries
- **No aggregations**: No `count()`, `sum()`, `avg()` in MVP
- **No GROUP BY**: Use dashboard endpoints for aggregations
- **No subqueries**: Except for anti-join pattern (`NOT field = value`)
- **Two-way joins only**: Can join assets→software or assets→findings, not both

## Future Enhancements

Planned for post-MVP:

- Aggregations (`count()`, `sum()`, `avg()`)
- `GROUP BY` support
- N-way joins (3+ entities)
- Subquery expressions
- `SELECT` clause (field projection)
- Window functions
- CTEs (Common Table Expressions)

## See Also

- [OQL Examples](./oql-examples.md) - Real-world query examples
- [API Documentation](./architecture.md) - Complete API reference
- [Query Framework](./unified-queries.md) - Unified query system
