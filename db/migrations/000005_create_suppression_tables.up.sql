-- +migrate Up
-- Create suppressions, suppression_reviews, and tenant_policy_state tables

CREATE TABLE suppressions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    state VARCHAR(50) NOT NULL,
    scope_type VARCHAR(50) NOT NULL,
    scope_ref VARCHAR(255),
    match_type VARCHAR(50) NOT NULL,
    match_value TEXT NOT NULL,
    goal_status VARCHAR(50) NOT NULL,
    reason TEXT NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id),
    approved_by BIGINT REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_state CHECK (state IN ('pending', 'approved', 'rejected', 'expired', 'revoked')),
    CONSTRAINT valid_match_type CHECK (match_type IN ('cve'))
);

CREATE TABLE suppression_reviews (
    id BIGSERIAL PRIMARY KEY,
    suppression_id BIGINT NOT NULL REFERENCES suppressions(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    actor BIGINT NOT NULL REFERENCES users(id),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    comment TEXT,
    CONSTRAINT valid_action CHECK (action IN ('proposed', 'approved', 'rejected', 'revoked'))
);

CREATE TABLE tenant_policy_state (
    tenant_id BIGINT PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    policy_revision BIGINT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for suppressions
CREATE INDEX idx_suppressions_tenant_id ON suppressions(tenant_id);
CREATE INDEX idx_suppressions_state ON suppressions(tenant_id, state);
CREATE INDEX idx_suppressions_match_type ON suppressions(tenant_id, match_type);
CREATE INDEX idx_suppressions_active ON suppressions(tenant_id, state) WHERE state IN ('approved', 'pending');

-- Index for suppression_reviews
CREATE INDEX idx_suppression_reviews_suppression_id ON suppression_reviews(suppression_id);
CREATE INDEX idx_suppression_reviews_timestamp ON suppression_reviews(timestamp DESC);

-- Triggers for updated_at
CREATE TRIGGER update_suppressions_updated_at BEFORE UPDATE ON suppressions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenant_policy_state_updated_at BEFORE UPDATE ON tenant_policy_state
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
