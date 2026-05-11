CREATE TABLE IF NOT EXISTS job_history (
        history_id BIGSERIAL PRIMARY KEY,
        job_id INT NOT NULL,
        user_id INT NULL,
        changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        change_type VARCHAR(50) NOT NULL,
        field_name VARCHAR(100) NULL,
        old_value TEXT NULL,
        new_value TEXT NULL,
        description TEXT NULL,
        ip_address VARCHAR(45) NULL,
        user_agent VARCHAR(255) NULL
);
CREATE INDEX IF NOT EXISTS idx_job_id ON job_history(job_id);
CREATE INDEX IF NOT EXISTS idx_user_id ON job_history(user_id);
CREATE INDEX IF NOT EXISTS idx_changed_at ON job_history(changed_at);
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_job_history_type') THEN
        ALTER TABLE job_history
            ADD CONSTRAINT chk_job_history_type CHECK (change_type IN ('created','updated','status_changed','device_added','device_removed','deleted'));
    END IF;

    IF to_regclass('public.jobs') IS NOT NULL
       AND NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_job_history_job') THEN
        ALTER TABLE job_history
            ADD CONSTRAINT fk_job_history_job FOREIGN KEY (job_id) REFERENCES jobs(jobid) ON DELETE CASCADE;
    END IF;

    IF to_regclass('public.users') IS NOT NULL
       AND NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_job_history_user') THEN
        ALTER TABLE job_history
            ADD CONSTRAINT fk_job_history_user FOREIGN KEY (user_id) REFERENCES users(userid) ON DELETE SET NULL;
    END IF;
END$$;

-- Add created_by and updated_by fields to jobs table
ALTER TABLE jobs
    ADD COLUMN IF NOT EXISTS created_by INT NULL,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_by INT NULL,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP;

-- Add foreign keys for created_by and updated_by
DO $$
BEGIN
    IF to_regclass('jobs') IS NOT NULL THEN
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_jobs_created_by') THEN
            ALTER TABLE jobs ADD CONSTRAINT fk_jobs_created_by FOREIGN KEY (created_by) REFERENCES users(userid) ON DELETE SET NULL;
        ELSE
            RAISE NOTICE 'Constraint fk_jobs_created_by already exists, skipping';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_jobs_updated_by') THEN
            ALTER TABLE jobs ADD CONSTRAINT fk_jobs_updated_by FOREIGN KEY (updated_by) REFERENCES users(userid) ON DELETE SET NULL;
        ELSE
            RAISE NOTICE 'Constraint fk_jobs_updated_by already exists, skipping';
        END IF;
    ELSE
        RAISE NOTICE 'jobs relation not found; skipping FK creation for migration 028';
    END IF;
END$$;

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_created_by ON jobs(created_by);
CREATE INDEX IF NOT EXISTS idx_updated_by ON jobs(updated_by);
CREATE INDEX IF NOT EXISTS idx_created_at ON jobs(created_at);
CREATE INDEX IF NOT EXISTS idx_updated_at ON jobs(updated_at);
