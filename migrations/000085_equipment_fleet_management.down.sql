DROP TRIGGER IF EXISTS update_maintenance_logs_modtime ON maintenance_logs;
DROP TABLE IF EXISTS maintenance_logs;
DROP TRIGGER IF EXISTS update_equipment_allocations_modtime ON equipment_allocations;
DROP TABLE IF EXISTS equipment_allocations;
DROP TRIGGER IF EXISTS update_fleet_assets_modtime ON fleet_assets;
DROP TABLE IF EXISTS fleet_assets;
-- Note: btree_gist extension left in place as other schemas may depend on it
