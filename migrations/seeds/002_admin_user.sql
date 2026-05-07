-- Idempotent seed: create an `admin` user with a temporary password and force password change
-- Uses pgcrypto to hash the temporary password in-database. Change password on first login.
BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS public.seed_marker (name text PRIMARY KEY, applied_at timestamptz DEFAULT now());
INSERT INTO public.seed_marker (name) VALUES ('admin_seed') ON CONFLICT DO NOTHING;

-- Insert admin user with bcrypt-hashed temporary password and force a password change on first login.
-- Change the temporary password below if you want a different initial password.
INSERT INTO users (username, email, password_hash, is_active, force_password_change, created_at)
VALUES (
	'admin',
	'admin@example.test',
	crypt('TemporaryAdmin!2026', gen_salt('bf', 12)),
	TRUE,
	TRUE,
	NOW()
)
ON CONFLICT (username) DO NOTHING;

-- Assign the admin role to the seeded user if roles table contains an `admin` role
INSERT INTO user_roles (userid, roleid, assigned_by, assigned_at, is_active)
SELECT u.userid, r.roleid, NULL, NOW(), TRUE
FROM users u JOIN roles r ON r.name = 'admin'
WHERE u.username = 'admin'
ON CONFLICT (userid, roleid) DO NOTHING;

COMMIT;
