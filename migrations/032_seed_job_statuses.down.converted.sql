-- AUTO-CONVERTED (heuristic)
-- Source: 032_seed_job_statuses.down.sql.bak
-- Review this file for correctness before applying to Postgres.

DO $$
BEGIN
	IF to_regclass('status') IS NOT NULL THEN
		DELETE FROM status WHERE statusid IN (1, 2, 3, 4, 5);
		RAISE NOTICE 'Deleted seeded job statuses from status table';
	ELSE
		RAISE NOTICE 'status relation not found; skipping deletion in migration 032';
	END IF;
END$$;
