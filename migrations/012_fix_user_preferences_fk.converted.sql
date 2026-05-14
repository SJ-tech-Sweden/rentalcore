-- AUTO-CONVERTED (heuristic)
-- Source: 012_fix_user_preferences_fk.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Fix Foreign Key Constraint for User Preferences
-- The constraint was backwards - users should not reference user_preferences
-- Instead, user_preferences should reference users

-- First, drop the incorrect foreign key constraint
DO $$
DECLARE
  pref_exists boolean := to_regclass('public.user_preferences') IS NOT NULL;
  users_exists boolean := to_regclass('public.users') IS NOT NULL;
  pref_col text := NULL;
  users_col text := NULL;
  fk_exists boolean := false;
BEGIN
  IF users_exists THEN
    -- Drop the MySQL-named FK if present (Postgres uses DROP CONSTRAINT)
    EXECUTE 'ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS fk_user_preferences_user';
  END IF;

  IF pref_exists AND users_exists THEN
    -- detect common column name variants for the preference -> users relationship
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='user_preferences' AND column_name='user_id') THEN
      pref_col := 'user_id';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='user_preferences' AND column_name='userID') THEN
      pref_col := 'userID';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='user_preferences' AND column_name='userid') THEN
      pref_col := 'userid';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='userID') THEN
      users_col := 'userID';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='id') THEN
      users_col := 'id';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='user_id') THEN
      users_col := 'user_id';
    END IF;

    -- ensure FK not already present
    SELECT EXISTS(SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_preferences_user') INTO fk_exists;

    IF NOT fk_exists THEN
      IF pref_col IS NOT NULL AND users_col IS NOT NULL THEN
        EXECUTE format('ALTER TABLE user_preferences ADD CONSTRAINT fk_user_preferences_user FOREIGN KEY (%I) REFERENCES users(%I) ON DELETE CASCADE ON UPDATE CASCADE', pref_col, users_col);
      ELSE
        RAISE NOTICE 'Could not find matching columns for user_preferences/users FK; skipping addition';
      END IF;
    ELSE
      RAISE NOTICE 'Foreign key fk_user_preferences_user already exists; skipping';
    END IF;
  ELSE
    RAISE NOTICE 'user_preferences or users relation not found; skipping FK fix';
  END IF;
END$$;