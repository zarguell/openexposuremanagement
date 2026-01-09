-- +migrate Down
-- Drop asset tables

DROP TRIGGER IF EXISTS update_assets_updated_at ON assets;
DROP INDEX IF EXISTS idx_assets_active;
DROP INDEX IF EXISTS idx_asset_identifiers_recency;
DROP INDEX IF EXISTS idx_asset_identifiers_asset_id;
DROP INDEX IF EXISTS idx_asset_identifiers_lookup;
DROP INDEX IF EXISTS idx_assets_canonical_name;
DROP INDEX IF EXISTS idx_assets_last_seen;
DROP INDEX IF EXISTS idx_assets_tenant_id;
DROP TABLE IF EXISTS asset_identifiers;
DROP TABLE IF EXISTS assets;
