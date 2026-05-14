-- Fix corrupted datetime values in company_settings table (Postgres-safe guard)
DO $$
BEGIN
   IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'company_settings') THEN
      UPDATE company_settings
      SET created_at = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
      WHERE created_at IS NULL
          OR updated_at IS NULL;
   END IF;
END$$;
