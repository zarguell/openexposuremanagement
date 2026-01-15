# Unified Queries - Dot-Walking Syntax

## Overview

The unified query API enables cross-entity correlation queries using **dot-walking syntax** - a simple way to filter across related entities without explicit JOIN configuration.

## API Endpoint

```
POST /api/v1/query/unified
```

## Query Syntax

### Basic Structure

```json
{
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true}
  ],
  "limit": 100
}
```

### Dot-Walking for Related Entities

Use dot notation to reference fields on related entities:

- `software.vendor` - Software vendor field
- `software.product_name` - Software product name
- `findings.severity` - Finding severity
- `findings.cve` - CVE ID

**Example: Assets with specific software**

```json
{
  "filters": [
    {"field": "software.vendor", "operator": "eq", "value": "Microsoft"}
  ],
  "limit": 100
}
```

This automatically generates an INNER JOIN to software_inventory and filters by vendor.

### Anti-Join (NOT EXISTS) Pattern

Use `negate: true` to find assets **without** matching related records:

**Example: Assets without CrowdStrike**

```json
{
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true},
    {"field": "software.vendor", "operator": "eq", "value": "CrowdStrike", "negate": true}
  ],
  "limit": 100
}
```

This generates a NOT EXISTS subquery to find active assets that don't have CrowdStrike installed.

### Multiple Entity Joins

Filter on multiple related entities in a single query:

**Example: Assets with Log4j and CVE-2021-44228**

```json
{
  "filters": [
    {"field": "software.product_name", "operator": "eq", "value": "Log4j"},
    {"field": "findings.cve", "operator": "eq", "value": "CVE-2021-44228"}
  ],
  "limit": 100
}
```

This generates INNER JOINs to both software_inventory and findings.

## Common Use Cases

### Assets Missing Specific Software

Find assets without CrowdStrike installed:

```json
{
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true},
    {"field": "software.vendor", "operator": "eq", "value": "CrowdStrike", "negate": true}
  ],
  "limit": 100
}
```

**SQL equivalent:**
```sql
SELECT * FROM assets_extended AS assets
WHERE assets.is_active = true
AND NOT EXISTS (
  SELECT 1 FROM software_inventory
  WHERE software_inventory.asset_id = assets.id
  AND software_inventory.vendor = 'CrowdStrike'
)
```

### Assets with Exploitable CVEs

Find assets with CISA KEV vulnerabilities:

```json
{
  "filters": [
    {"field": "findings.is_kev", "operator": "eq", "value": true},
    {"field": "findings.effective_status", "operator": "eq", "value": "open"}
  ],
  "limit": 100
}
```

**SQL equivalent:**
```sql
SELECT DISTINCT assets.* FROM assets_extended AS assets
INNER JOIN findings ON findings.asset_id = assets.id
WHERE findings.is_kev = true
AND findings.effective_status = 'open'
```

### Software Vulnerability Correlation

Find assets with specific software and CVEs:

```json
{
  "filters": [
    {"field": "software.product_name", "operator": "eq", "value": "Apache Tomcat"},
    {"field": "findings.severity", "operator": "eq", "value": "critical"}
  ],
  "limit": 100
}
```

## Translation Rules

The backend automatically determines JOIN logic based on field references:

1. **No dot notation** → Simple query on primary entity (assets)
2. **Dot notation without negate** → INNER JOIN (must have matching record)
3. **Dot notation with negate: true** → NOT EXISTS subquery (anti-join)

**DISTINCT handling:**
- INNER JOINs automatically add DISTINCT to prevent duplicate rows
- NOT EXISTS queries don't need DISTINCT (no duplicates possible)

## Supported Fields

### Primary Entity (Assets)
- `canonical_name`
- `hostname_norm`
- `shortname_norm`
- `ipv4`
- `first_seen_at`
- `last_seen_at`
- `is_active`

### Software (via `software.*`)
- `software.vendor`
- `software.product_name`
- `software.version`
- `software.cpe_string`
- `software.install_path`
- `software.first_seen_at`
- `software.last_seen_at`

### Findings (via `findings.*`)
- `findings.severity`
- `findings.scanner_status`
- `findings.effective_status`
- `findings.cve`
- `findings.epss_score`
- `findings.is_kev`
- `findings.first_observed_at`
- `findings.last_observed_at`

## Performance Guidelines

- **Max rows**: 5,000 (enforced)
- **Timeout**: 5 seconds
- **Filters**: Recommended on primary entity for performance
- **Indexes**: Composite indexes support common join patterns

## Known Limitations (MVP)

- **No N-way joins** (3+ entities) - only software and findings supported
- **No aggregations on joined fields** - aggregations work on primary entity only
- **No result caching** - each query executes against database

## Templates

The API provides pre-built query templates:

| Template ID | Name | Description |
|-------------|------|-------------|
| `missing_software` | Assets Missing Critical Software | Find assets without specific software |
| `exploitable_cves` | Assets with Exploitable CVEs | High EPSS or CISA KEV |
| `software_vulnerabilities` | Vulnerabilities by Software | Findings affecting specific products |

## Response Format

```json
{
  "data": [
    {
      "id": 1,
      "canonical_name": "server1.local",
      "hostname_norm": "server1",
      "is_active": true
    }
  ],
  "meta": {
    "total_rows": 1,
    "execution_time_ms": 25,
    "has_more": false
  }
}
```

## Benefits of Dot-Walking Syntax

- **Simpler**: No need to understand JOIN types or ON conditions
- **Intuitive**: Dot notation is familiar from JSON/JavaScript
- **Less backend exposure**: Users don't see SQL implementation details
- **Easier to extend**: Add new relationships without breaking existing queries
