# OQL Query Examples

This document provides real-world OQL query examples organized by use case.

## Table of Contents

- [Asset Discovery](#asset-discovery)
- [Vulnerability Management](#vulnerability-management)
- [Software Inventory](#software-inventory)
- [Compliance & Security](#compliance--security)
- [Performance & Operations](#performance--operations)
- [Complex Queries](#complex-queries)

## Asset Discovery

### List All Active Assets

```sql
is_active = true order by canonical_name limit 100
```

### Find Inactive Assets

```sql
is_active = false order by last_seen_at desc limit 50
```

### Find Assets by Hostname Pattern

```sql
hostname like "web%" AND is_active = true
```

### Find Assets Without Owner Team

```sql
owner_team_id is null AND is_active = true
```

### Find Recently Created Assets

```sql
first_seen_at >= "2024-01-01" order by first_seen_at desc limit 20
```

### Find Assets Not Seen Recently

```sql
last_seen_at < "2024-01-01" AND is_active = true order by last_seen_at asc limit 50
```

## Vulnerability Management

### Find Assets with Critical Vulnerabilities

```sql
findings.severity = "critical" AND findings.effective_status = "open"
```

### Find Assets with Exploitable CVEs (High EPSS)

```sql
findings.severity in ("critical", "high") AND findings.epss_score > 0.9
order by findings.epss_score desc limit 100
```

### Find Assets with Known Exploited Vulnerabilities (KEV)

```sql
findings.is_kev = true AND findings.effective_status = "open"
order by findings.kev_due_date asc limit 50
```

### Find Assets with Log4j Vulnerabilities

```sql
findings.cve = "CVE-2021-44228" AND findings.effective_status = "open"
```

### Find Assets with Any CVE from 2024

```sql
findings.cve like "CVE-2024-%" AND findings.effective_status = "open"
```

### Find Critical Vulnerabilities Approaching KEV Due Date

```sql
findings.severity = "critical" AND findings.is_kev = true
AND findings.kev_due_date < "2024-03-01"
order by findings.kev_due_date asc limit 20
```

### Count Assets by Severity (via Dashboard)

```sql
-- Note: This requires dashboard aggregation endpoint
-- For individual queries, use:
findings.severity = "critical" limit 5000
findings.severity = "high" limit 5000
findings.severity = "medium" limit 5000
findings.severity = "low" limit 5000
```

## Software Inventory

### Find Assets with Microsoft Software

```sql
software.vendor = "Microsoft" AND is_active = true
order by canonical_name limit 100
```

### Find Assets with Specific Product

```sql
software.product_name = "Log4j" AND is_active = true
```

### Find Assets with Specific Software Version

```sql
software.vendor = "Apache" AND software.product_name = "Tomcat"
AND software.version = "9.0.1"
```

### Find Assets by CPE Pattern

```sql
software.cpe_23 like "cpe:2.3:a:microsoft:windows:%" AND is_active = true
```

### Find Assets with Multiple Software Products

```sql
software.vendor = "Microsoft" AND software.product_name = "SQL Server"
OR software.vendor = "Oracle" AND software.product_name = "Database"
```

### Find Software Installed Recently

```sql
software.first_seen_at >= "2024-01-01" AND is_active = true
order by software.first_seen_at desc limit 50
```

### Find Software No Longer Seen

```sql
software.last_seen_at < "2024-01-01" AND is_active = true
order by software.last_seen_at asc limit 50
```

## Compliance & Security

### Assets Missing EDR Software

```sql
is_active = true AND NOT software.vendor = "CrowdStrike"
AND NOT software.vendor = "SentinelOne"
AND NOT software.vendor = "Carbon Black"
```

### Assets Without Antivirus

```sql
is_active = true AND NOT software.vendor = "Symantec"
AND NOT software.vendor = "McAfee"
AND NOT software.vendor = "Trend Micro"
AND NOT software.vendor = "Kaspersky"
```

### Internet-Facing Assets with Critical Vulnerabilities

```sql
internet_facing = true AND findings.severity = "critical"
AND findings.effective_status = "open"
```

### Production Assets with Unpatched Software

```sql
environment = "production" AND software.vendor = "Adobe"
AND software.version < "2024.0.1"
```

### Assets Out of Compliance (Missing Patches)

```sql
is_active = true AND (
  NOT software.vendor = "Microsoft" OR software.version < "10.0.19045"
) AND findings.cve like "CVE-2024-%"
```

### DMZ Assets with Any Vulnerabilities

```sql
environment = "DMZ" AND findings.effective_status = "open"
order by findings.severity desc limit 100
```

## Performance & Operations

### Top 100 Assets by Finding Count

```sql
-- Note: This requires aggregation support in future release
-- For now, use basic filter:
is_active = true AND findings.severity in ("critical", "high")
order by canonical_name asc limit 100
```

### Find Assets with Most Software Installed

```sql
is_active = true order by canonical_name limit 100
-- Filter for software-rich assets after getting results
```

### Recently Updated Assets

```sql
last_seen_at >= "2024-01-01" AND is_active = true
order by last_seen_at desc limit 20
```

### Stale Assets (Not Seen in 30 Days)

```sql
last_seen_at < (current_date - interval '30 days') AND is_active = true
order by last_seen_at asc limit 100
```

## Complex Queries

### High-Risk Assets: Critical + Exploitable + Exposed

```sql
is_active = true
AND findings.severity = "critical"
AND findings.epss_score > 0.9
AND findings.is_kev = true
AND internet_facing = true
order by findings.epss_score desc limit 20
```

### Production Assets Missing Critical Software

```sql
environment = "production"
AND is_active = true
AND NOT software.vendor = "CrowdStrike"
AND (findings.severity = "critical" OR findings.severity = "high")
order by findings.epss_score desc limit 50
```

### Assets with Legacy Software (Old Versions)

```sql
is_active = true
AND software.vendor = "Microsoft"
AND software.product_name = "Windows"
AND software.version like "7%"
OR software.version like "Server 2008%"
```

### Vulnerable Assets in Critical Environments

```sql
(environment = "production" OR environment = "DMZ")
AND is_active = true
AND findings.effective_status = "open"
AND findings.severity in ("critical", "high")
order by findings.severity desc, findings.epss_score desc limit 100
```

### Cross-Entity: Software with Vulnerabilities

```sql
software.vendor = "Apache"
AND findings.severity = "critical"
AND findings.cve like "CVE-2024-%"
```

### Comprehensive Security Posture Query

```sql
is_active = true
AND (internet_facing = true OR environment = "DMZ")
AND (
  findings.severity = "critical"
  OR findings.is_kev = true
  OR findings.epss_score > 0.9
)
AND NOT software.vendor in ("CrowdStrike", "SentinelOne")
order by findings.epss_score desc, findings.last_observed_at desc limit 50
```

## Query Templates

### Template 1: Vulnerable Assets Dashboard

```sql
-- Critical vulnerabilities
findings.severity = "critical" AND findings.effective_status = "open"

-- High severity with high EPSS
findings.severity = "high" AND findings.epss_score > 0.8 AND findings.effective_status = "open"

-- Known exploited vulnerabilities
findings.is_kev = true AND findings.effective_status = "open"
```

### Template 2: Software Compliance

```sql
-- Missing EDR
is_active = true AND NOT software.vendor = "CrowdStrike"

-- Outdated software
software.vendor = "Microsoft" AND software.version < "10.0.19045" AND is_active = true

-- End-of-life software
software.vendor = "Microsoft" AND software.version like "Server 2008%" AND is_active = true
```

### Template 3: Risk-Based Prioritization

```sql
-- Tier 1: Critical + Exploitable + Exposed
findings.severity = "critical" AND findings.epss_score > 0.9 AND internet_facing = true

-- Tier 2: Critical + High EPSS
findings.severity = "critical" AND findings.epss_score > 0.7

-- Tier 3: High + KEV
findings.severity = "high" AND findings.is_kev = true
```

## Tips for Building Queries

1. **Start with filters**: Begin with `is_active = true` to narrow scope
2. **Add conditions incrementally**: Add one condition at a time
3. **Use EXPLAIN endpoint**: See how OQL translates to SQL
4. **Set LIMIT early**: Start with small limits (10-50) for faster results
5. **Test with VALIDATE**: Check syntax before executing complex queries
6. **Use parentheses**: Group complex logical expressions explicitly
7. **Leverage dot-walking**: Use `software.vendor` instead of manual joins
8. **Combine with NOT**: Find missing software with `NOT software.vendor = "X"`

## Converting JSON to OQL

### Before (JSON)

```json
{
  "filters": [
    {"field": "is_active", "operator": "eq", "value": true},
    {"field": "findings.severity", "operator": "eq", "value": "critical"},
    {"field": "findings.epss_score", "operator": "gt", "value": 0.9}
  ],
  "sort": [{"field": "findings.epss_score", "order": "desc"}],
  "limit": 20
}
```

### After (OQL)

```sql
is_active = true AND findings.severity = "critical" AND findings.epss_score > 0.9
order by findings.epss_score desc limit 20
```

## Real-World Scenarios

### Scenario 1: Quarterly Security Review

Find all critical vulnerabilities that need attention before Q1 review:

```sql
findings.severity = "critical"
AND findings.effective_status = "open"
AND findings.first_observed_at < "2024-01-01"
order by findings.first_observed_at asc limit 200
```

### Scenario 2: Incident Response

During an incident, find all assets vulnerable to a specific CVE:

```sql
findings.cve = "CVE-2024-1234"
AND findings.effective_status = "open"
order by findings.last_observed_at desc limit 5000
```

### Scenario 3: Software Audit

Find all assets running software from a specific vendor for licensing audit:

```sql
software.vendor = "Oracle" AND is_active = true
order by canonical_name asc limit 5000
```

### Scenario 4: Patch Management

Prioritize patching for assets with the most exploitable vulnerabilities:

```sql
findings.epss_score > 0.8
AND findings.effective_status = "open"
AND is_active = true
order by findings.epss_score desc, findings.severity desc limit 100
```

### Scenario 5: New Asset Onboarding

Check if newly discovered assets have proper security software:

```sql
first_seen_at >= "2024-01-01"
AND is_active = true
AND NOT software.vendor = "CrowdStrike"
order by first_seen_at desc limit 50
```

## Common Query Patterns

### Pattern 1: Existence Check

Find assets that have ANY of a type:

```sql
software.cpe_23 is not null AND is_active = true
```

### Pattern 2: Exclusion Filter

Find assets excluding specific values:

```sql
is_active = true AND software.vendor != "Microsoft" AND software.vendor != "Oracle"
```

### Pattern 3: Range Query

Find assets within a range:

```sql
findings.cvss_score >= 7.0 AND findings.cvss_score < 10.0
```

### Pattern 4: Multi-Condition Filter

Find assets matching ALL conditions:

```sql
is_active = true
AND findings.severity = "critical"
AND findings.epss_score > 0.9
AND findings.is_kev = true
```

### Pattern 5: Temporal Filter

Find assets based on time:

```sql
findings.first_observed_at >= "2024-01-01"
AND findings.first_observed_at < "2024-02-01"
```

## Learning Resources

- Start with simple queries and add complexity gradually
- Use the `/validate` endpoint to check syntax without executing
- Use the `/explain` endpoint to see how queries translate to SQL
- Review [OQL Syntax Reference](./oql.md) for complete syntax details
- Check [API Documentation](./architecture.md) for endpoint details
