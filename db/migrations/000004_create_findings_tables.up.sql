-- +migrate Up
-- Create finding_definitions, finding_definition_aliases, and finding_instances tables

CREATE TABLE finding_definitions (
    definition_uid VARCHAR(255) PRIMARY KEY,
    source VARCHAR(100) NOT NULL,
    source_definition_id VARCHAR(255) NOT NULL,
    title TEXT NOT NULL,
    severity_default VARCHAR(50) NOT NULL,
    references_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source, source_definition_id)
);

CREATE TABLE finding_definition_aliases (
    id BIGSERIAL PRIMARY KEY,
    definition_uid VARCHAR(255) NOT NULL REFERENCES finding_definitions(definition_uid) ON DELETE CASCADE,
    alias_type VARCHAR(50) NOT NULL,
    alias_value VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(definition_uid, alias_type, alias_value)
);

CREATE TABLE finding_instances (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    asset_id BIGINT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    definition_uid VARCHAR(255) NOT NULL REFERENCES finding_definitions(definition_uid) ON DELETE CASCADE,
    scanner_status VARCHAR(50) NOT NULL,
    first_observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    evidence_json JSONB,
    effective_status VARCHAR(50) NOT NULL,
    effective_reason TEXT,
    effective_revision BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, asset_id, definition_uid)
);

-- Indexes for finding_definitions
CREATE INDEX idx_finding_definitions_source ON finding_definitions(source);

-- Indexes for finding_definition_aliases (CVE lookups)
CREATE INDEX idx_finding_definition_aliases_type_value ON finding_definition_aliases(alias_type, alias_value);
CREATE INDEX idx_finding_definition_aliases_definition_uid ON finding_definition_aliases(definition_uid);

-- Critical indexes for finding_instances
CREATE INDEX idx_finding_instances_tenant_id ON finding_instances(tenant_id);
CREATE INDEX idx_finding_instances_asset_id ON finding_instances(asset_id);
CREATE INDEX idx_finding_instances_definition_uid ON finding_instances(definition_uid);

-- Composite index for tenant filters (effective_status, definition_uid, last_observed_at)
CREATE INDEX idx_finding_instances_tenant_status ON finding_instances(tenant_id, effective_status);
CREATE INDEX idx_finding_instances_tenant_definition ON finding_instances(tenant_id, definition_uid);
CREATE INDEX idx_finding_instances_tenant_last_observed ON finding_instances(tenant_id, last_observed_at DESC);

-- Index for effective recompute queries
CREATE INDEX idx_finding_instances_effective_revision ON finding_instances(effective_revision);

-- GIN index for JSONB queries
CREATE INDEX idx_finding_instances_evidence ON finding_instances USING GIN (evidence_json);

-- Triggers for updated_at
CREATE TRIGGER update_finding_definitions_updated_at BEFORE UPDATE ON finding_definitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_finding_instances_updated_at BEFORE UPDATE ON finding_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
