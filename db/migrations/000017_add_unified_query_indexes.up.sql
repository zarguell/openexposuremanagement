-- +migrate Up
-- Add indexes for unified query JOIN performance

-- Composite index for assets LEFT JOIN software_inventory
-- Covers: SELECT * FROM assets LEFT JOIN asset_software ON assets.id = asset_software.asset_id
-- Supports queries filtering by tenant and asset, with software lookup
CREATE INDEX idx_asset_software_join_covering ON asset_software(tenant_id, asset_id, software_id)
INCLUDE (last_seen_at);

-- Composite index for assets LEFT JOIN findings
-- Covers: SELECT * FROM assets LEFT JOIN finding_instances ON assets.id = finding_instances.asset_id
-- Supports queries filtering by tenant and asset, with finding status and definition
CREATE INDEX idx_finding_instances_asset_join ON finding_instances(tenant_id, asset_id, definition_uid)
INCLUDE (effective_status, last_observed_at);

-- Index for software vendor/product queries (common filter)
-- Supports queries filtering by vendor, product, and version
CREATE INDEX idx_software_vendor_product_lookup ON software(vendor, product_name, version);

-- Index for finding definition lookups by source
-- Supports queries filtering findings by scanner source
CREATE INDEX idx_finding_definitions_source_lookup ON finding_definitions(source, source_definition_id);
