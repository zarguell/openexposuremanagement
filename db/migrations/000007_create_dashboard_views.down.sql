-- +migrate Down
-- Drop dashboard materialized views

DROP MATERIALIZED VIEW IF EXISTS mv_dashboard_open_findings CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_dashboard_assets CASCADE;
DROP MATERIALIZED VIEW IF EXISTS mv_dashboard_counts CASCADE;
