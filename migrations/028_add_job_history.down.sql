-- Migration 028 down (noop)
DO $$
BEGIN
    RAISE NOTICE 'Skipping MySQL-original migration 028_add_job_history.down.sql; converted migration should be used instead.';
END $$;
