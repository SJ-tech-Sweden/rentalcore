-- AUTO-CONVERTED (heuristic)
-- Source: 041_add_job_cables.up.sql.bak
-- Review this file for correctness before applying to Postgres.

DO $$
BEGIN
    IF to_regclass('cables') IS NULL THEN
        RAISE NOTICE 'cables relation not found; skipping migration 041_add_job_cables.up';
    ELSE
        CREATE TABLE IF NOT EXISTS job_cables (
            jobid     INTEGER NOT NULL,
            "cableID" INTEGER NOT NULL,
            PRIMARY KEY (jobid, "cableID"),
            FOREIGN KEY (jobid) REFERENCES jobs(jobid) ON DELETE CASCADE,
            FOREIGN KEY ("cableID") REFERENCES cables("cableID") ON DELETE CASCADE
        );
        RAISE NOTICE 'Applied migration 041_add_job_cables.up';
    END IF;
END$$;
