-- 059_fix_user_roles_mapping.sql
-- Safely insert admin role mapping into `user_roles`, detecting column names
DO $$
DECLARE
  user_col text;
  role_col text;
  mapping_sql text;
  roles_pk text;
  users_pk text;
BEGIN
  IF to_regclass('public.user_roles') IS NULL THEN
    RAISE NOTICE 'user_roles table not found, skipping';
    RETURN;
  END IF;

  SELECT column_name INTO user_col
  FROM information_schema.columns
  WHERE table_schema='public' AND table_name='user_roles' AND column_name IN ('userid','user_id')
  LIMIT 1;

  SELECT column_name INTO role_col
  FROM information_schema.columns
  WHERE table_schema='public' AND table_name='user_roles' AND column_name IN ('role_id','roleid','role')
  LIMIT 1;

  IF user_col IS NULL OR role_col IS NULL THEN
    RAISE NOTICE 'user_roles columns not recognized: user_col=%, role_col=%', user_col, role_col;
    RETURN;
  END IF;

  -- detect roles primary key column (id, role_id, roleid)
  SELECT column_name INTO roles_pk
  FROM information_schema.columns
  WHERE table_schema='public' AND table_name='roles' AND column_name IN ('id','role_id','roleid')
  LIMIT 1;

  IF roles_pk IS NULL THEN
    RAISE NOTICE 'roles primary key column not found, skipping mapping';
    RETURN;
  END IF;

  -- detect users primary key / id column
  SELECT column_name INTO users_pk
  FROM information_schema.columns
  WHERE table_schema='public' AND table_name='users' AND column_name IN ('userid','id','user_id','userID')
  LIMIT 1;

  IF users_pk IS NULL THEN
    RAISE NOTICE 'users primary key column not found, skipping mapping';
    RETURN;
  END IF;

  mapping_sql := format($fmt$
    WITH u AS (SELECT %I AS user_pk FROM users WHERE username='admin')
    INSERT INTO user_roles (%I, %I)
    SELECT u.user_pk, r.%I
    FROM u, roles r
    WHERE r.name = 'admin'
      AND NOT EXISTS (
        SELECT 1 FROM user_roles ur WHERE ur.%I = u.user_pk AND ur.%I = r.%I
      );
  $fmt$, users_pk, user_col, role_col, roles_pk, user_col, role_col, roles_pk);

  EXECUTE mapping_sql;
END
$$;
