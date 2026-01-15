# Simplified Unified Query API Design

## Problem

The current unified query API is too complex and exposes too much SQL implementation detail:
- Users must specify JOIN types (left, inner, etc.)
- Users must specify ON conditions
- Users must think about table relationships
- It's basically just SQL in JSON - not user-friendly

## Proposed Solution: Dot-Walking API

Instead of exposing JOIN semantics, use dot notation to "walk" across related entities. The backend automatically figures out the right JOIN logic.

### Entity Relationships

Assets have natural relationships to:
- `software` → asset's installed software (one-to-many)
- `findings` → asset's vulnerability findings (one-to-many)

### Field Reference Syntax

Use dot notation to reference fields on related entities:
- `software.vendor` → software inventory vendor field
- `software.product_name` → software product name
- `findings.severity` → finding severity
- `findings.cve` → CVE ID

### Filter Syntax

```json
{
  "filters": [
    // Filter on primary entity (assets)
    {"field": "is_active", "operator": "eq", "value": true},

    // Filter on related entity using dot notation
    {"field": "software.vendor", "operator": "eq", "value": "CrowdStrike"},

    // Negate filter (NOT EXISTS for related entities)
    {"field": "software.vendor", "operator": "eq", "value": "CrowdStrike", "negate": true}
  ]
}
```

### Examples

#### Assets Without CrowdStrike
```json
{
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true},
    {"field": "software.vendor", "operator": "eq", "value": "CrowdStrike", "negate": true}
  ],
  "limit": 50
}
```
**SQL equivalent:**
```sql
SELECT * FROM assets
WHERE is_active = true
AND NOT EXISTS (
  SELECT 1 FROM software_inventory
  WHERE software_inventory.asset_id = assets.id
  AND software_inventory.vendor = 'CrowdStrike'
)
```

#### Assets with Critical CVEs
```json
{
  "filters": [
    {"field": "findings.severity", "operator": "eq", "value": "critical"},
    {"field": "findings.effective_status", "operator": "eq", "value": "open"}
  ],
  "limit": 50
}
```
**SQL equivalent:**
```sql
SELECT DISTINCT assets.* FROM assets
INNER JOIN findings ON findings.asset_id = assets.id
WHERE findings.severity = 'critical'
AND findings.effective_status = 'open'
```

#### Assets with Log4j and CVE-2021-44228
```json
{
  "filters": [
    {"field": "software.product_name", "operator": "eq", "value": "Log4j"},
    {"field": "findings.cve", "operator": "eq", "value": "CVE-2021-44228"}
  ],
  "limit": 50
}
```
**SQL equivalent:**
```sql
SELECT DISTINCT assets.* FROM assets
INNER JOIN software_inventory ON software_inventory.asset_id = assets.id
INNER JOIN findings ON findings.asset_id = assets.id
WHERE software_inventory.product_name = 'Log4j'
AND findings.cve = 'CVE-2021-44228'
```

### Translation Rules

The backend automatically determines JOIN logic based on field references:

1. **No related entity fields** → Simple query on primary entity
2. **Related entity without negate** → INNER JOIN (must have matching record)
3. **Related entity with negate: true** → NOT EXISTS subquery

### Implementation Plan

1. **Update parser** to recognize dot notation in field names
2. **Auto-detect relationships** from field prefixes (software.*, findings.*)
3. **Generate appropriate SQL**:
   - INNER JOIN for positive filters
   - NOT EXISTS for negated filters
   - Automatically add DISTINCT when joining to prevent duplicates
4. **Remove explicit JOIN config** from API (backward compatibility layer?)
5. **Update UI** to use simpler dot-walking syntax
6. **Remove join.filter** concept (too confusing)

### Benefits

- **Simpler**: No need to understand JOIN types
- **Intuitive**: Dot notation is familiar from JSON/JavaScript
- **Less backend exposure**: Users don't see SQL implementation details
- **Easier to extend**: Add new relationships without breaking existing queries

### Migration Strategy

For backward compatibility:
1. Keep old API working for now
2. Add new simplified API
3. Update UI to use new API
4. Deprecate old API in future version

### Open Questions

1. Should we support dot notation for aggregations? E.g., `count: {"field": "software.id"}`
2. Should we support multiple levels of dot walking? E.g., `findings.definition.title`
3. How to handle sorting on joined fields? E.g., `sort: [{"field": "software.product_name"}]`
