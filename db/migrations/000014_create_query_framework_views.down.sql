-- +migrate Down
-- Drop query framework views

DROP VIEW IF EXISTS findings CASCADE;
DROP VIEW IF EXISTS assets_extended CASCADE;
