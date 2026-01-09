-- +migrate Down
-- Drop tenants table

DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_tenants_name;
DROP TABLE IF EXISTS tenants;
