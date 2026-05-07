-- 058_seed_admin_if_missing.sql
-- Idempotent: creates an `admin` user with a temporary password only when missing
BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Insert admin user if not present
INSERT INTO users (username, email, password_hash, is_active, is_admin, force_password_change, created_at)
SELECT 'admin','admin@example.com', crypt('TemporaryAdmin!2026', gen_salt('bf', 12)), true, true, true, now()
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username='admin');

-- If admin exists but has no password, set a temporary password without overwriting non-empty hashes
UPDATE users
SET password_hash = crypt('TemporaryAdmin!2026', gen_salt('bf', 12)),
    force_password_change = true,
    is_active = true
WHERE username = 'admin'
  AND (password_hash IS NULL OR password_hash = '');

-- Assign admin role mapping safely by detecting actual column names
DO $$
DECLARE
  user_col text;
  role_col text;
  mapping_sql text;
  roles_pk text;
BEGIN
  IF to_regclass('public.roles') IS NULL OR to_regclass('public.user_roles') IS NULL THEN
    RAISE NOTICE 'roles or user_roles table not present, skipping role mapping';
    RETURN;
  END IF;

  SELECT column_name INTO user_col
  FROM information_schema.columns
  WHERE table_name='user_roles' AND column_name IN ('userid','user_id')
  LIMIT 1;

  SELECT column_name INTO role_col
  FROM information_schema.columns
  WHERE table_name='user_roles' AND column_name IN ('role_id','roleid','role')
  LIMIT 1;

  -- detect primary key column on roles table (commonly 'id' or 'role_id')
  SELECT column_name INTO roles_pk
  FROM information_schema.columns
  WHERE table_name='roles' AND column_name IN ('id','role_id','roleid')
  LIMIT 1;

  IF roles_pk IS NULL THEN
    RAISE NOTICE 'roles primary key column not found, skipping role mapping';
    RETURN;
  END IF;

  IF user_col IS NULL OR role_col IS NULL THEN
    RAISE NOTICE 'user_roles columns not recognized: user_col=%, role_col=%', user_col, role_col;
    RETURN;
  END IF;

  mapping_sql := format($fmt$
    WITH u AS (SELECT userid FROM users WHERE username='admin')
    INSERT INTO user_roles (%I, %I)
    SELECT u.userid, r.%I
    FROM u, roles r
    WHERE r.name = 'admin'
      AND NOT EXISTS (
        SELECT 1 FROM user_roles ur WHERE ur.%I = u.userid AND ur.%I = r.%I
      );
  $fmt$, user_col, role_col, roles_pk, user_col, role_col, roles_pk);

  EXECUTE mapping_sql;
END
$$;

COMMIT;
