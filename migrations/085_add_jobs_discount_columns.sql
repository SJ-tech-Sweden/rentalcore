-- 085_add_jobs_discount_columns.sql
-- Ensure `jobs.discount` and `jobs.discount_type` exist for job creation flow.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'jobs') THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'jobs' AND column_name = 'discount'
        ) THEN
            ALTER TABLE public.jobs ADD COLUMN discount numeric(12,2) NOT NULL DEFAULT 0;
        END IF;

        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'jobs' AND column_name = 'discount_type'
        ) THEN
            ALTER TABLE public.jobs ADD COLUMN discount_type varchar(20) NOT NULL DEFAULT 'amount';
        END IF;
    END IF;
END$$;
