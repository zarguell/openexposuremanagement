-- +migrate Down
-- Drop software catalog and asset_software tables

DROP MATERIALIZED VIEW IF EXISTS mv_asset_software_summary;

DROP TABLE IF EXISTS asset_software;
DROP TABLE IF EXISTS software;
