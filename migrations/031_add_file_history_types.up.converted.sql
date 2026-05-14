-- AUTO-CONVERTED (heuristic)
-- Source: 031_add_file_history_types.up.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Add file_added and file_removed to job_history change_type enum
-- This allows tracking who uploaded/removed files from jobs

-- For SQLite (which the app uses at runtime), ENUM is stored as TEXT with CHECK
-- For PostgreSQL, we need to modify the enum type

-- SQLite compatible version (TEXT column, no enum modification needed)
-- The application model already accepts the new values

-- For MySQL, modify the ENUM:
-- Note: This is a no-op if using SQLite since SQLite doesn't have ENUM
DO $$
BEGIN
    IF to_regclass('job_history') IS NOT NULL THEN
        -- Replace or add a CHECK constraint that allows the new file_added/file_removed types
        ALTER TABLE job_history DROP CONSTRAINT IF EXISTS chk_job_history_type;
        ALTER TABLE job_history ADD CONSTRAINT chk_job_history_type CHECK (
            change_type IN (
                'created', 'updated', 'status_changed', 'device_added', 'device_removed', 'deleted', 'file_added', 'file_removed'
            )
        );
        RAISE NOTICE 'Applied migration 031_add_file_history_types.up: updated CHECK constraint on job_history.change_type';
    ELSE
        RAISE NOTICE 'job_history relation not found; skipping migration 031_add_file_history_types.up';
    END IF;
END$$;
