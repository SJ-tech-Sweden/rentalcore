-- AUTO-CONVERTED (heuristic)
-- Source: 028_add_job_history.down.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Remove indexes from jobs table
-- Remove indexes from jobs table (postgreSQL uses standalone DROP INDEX)
DROP INDEX IF EXISTS idx_created_by;
DROP INDEX IF EXISTS idx_updated_by;
DROP INDEX IF EXISTS idx_created_at;
DROP INDEX IF EXISTS idx_updated_at;

-- Remove foreign keys and columns from jobs table only if the table exists
DO $$
DECLARE
    col_has_dep boolean;
BEGIN
    IF to_regclass('jobs') IS NOT NULL THEN
        ALTER TABLE jobs DROP CONSTRAINT IF EXISTS fk_jobs_created_by;
        ALTER TABLE jobs DROP CONSTRAINT IF EXISTS fk_jobs_updated_by;

        -- Drop columns only when there are no dependent objects (views, constraints, etc.)
        -- created_by
        SELECT EXISTS(
            SELECT 1 FROM pg_catalog.pg_depend d
            JOIN pg_catalog.pg_attribute a ON d.refobjid = a.attrelid AND d.refobjsubid = a.attnum
            WHERE a.attrelid = 'jobs'::regclass AND a.attname = 'created_by'
        ) INTO col_has_dep;
        IF NOT col_has_dep THEN
            ALTER TABLE jobs DROP COLUMN IF EXISTS created_by;
        ELSE
            RAISE NOTICE 'Skipping drop of jobs.created_by; dependent objects exist';
        END IF;

        -- created_at
        SELECT EXISTS(
            SELECT 1 FROM pg_catalog.pg_depend d
            JOIN pg_catalog.pg_attribute a ON d.refobjid = a.attrelid AND d.refobjsubid = a.attnum
            WHERE a.attrelid = 'jobs'::regclass AND a.attname = 'created_at'
        ) INTO col_has_dep;
        IF NOT col_has_dep THEN
            ALTER TABLE jobs DROP COLUMN IF EXISTS created_at;
        ELSE
            RAISE NOTICE 'Skipping drop of jobs.created_at; dependent objects exist';
        END IF;

        -- updated_by
        SELECT EXISTS(
            SELECT 1 FROM pg_catalog.pg_depend d
            JOIN pg_catalog.pg_attribute a ON d.refobjid = a.attrelid AND d.refobjsubid = a.attnum
            WHERE a.attrelid = 'jobs'::regclass AND a.attname = 'updated_by'
        ) INTO col_has_dep;
        IF NOT col_has_dep THEN
            ALTER TABLE jobs DROP COLUMN IF EXISTS updated_by;
        ELSE
            RAISE NOTICE 'Skipping drop of jobs.updated_by; dependent objects exist';
        END IF;

        -- updated_at
        SELECT EXISTS(
            SELECT 1 FROM pg_catalog.pg_depend d
            JOIN pg_catalog.pg_attribute a ON d.refobjid = a.attrelid AND d.refobjsubid = a.attnum
            WHERE a.attrelid = 'jobs'::regclass AND a.attname = 'updated_at'
        ) INTO col_has_dep;
        IF NOT col_has_dep THEN
            ALTER TABLE jobs DROP COLUMN IF EXISTS updated_at;
        ELSE
            RAISE NOTICE 'Skipping drop of jobs.updated_at; dependent objects exist';
        END IF;

    ELSE
        RAISE NOTICE 'jobs relation not found; skipping FK/column drops for migration 028';
    END IF;
END $$;

-- Drop job_history table if it exists
DROP TABLE IF EXISTS job_history;
