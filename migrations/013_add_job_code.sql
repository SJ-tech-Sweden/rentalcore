ALTER TABLE jobs
  ADD COLUMN IF NOT EXISTS job_code VARCHAR(16);

UPDATE jobs
SET job_code = 'JOB' || LPAD(jobid::text, 6, '0')
WHERE job_code IS NULL OR job_code = '';

ALTER TABLE jobs
  ALTER COLUMN job_code SET NOT NULL;
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON c.conrelid = t.oid
        WHERE c.conname = 'ux_jobs_job_code' AND t.relname = 'jobs'
    ) THEN
    EXECUTE 'ALTER TABLE jobs ADD CONSTRAINT ux_jobs_job_code UNIQUE (job_code)';
    END IF;
END$$;
