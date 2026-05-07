-- AUTO-CONVERTED (heuristic)
-- Source: 031_add_file_history_types.down.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Rollback file_added and file_removed from job_history change_type enum
-- This reverts the enum to its original values

-- First, update any existing file_added/file_removed entries to 'updated'
UPDATE job_history SET change_type = 'updated' WHERE change_type IN ('file_added', 'file_removed');

-- For MySQL, revert the ENUM:
DO $$
BEGIN
    IF to_regclass('job_history') IS NOT NULL THEN
        -- Remove the CHECK constraint added by the converted up-migration
        ALTER TABLE job_history DROP CONSTRAINT IF EXISTS chk_job_history_type;

        -- Set column type back to plain varchar and enforce NOT NULL
        ALTER TABLE job_history ALTER COLUMN change_type TYPE varchar(50);
        ALTER TABLE job_history ALTER COLUMN change_type SET NOT NULL;
        RAISE NOTICE 'Reverted job_history.change_type to varchar(50) and set NOT NULL';
    ELSE
        RAISE NOTICE 'job_history relation not found; skipping down migration 031';
    END IF;
END$$;
