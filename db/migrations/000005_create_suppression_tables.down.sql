-- +migrate Down
-- Drop suppression tables

DROP TRIGGER IF EXISTS update_tenant_policy_state_updated_at ON tenant_policy_state;
DROP TRIGGER IF EXISTS update_suppressions_updated_at ON suppressions;
DROP INDEX IF EXISTS idx_suppression_reviews_timestamp;
DROP INDEX IF EXISTS idx_suppression_reviews_suppression_id;
DROP INDEX IF EXISTS idx_suppressions_active;
DROP INDEX IF EXISTS idx_suppressions_match_type;
DROP INDEX IF EXISTS idx_suppressions_state;
DROP INDEX IF EXISTS idx_suppressions_tenant_id;
DROP TABLE IF EXISTS suppression_reviews;
DROP TABLE IF EXISTS tenant_policy_state;
DROP TABLE IF EXISTS suppressions;
