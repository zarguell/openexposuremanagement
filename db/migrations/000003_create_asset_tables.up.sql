-- +migrate Up
-- Create assets and asset_identifiers tables

CREATE TABLE assets (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    canonical_name VARCHAR(512) NOT NULL,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    owner_team_id BIGINT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, canonical_name)
);

CREATE TABLE asset_identifiers (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    id_type VARCHAR(50) NOT NULL,
    id_value VARCHAR(512) NOT NULL,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    source VARCHAR(100) NOT NULL,
    CONSTRAINT asset_identifiers_unique UNIQUE(tenant_id, id_type, id_value, asset_id)
);

-- Indexes to support matching algorithm
CREATE INDEX idx_assets_tenant_id ON assets(tenant_id);
CREATE INDEX idx_assets_last_seen ON assets(tenant_id, last_seen_at DESC);
CREATE INDEX idx_assets_canonical_name ON assets(tenant_id, canonical_name);

-- Composite index for identifier lookups (critical for matching)
CREATE INDEX idx_asset_identifiers_lookup ON asset_identifiers(tenant_id, id_type, id_value);
CREATE INDEX idx_asset_identifiers_asset_id ON asset_identifiers(asset_id);
CREATE INDEX idx_asset_identifiers_recency ON asset_identifiers(tenant_id, last_seen_at DESC);

-- Partial index for active assets
CREATE INDEX idx_assets_active ON assets(tenant_id, is_active) WHERE is_active = TRUE;

-- Triggers for updated_at
CREATE TRIGGER update_assets_updated_at BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
