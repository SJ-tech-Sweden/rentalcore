-- Ensure `force_password_change` column exists and set admin password if empty
-- Idempotent: will not overwrite an existing non-empty password_hash

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='users' AND column_name='force_password_change') THEN
    ALTER TABLE users ADD COLUMN force_password_change BOOLEAN DEFAULT FALSE;
  END IF;
END;
$$ LANGUAGE plpgsql;

-- Ensure pgcrypto is available for hashing the temporary password in-database
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Use an unpredictable generated password hash.
-- This update will only run when password_hash is NULL or an empty string.
UPDATE users
SET password_hash = crypt(encode(gen_random_bytes(24), 'base64'), gen_salt('bf', 12)),
    force_password_change = TRUE,
    is_active = TRUE
WHERE username = 'admin' AND (password_hash IS NULL OR password_hash = '');
