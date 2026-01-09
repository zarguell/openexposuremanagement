-- +migrate Up
-- Create intel_cve and intel_sync_runs tables

CREATE TABLE intel_cve (
    cve VARCHAR(20) PRIMARY KEY,
    -- NVD fields
    description TEXT,
    cvss_score NUMERIC(3, 1),
    cvss_vector VARCHAR(200),
    -- EPSS fields
    epss_score NUMERIC(5, 4),
    epss_percentile NUMERIC(5, 2),
    -- CISA KEV fields
    is_kev BOOLEAN NOT NULL DEFAULT FALSE,
    kev_date_added DATE,
    kev_due_date DATE,
    -- Metadata
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE intel_sync_runs (
    id BIGSERIAL PRIMARY KEY,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL,
    error_text TEXT,
    source VARCHAR(100) NOT NULL,
    CONSTRAINT valid_sync_status CHECK (status IN ('running', 'completed', 'failed'))
);

-- Indexes for intel_cve
CREATE INDEX idx_intel_cve_kev ON intel_cve(is_kev) WHERE is_kev = TRUE;
CREATE INDEX idx_intel_cve_cvss_score ON intel_cve(cvss_score DESC);
CREATE INDEX idx_intel_cve_epss_score ON intel_cve(epss_score DESC);
CREATE INDEX idx_intel_cve_epss_percentile ON intel_cve(epss_percentile DESC);

-- Indexes for intel_sync_runs
CREATE INDEX idx_intel_sync_runs_started_at ON intel_sync_runs(started_at DESC);
CREATE INDEX idx_intel_sync_runs_source ON intel_sync_runs(source);

-- Trigger for updated_at on intel_cve
CREATE TRIGGER update_intel_cve_updated_at BEFORE UPDATE ON intel_cve
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
