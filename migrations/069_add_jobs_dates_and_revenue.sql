-- Migration 069: Add date and revenue columns to `jobs` (compatibility)
-- Purpose: Some deployments expect `startdate`, `enddate`, `revenue`, and
-- `final_revenue` columns on the `jobs` table. Add them idempotently.

DO $$
BEGIN
    IF to_regclass('jobs') IS NULL THEN
        RAISE NOTICE 'jobs relation not found; skipping migration 069_add_jobs_dates_and_revenue';
        RETURN;
    END IF;

    -- startdate
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'jobs' AND column_name = 'startdate'
    ) THEN
        ALTER TABLE jobs ADD COLUMN startdate DATE;
        RAISE NOTICE 'Added column jobs.startdate';
    END IF;

    -- enddate
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'jobs' AND column_name = 'enddate'
    ) THEN
        ALTER TABLE jobs ADD COLUMN enddate DATE;
        RAISE NOTICE 'Added column jobs.enddate';
    END IF;

    -- revenue
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'jobs' AND column_name = 'revenue'
    ) THEN
        ALTER TABLE jobs ADD COLUMN revenue numeric(12,2) DEFAULT 0;
        RAISE NOTICE 'Added column jobs.revenue';
    END IF;

    -- final_revenue
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'jobs' AND column_name = 'final_revenue'
    ) THEN
        ALTER TABLE jobs ADD COLUMN final_revenue numeric(12,2) DEFAULT 0;
        RAISE NOTICE 'Added column jobs.final_revenue';
    END IF;

END
$$;
