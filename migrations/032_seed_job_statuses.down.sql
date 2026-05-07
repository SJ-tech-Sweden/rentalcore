-- Migration 032 down (noop)
DO $$
BEGIN
	RAISE NOTICE 'Skipping MySQL-original migration 032_seed_job_statuses.down.sql; converted migration should be used instead.';
END $$;
