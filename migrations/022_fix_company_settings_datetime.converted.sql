-- AUTO-CONVERTED (heuristic)
-- Source: 022_fix_company_settings_datetime.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Fix NULL TIMESTAMP values in company_settings table.
-- MySQL zero-date literals are invalid in PostgreSQL and cannot be stored in
-- these TIMESTAMP columns after import, so this converted migration only
-- normalizes NULL values.
DO $$
BEGIN
   IF to_regclass('public.company_settings') IS NOT NULL THEN
      UPDATE company_settings
      SET created_at = CURRENT_TIMESTAMP,
         updated_at = CURRENT_TIMESTAMP
      WHERE created_at IS NULL
         OR updated_at IS NULL;
   ELSE
      RAISE NOTICE 'company_settings relation not found; skipping migration 022_fix_company_settings_datetime';
   END IF;
END $$;
