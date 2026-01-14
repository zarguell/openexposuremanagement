-- Saved queries table for dashboard widgets and future user-saved queries
CREATE TABLE saved_queries (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('assets', 'findings')),
    query_json JSONB NOT NULL,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by BIGINT REFERENCES users(id),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

-- Index for tenant lookups
CREATE INDEX idx_saved_queries_tenant ON saved_queries(tenant_id);

-- Index for system queries (used by dashboard)
CREATE INDEX idx_saved_queries_system ON saved_queries(is_system) WHERE is_system = true;

-- Index for entity type lookups
CREATE INDEX idx_saved_queries_entity_type ON saved_queries(tenant_id, entity_type);
