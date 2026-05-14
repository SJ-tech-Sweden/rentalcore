-- AUTO-CONVERTED (heuristic)
-- Source: 011_add_email_settings.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Add email configuration fields to company_settings table
DO $$
BEGIN
	IF to_regclass('public.company_settings') IS NOT NULL THEN
		EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS smtp_host VARCHAR(255)';
		EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS smtp_port INT';
		EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS smtp_username VARCHAR(255)';
		EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS smtp_password VARCHAR(255)';
		EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS smtp_from_email VARCHAR(255)';
		EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS smtp_from_name VARCHAR(255)';
		EXECUTE 'ALTER TABLE company_settings ADD COLUMN IF NOT EXISTS smtp_use_tls BOOLEAN DEFAULT TRUE';
	ELSE
		RAISE NOTICE 'company_settings relation not found; skipping email settings migration 011';
	END IF;
END$$;