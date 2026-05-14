-- 086_ensure_company_settings_modern_columns.sql
-- Ensure company_settings has modern columns required by CompanyHandler/CompanySettings model.

DO $$
BEGIN
    -- Create table if missing (modern-compatible shape)
    IF to_regclass('public.company_settings') IS NULL THEN
        CREATE TABLE public.company_settings (
            id SERIAL PRIMARY KEY,
            company_name TEXT NOT NULL DEFAULT 'RentalCore',
            address_line1 TEXT,
            address_line2 TEXT,
            city TEXT,
            state TEXT,
            postal_code TEXT,
            country TEXT,
            phone TEXT,
            email TEXT,
            website TEXT,
            tax_number TEXT,
            vat_number TEXT,
            logo_path TEXT,
            bank_name TEXT,
            iban TEXT,
            bic TEXT,
            account_holder TEXT,
            ceo_name TEXT,
            register_court TEXT,
            register_number TEXT,
            footer_text TEXT,
            payment_terms_text TEXT,
            smtp_host TEXT,
            smtp_port INT,
            smtp_username TEXT,
            smtp_password TEXT,
            smtp_from_email TEXT,
            smtp_from_name TEXT,
            smtp_use_tls BOOLEAN DEFAULT TRUE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    END IF;

    -- Add missing columns for legacy/minimal compatibility tables
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS company_name TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS address_line1 TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS address_line2 TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS city TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS state TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS postal_code TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS country TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS phone TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS email TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS website TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS tax_number TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS vat_number TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS logo_path TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS bank_name TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS iban TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS bic TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS account_holder TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS ceo_name TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS register_court TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS register_number TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS footer_text TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS payment_terms_text TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS smtp_host TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS smtp_port INT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS smtp_username TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS smtp_password TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS smtp_from_email TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS smtp_from_name TEXT;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS smtp_use_tls BOOLEAN DEFAULT TRUE;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE public.company_settings ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

    -- Backfill modern company_name from legacy name if present
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'company_settings' AND column_name = 'name'
    ) THEN
        EXECUTE 'UPDATE public.company_settings SET company_name = COALESCE(NULLIF(company_name, ''''), name, ''RentalCore'')';
    ELSE
        UPDATE public.company_settings
        SET company_name = COALESCE(NULLIF(company_name, ''), 'RentalCore');
    END IF;

    -- Ensure at least one row exists for UI load/save flow
    IF NOT EXISTS (SELECT 1 FROM public.company_settings) THEN
        INSERT INTO public.company_settings (company_name)
        VALUES ('RentalCore');
    END IF;

    -- Make company_name non-null after backfill
    ALTER TABLE public.company_settings ALTER COLUMN company_name SET DEFAULT 'RentalCore';
    ALTER TABLE public.company_settings ALTER COLUMN company_name SET NOT NULL;
END$$;
