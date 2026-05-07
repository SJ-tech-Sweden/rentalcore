-- Original MySQL migration converted to no-op for Postgres.
-- The Postgres-compatible migration is in `001_initial_schema.converted.sql`.
DO $$
BEGIN
  RAISE NOTICE 'Skipping MySQL-specific migration 001_initial_schema.sql; converted migration applied instead.';
END;
$$ LANGUAGE plpgsql;