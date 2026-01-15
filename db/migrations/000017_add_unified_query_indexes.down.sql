-- +migrate Down
-- Remove indexes for unified query JOIN performance

DROP INDEX IF EXISTS idx_asset_software_join_covering;
DROP INDEX IF EXISTS idx_finding_instances_asset_join;
DROP INDEX IF EXISTS idx_software_vendor_product_lookup;
DROP INDEX IF EXISTS idx_finding_definitions_source_lookup;
