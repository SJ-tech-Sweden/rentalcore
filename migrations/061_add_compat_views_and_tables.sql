-- 061_add_compat_views_and_tables.sql
-- Compatibility objects to satisfy legacy queries expecting different table names/columns
BEGIN;

-- Create compatibility view `status` that maps to `statuses` (if not already present)
DO $$
BEGIN
  IF to_regclass('public.status') IS NULL THEN
    IF to_regclass('public.statuses') IS NOT NULL THEN
      EXECUTE $sql$CREATE OR REPLACE VIEW status AS SELECT id AS statusid, name AS status FROM statuses$sql$;
    END IF;
  END IF;
END
$$ LANGUAGE plpgsql;

-- Create minimal `cases` table if missing
CREATE TABLE IF NOT EXISTS cases (
  caseid SERIAL PRIMARY KEY,
  title TEXT,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create minimal `company_settings` table if missing
CREATE TABLE IF NOT EXISTS company_settings (
  id SERIAL PRIMARY KEY,
  name TEXT,
  value TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create minimal `user_2fa` table used for 2FA checks
CREATE TABLE IF NOT EXISTS user_2fa (
  id SERIAL PRIMARY KEY,
  user_id INTEGER,
  is_enabled BOOLEAN DEFAULT FALSE
);

-- Create minimal `user_dashboard_widgets` table used when rendering dashboard
CREATE TABLE IF NOT EXISTS user_dashboard_widgets (
  id SERIAL PRIMARY KEY,
  user_id INTEGER,
  widget JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMIT;
