-- Migration 068: Add `description` column to `jobs` (compatibility)
-- Purpose: Some deployments have queries that reference `jobs.description`.
-- This migration adds the column idempotently if it does not exist.

DO $$
BEGIN
    IF to_regclass('jobs') IS NULL THEN
        RAISE NOTICE 'jobs relation not found; skipping migration 068_add_jobs_description_column';
        RETURN;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'jobs' AND column_name = 'description'
    ) THEN
        ALTER TABLE jobs ADD COLUMN description TEXT;
        RAISE NOTICE 'Added column jobs.description';
    ELSE
        RAISE NOTICE 'jobs.description already exists; skipping';
    END IF;
END
$$;
