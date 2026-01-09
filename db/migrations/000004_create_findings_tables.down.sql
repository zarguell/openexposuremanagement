-- +migrate Down
-- Drop findings tables

DROP TRIGGER IF EXISTS update_finding_instances_updated_at ON finding_instances;
DROP TRIGGER IF EXISTS update_finding_definitions_updated_at ON finding_definitions;
DROP INDEX IF EXISTS idx_finding_instances_evidence;
DROP INDEX IF EXISTS idx_finding_instances_effective_revision;
DROP INDEX IF EXISTS idx_finding_instances_tenant_last_observed;
DROP INDEX IF EXISTS idx_finding_instances_tenant_definition;
DROP INDEX IF EXISTS idx_finding_instances_tenant_status;
DROP INDEX IF EXISTS idx_finding_instances_definition_uid;
DROP INDEX IF EXISTS idx_finding_instances_asset_id;
DROP INDEX IF EXISTS idx_finding_instances_tenant_id;
DROP INDEX IF EXISTS idx_finding_definition_aliases_definition_uid;
DROP INDEX IF EXISTS idx_finding_definition_aliases_type_value;
DROP INDEX IF EXISTS idx_finding_definitions_source;
DROP TABLE IF EXISTS finding_instances;
DROP TABLE IF EXISTS finding_definition_aliases;
DROP TABLE IF EXISTS finding_definitions;
