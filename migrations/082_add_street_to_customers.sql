-- 082_add_street_to_customers.sql
-- Add `street` column to `customers` table if missing (idempotent)

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'street'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN street varchar;
    END IF;
END$$;
