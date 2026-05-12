-- 054_backfill_denorm_from_api.sql
-- Template for backfilling denormalized columns from WarehouseCore API
-- This is a helper template. Prefer running the Go backfill tool included at /tools/backfill

-- Example SQL to update jobs using an external tool:
-- UPDATE jobs SET cable_snapshot = $1, updated_at = NOW() WHERE jobid = $2;

-- The actual backfill operation is implemented by the tools/backfill Go program which
-- fetches authoritative data from WarehouseCore and updates denormalized columns.
