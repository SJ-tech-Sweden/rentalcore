-- Migration 042 rollback: remove cable_snapshot column from job_cables
--
-- NOTE: run this only after disabling the CABLE_SNAPSHOT_ENABLED feature flag
--       so that in-flight requests do not try to read a column that no longer
--       exists.  The original cross-service FK to cables("cableID") is kept
--       intact by this migration; it is only removed in a future PR.

-- Migration 042 down (noop)
DO $$
BEGIN
    RAISE NOTICE 'Skipping MySQL-original migration 042_add_cable_snapshot.down.sql; converted migration should be used instead.';
END $$;
