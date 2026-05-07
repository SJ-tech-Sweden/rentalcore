-- AUTO-CONVERTED (heuristic)
-- Source: 009_company_settings_german_fields.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Migration: Add German Business Fields to Company Settings
-- Description: Adds banking, legal, and invoice text fields for German compliance (GoBD)
-- Date: 2025-06-20

-- Add German banking fields
DO $$
BEGIN
    IF to_regclass('public.company_settings') IS NOT NULL THEN
        -- Add columns if they don't already exist
        EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS bank_name VARCHAR(255) NULL';
        EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS iban VARCHAR(34) NULL';
        EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS bic VARCHAR(11) NULL';
        EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS account_holder VARCHAR(255) NULL';

        EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS ceo_name VARCHAR(255) NULL';
        EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS register_court VARCHAR(255) NULL';
        EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS register_number VARCHAR(100) NULL';

        EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS footer_text TEXT NULL';
        EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS payment_terms_text TEXT NULL';

        -- Update existing record with English defaults if it exists (user preference)
        IF EXISTS (SELECT 1 FROM company_settings WHERE id = 1) THEN
            UPDATE company_settings 
            SET 
                country = 'United Kingdom',
                footer_text = 'Thank you for your business.\n\nIf you have questions about this invoice, please contact us.',
                payment_terms_text = 'Payable within 30 days net.\n\nLate payments may incur interest as permitted by law.'
            WHERE id = 1;
        END IF;

        -- Add indexes for better performance if columns exist
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='company_settings' AND column_name='iban') THEN
            EXECUTE 'CREATE INDEX IF NOT EXISTS idx_company_settings_iban ON company_settings (iban)';
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='company_settings' AND column_name='register_number') THEN
            EXECUTE 'CREATE INDEX IF NOT EXISTS idx_company_settings_register_number ON company_settings (register_number)';
        END IF;

        RAISE NOTICE 'Company settings migration completed (English defaults applied if record existed)';
    ELSE
        RAISE NOTICE 'company_settings relation not found; skipping migration 009';
    END IF;
END$$;