-- Add file_added and file_removed to job_history change_type enum
-- This allows tracking who uploaded/removed files from jobs

-- For SQLite (which the app uses at runtime), ENUM is stored as TEXT with CHECK
-- For PostgreSQL, we need to modify the enum type

-- SQLite compatible version (TEXT column, no enum modification needed)
-- The application model already accepts the new values

DO $$
BEGIN
    -- Ensure the change_type column exists and add a CHECK constraint with the new values
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_history' AND column_name = 'change_type') THEN
        IF NOT EXISTS (
            SELECT 1 FROM pg_constraint c
            JOIN pg_class t ON c.conrelid = t.oid
            WHERE t.relname = 'job_history' AND c.conname = 'chk_job_history_change_type'
        ) THEN
            EXECUTE 'ALTER TABLE job_history ADD CONSTRAINT chk_job_history_change_type CHECK (change_type IN (''created'',''updated'',''status_changed'',''device_added'',''device_removed'',''deleted'',''file_added'',''file_removed''))';
        END IF;
    END IF;
END$$;
