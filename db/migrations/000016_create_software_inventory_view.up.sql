-- +migrate Up
-- Create software_inventory view for query framework
-- This view joins asset_software with software and assets to support querying
-- which assets have which software installed

CREATE OR REPLACE VIEW software_inventory AS
SELECT
    asw.id,
    asw.tenant_id,
    asw.asset_id,
    asw.software_id,
    asw.source,
    asw.install_path,
    asw.first_seen_at,
    asw.last_seen_at,
    asw.created_at,
    asw.updated_at,
    -- From software table
    s.cpe_string,
    s.vendor,
    s.product_name,
    s.version,
    s.edition,
    s.target_hw,
    s.lang,
    s.title_formatted,
    -- From assets table
    a.canonical_name AS asset_name,
    a.is_active AS asset_is_active
FROM asset_software asw
JOIN software s ON s.id = asw.software_id
JOIN assets a ON a.id = asw.asset_id;

-- Create index on tenant_id for performance (view queries benefit from underlying indexes)
-- Note: The underlying tables already have indexes on:
-- - asset_software(tenant_id, asset_id, software_id)
-- - software(cpe_string)
-- - assets(id)

-- Comment on view for documentation
COMMENT ON VIEW software_inventory IS 'Software inventory view joining asset_software, software, and assets for query framework';
