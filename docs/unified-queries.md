# Unified Queries

## Overview

The unified query API enables cross-entity correlation queries using 2-way LEFT JOINs between assets, software inventory, and findings.

## API Endpoint

```
POST /api/v1/query/unified
```

## Query Syntax

### Basic Structure

```json
{
  "primary_entity": "assets",
  "join": {
    "entity": "software_inventory",
    "type": "left",
    "on": {
      "primary": "id",
      "joined": "asset_id"
    }
  },
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true}
  ],
  "limit": 100
}
```

### Supported Entity Relationships

| Primary Entity | Join Entity | On Condition |
|----------------|-------------|--------------|
| assets | software_inventory | assets.id = software_inventory.asset_id |
| assets | findings | assets.id = findings.asset_id |

## Common Use Cases

### Assets Missing Specific Software

Find assets without CrowdStrike installed:

```json
{
  "primary_entity": "assets",
  "join": {
    "entity": "software_inventory",
    "type": "left",
    "on": {"primary": "id", "joined": "asset_id"}
  },
  "filters": [
    {"field": "product_name", "operator": "eq", "value": "CrowdStrike Falcon"}
  ],
  "limit": 100
}
```

Returns assets where `product_name` is NULL (software not installed).

### Assets with Exploitable CVEs

Find assets with high EPSS scores:

```json
{
  "primary_entity": "assets",
  "join": {
    "entity": "findings",
    "type": "left",
    "on": {"primary": "id", "joined": "asset_id"}
  },
  "filters": [
    {"field": "epss_score", "operator": "gte", "value": 0.9}
  ],
  "limit": 100
}
```

### Software Vulnerability Correlation

Find CVEs affecting specific software:

```json
{
  "primary_entity": "findings",
  "filters": [
    {"field": "cve", "operator": "eq", "value": "CVE-2021-44228"}
  ],
  "limit": 50
}
```

Findings view already includes asset information.

## Performance Guidelines

- **Max rows**: 5,000 (enforced)
- **Timeout**: 5 seconds
- **Join type**: LEFT JOIN only (prevents data explosion)
- **Filters**: Required on primary entity for performance
- **Indexes**: Composite indexes support common join patterns

## Known Limitations (MVP)

- **No N-way joins** (3+ entities) - only 2-way LEFT JOIN supported
- **No aggregations on joined fields** - aggregations work on primary entity only
- **No subquery pushdown optimization** - filters are applied after JOIN
- **Template parameter substitution not yet implemented** - placeholders like `{{software_name}}` require manual replacement
- **UI query builder not yet implemented** - backend-only API at this time
- **No result caching** - each query executes against database

## Future Enhancements

- N-way JOIN support (3+ entity joins)
- Subquery pushdown for better performance
- Query result caching (5-minute TTL)
- Materialized views for common join patterns
- Template parameter substitution in backend
- UI query builder with visual interface
- Custom dashboard builder with unified query widgets

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
      "vendor": "Microsoft",
      "product_name": "Windows Server",
      "cve": "CVE-2021-44228",
      "epss_score": 0.95,
      "is_kev": true
    }
  ],
  "meta": {
    "total_rows": 1,
    "execution_time_ms": 25,
    "has_more": false
  }
}
```
