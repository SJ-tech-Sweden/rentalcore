-- Fix Foreign Key Constraint for User Preferences
-- The constraint was backwards - users should not reference user_preferences
-- Instead, user_preferences should reference users

-- First, drop the incorrect foreign key constraint
DO $$
BEGIN
    RAISE NOTICE 'Skipping MySQL-original migration 012_fix_user_preferences_fk.sql; converted migration should be used instead.';
END$$;