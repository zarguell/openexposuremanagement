-- +migrate Up
-- Create software catalog and asset_software junction tables

-- Software catalog: stores each unique software product once
CREATE TABLE software (
    id BIGSERIAL PRIMARY KEY,
    cpe_string VARCHAR(500) UNIQUE NOT NULL,
    vendor VARCHAR(255) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    version VARCHAR(255),
    edition VARCHAR(255),
    target_hw VARCHAR(100),
    lang VARCHAR(50),
    title_formatted VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Junction table: current software on each asset
CREATE TABLE asset_software (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    software_id BIGINT NOT NULL REFERENCES software(id) ON DELETE CASCADE,
    source VARCHAR(100) NOT NULL,
    install_path VARCHAR(1024),
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, asset_id, software_id)
);

-- Indexes for software table
CREATE INDEX idx_software_vendor_product ON software(vendor, product_name);
CREATE INDEX idx_software_cpe ON software(cpe_string);
CREATE INDEX idx_software_product_version ON software(product_name, version);

-- Indexes for asset_software table
-- For "assets with specific software" queries
CREATE INDEX idx_asset_software_tenant_asset ON asset_software(tenant_id, asset_id);
CREATE INDEX idx_asset_software_tenant_software ON asset_software(tenant_id, software_id);
CREATE INDEX idx_asset_software_asset_lookup ON asset_software(asset_id);

-- Composite for complex queries (e.g., "software X on assets with findings Y")
CREATE INDEX idx_asset_software_tenant_vendor ON asset_software(tenant_id, software_id, asset_id);

-- Recency index for cleanup queries
CREATE INDEX idx_asset_software_recency ON asset_software(tenant_id, last_seen_at DESC);

-- Materialized view for software+asset summary queries
CREATE MATERIALIZED VIEW mv_asset_software_summary AS
SELECT
    a.id AS asset_id,
    a.tenant_id,
    a.canonical_name,
    s.id AS software_id,
    s.cpe_string,
    s.vendor,
    s.product_name,
    s.version,
    s.edition,
    s.title_formatted,
    asoft.first_seen_at,
    asoft.last_seen_at,
    asoft.install_path
FROM assets a
JOIN asset_software asoft ON asoft.asset_id = a.id
JOIN software s ON s.id = asoft.software_id
WHERE a.is_active = TRUE;

-- Unique index required for REFRESH MATERIALIZED VIEW CONCURRENTLY
CREATE UNIQUE INDEX idx_mv_asset_software_summary_unique
    ON mv_asset_software_summary(asset_id, software_id);

-- Indexes for materialized view queries
CREATE INDEX idx_mv_asset_software_summary_tenant
    ON mv_asset_software_summary(tenant_id);
CREATE INDEX idx_mv_asset_software_summary_software
    ON mv_asset_software_summary(software_id);
CREATE INDEX idx_mv_asset_software_summary_vendor_product
    ON mv_asset_software_summary(vendor, product_name);

-- Triggers for updated_at
CREATE TRIGGER update_software_updated_at BEFORE UPDATE ON software
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_asset_software_updated_at BEFORE UPDATE ON asset_software
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
