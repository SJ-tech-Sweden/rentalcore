-- AUTO-CONVERTED (heuristic)
-- Source: 042_add_cable_snapshot.down.sql.bak
-- Review this file for correctness before applying to Postgres.

DO $$
BEGIN
    IF to_regclass('job_cables') IS NULL THEN
        RAISE NOTICE 'job_cables relation not found; skipping down migration 042_add_cable_snapshot';
    ELSE
        DROP INDEX IF EXISTS idx_job_cables_snapshot_backfill;

        ALTER TABLE job_cables
            DROP COLUMN IF EXISTS cable_snapshot;

        RAISE NOTICE 'Applied down migration 042_add_cable_snapshot';
    END IF;
END$$;
