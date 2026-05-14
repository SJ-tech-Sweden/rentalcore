-- AUTO-CONVERTED (heuristic)
-- Source: 032_seed_job_statuses.up.sql.bak
-- Review this file for correctness before applying to Postgres.

DO $$
BEGIN
	IF to_regclass('status') IS NULL THEN
		RAISE NOTICE 'status relation not found; skipping migration 032_seed_job_statuses.up';
	ELSE
		INSERT INTO status (statusid, status) VALUES 
		(1, 'Draft'),
		(2, 'Confirmed'),
		(3, 'Active'),
		(4, 'Completed'),
		(5, 'Cancelled')
		ON CONFLICT (statusid) DO NOTHING;

		-- Update sequence to ensure next insert works
		PERFORM setval('status_statusid_seq', (SELECT COALESCE(MAX(statusid),0) FROM status));
		RAISE NOTICE 'Applied migration 032_seed_job_statuses.up';
	END IF;
END$$;
