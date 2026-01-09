-- +migrate Up
-- Create materialized views for dashboard analytics

-- Dashboard counts by effective status
CREATE MATERIALIZED VIEW mv_dashboard_counts AS
SELECT
    tenant_id,
    effective_status,
    COUNT(*) as count
FROM finding_instances
GROUP BY tenant_id, effective_status;

-- Create unique index for CONCURRENTLY refresh
CREATE UNIQUE INDEX mv_dashboard_counts_unique_idx ON mv_dashboard_counts(tenant_id, effective_status);

-- Dashboard summary for assets
CREATE MATERIALIZED VIEW mv_dashboard_assets AS
SELECT
    tenant_id,
    COUNT(*) as total_assets,
    COUNT(*) FILTER (WHERE is_active = TRUE) as active_assets,
    MAX(last_seen_at) as most_recent_asset_activity
FROM assets
GROUP BY tenant_id;

-- Create unique index for CONCURRENTLY refresh
CREATE UNIQUE INDEX mv_dashboard_assets_unique_idx ON mv_dashboard_assets(tenant_id);

-- Dashboard summary for open findings
CREATE MATERIALIZED VIEW mv_dashboard_open_findings AS
SELECT
    fi.tenant_id,
    COUNT(*) FILTER (WHERE fi.effective_status = 'open') as open_count,
    COUNT(*) FILTER (WHERE fi.effective_status IN ('accepted_risk', 'false_positive')) as suppressed_count,
    COUNT(*) FILTER (WHERE fd.severity_default = 'Critical') as critical_open_count,
    COUNT(*) FILTER (WHERE fd.severity_default = 'High') as high_open_count,
    MAX(fi.last_observed_at) as most_recent_finding
FROM finding_instances fi
JOIN finding_definitions fd ON fi.definition_uid = fd.definition_uid
GROUP BY fi.tenant_id;

-- Create unique index for CONCURRENTLY refresh
CREATE UNIQUE INDEX mv_dashboard_open_findings_unique_idx ON mv_dashboard_open_findings(tenant_id);

-- Indexes for query performance
CREATE INDEX idx_mv_dashboard_counts_tenant ON mv_dashboard_counts(tenant_id);
CREATE INDEX idx_mv_dashboard_assets_tenant ON mv_dashboard_assets(tenant_id);
CREATE INDEX idx_mv_dashboard_open_findings_tenant ON mv_dashboard_open_findings(tenant_id);

-- Grant permissions (optional, adjust as needed for your setup)
-- GRANT SELECT ON ALL MATERIALIZED VIEWS IN SCHEMA public TO oem_reader;
