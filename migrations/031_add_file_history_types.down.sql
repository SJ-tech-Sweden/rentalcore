-- Rollback file_added and file_removed from job_history change_type enum
-- This reverts the enum to its original values

-- First, update any existing file_added/file_removed entries to 'updated'
UPDATE job_history SET change_type = 'updated' WHERE change_type IN ('file_added', 'file_removed');

-- For MySQL, revert the ENUM:
-- Migration 031 down (noop)
DO $$
BEGIN
    RAISE NOTICE 'Skipping MySQL-original migration 031_add_file_history_types.down.sql; converted migration applied instead.';
END $$;
