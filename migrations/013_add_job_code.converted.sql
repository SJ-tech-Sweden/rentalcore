-- AUTO-CONVERTED (heuristic)
-- Source: 013_add_job_code.sql.bak
-- Review this file for correctness before applying to Postgres.

DO $$
DECLARE
  job_id_col text := NULL;
  constraint_exists boolean := false;
BEGIN
  IF to_regclass('public.jobs') IS NULL THEN
    RAISE NOTICE 'jobs relation not found; skipping job_code migration';
    RETURN;
  END IF;

  -- detect job id column name
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='jobID') THEN
    job_id_col := 'jobID';
  ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='job_id') THEN
    job_id_col := 'job_id';
  ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='id') THEN
    job_id_col := 'id';
  END IF;

  -- add column if missing
  EXECUTE 'ALTER TABLE jobs ADD COLUMN IF NOT EXISTS job_code VARCHAR(16) NULL';

  IF job_id_col IS NOT NULL THEN
    -- populate job_code using PostgreSQL casting; LPAD operates on text
    EXECUTE format('UPDATE jobs SET job_code = %L || LPAD(%I::text,6,%L) WHERE job_code IS NULL OR job_code = %L', 'JOB', job_id_col, '0', '');
  ELSE
    RAISE NOTICE 'Could not find job id column (jobID/job_id/id); skipping population of job_code';
  END IF;

  -- set NOT NULL if possible
  EXECUTE 'ALTER TABLE jobs ALTER COLUMN job_code SET NOT NULL';

  -- add unique constraint if not present
  SELECT EXISTS(SELECT 1 FROM pg_constraint WHERE conname = 'ux_jobs_job_code') INTO constraint_exists;
  IF NOT constraint_exists THEN
    EXECUTE 'ALTER TABLE jobs ADD CONSTRAINT ux_jobs_job_code UNIQUE (job_code)';
  END IF;
END$$;
