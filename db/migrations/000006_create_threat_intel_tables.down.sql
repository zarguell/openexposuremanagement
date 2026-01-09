-- +migrate Down
-- Drop threat intel tables

DROP TRIGGER IF EXISTS update_intel_cve_updated_at ON intel_cve;
DROP INDEX IF EXISTS idx_intel_sync_runs_source;
DROP INDEX IF EXISTS idx_intel_sync_runs_started_at;
DROP INDEX IF EXISTS idx_intel_cve_epss_percentile;
DROP INDEX IF EXISTS idx_intel_cve_epss_score;
DROP INDEX IF EXISTS idx_intel_cve_kev;
DROP TABLE IF EXISTS intel_sync_runs;
DROP TABLE IF EXISTS intel_cve;
