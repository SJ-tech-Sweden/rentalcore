-- Original MySQL migration converted to no-op for Postgres.
-- Reference backup is in `001_initial_schema.converted.todo`.
DO $$
BEGIN
  RAISE NOTICE 'Skipping MySQL-specific migration 001_initial_schema.sql; converted migration applied instead.';
END;
$$ LANGUAGE plpgsql;
