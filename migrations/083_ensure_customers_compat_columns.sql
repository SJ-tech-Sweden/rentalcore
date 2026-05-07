-- 083_ensure_customers_compat_columns.sql
-- Ensure legacy/minimal customers tables are compatible with the current Customer model.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'companyname'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN companyname text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'firstname'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN firstname text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'lastname'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN lastname text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'street'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN street text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'housenumber'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN housenumber text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'zip'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN zip text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'city'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN city text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'federalstate'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN federalstate text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'country'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN country text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'phonenumber'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN phonenumber text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'email'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN email text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'customertype'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN customertype text;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'customers' AND column_name = 'notes'
    ) THEN
        ALTER TABLE public.customers ADD COLUMN notes text;
    END IF;
END$$;
