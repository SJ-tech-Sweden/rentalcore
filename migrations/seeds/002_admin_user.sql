-- Idempotent seed: create an `admin` user with an unpredictable generated password hash.
-- Uses pgcrypto and avoids storing a known default password in the repository.
BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS public.seed_marker (name text PRIMARY KEY, applied_at timestamptz DEFAULT now());
INSERT INTO public.seed_marker (name) VALUES ('admin_seed') ON CONFLICT DO NOTHING;

-- Insert admin user with a bcrypt hash generated from cryptographically random bytes.
-- Password reset should be performed through approved user-management/admin flows
-- (for example create-production-user.sh or the admin reset endpoint/tooling).
INSERT INTO users (username, email, password_hash, is_active, force_password_change, created_at)
SELECT
	'admin',
	'admin@example.test',
	crypt(encode(gen_random_bytes(24), 'base64'), gen_salt('bf', 12)),
	TRUE,
	TRUE,
	NOW()
WHERE NOT EXISTS (
	SELECT 1
	FROM users
	WHERE username = 'admin' OR email = 'admin@example.test'
)
ON CONFLICT DO NOTHING;

-- Assign the admin role to the seeded user if roles table contains an `admin` role
INSERT INTO user_roles (userid, roleid, assigned_by, assigned_at, is_active)
SELECT u.userid, r.roleid, NULL, NOW(), TRUE
FROM users u JOIN roles r ON r.name = 'admin'
WHERE u.username = 'admin'
ON CONFLICT (userid, roleid) DO NOTHING;

COMMIT;
