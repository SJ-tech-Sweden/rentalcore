-- Idempotent migration to create user_passkeys for WebAuthn
CREATE TABLE IF NOT EXISTS user_passkeys (
  passkey_id SERIAL PRIMARY KEY,
  user_id INT NOT NULL,
  name VARCHAR(255) NOT NULL,
  credential_id VARCHAR(255) NOT NULL UNIQUE,
  public_key BYTEA,
  sign_count INT DEFAULT 0,
  aaguid BYTEA,
  is_active BOOLEAN DEFAULT TRUE,
  last_used TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'idx_user_passkeys_user') THEN
    CREATE INDEX idx_user_passkeys_user ON user_passkeys(user_id);
  END IF;
  -- `credential_id` has a UNIQUE constraint; no extra index required.
END;
$$ LANGUAGE plpgsql;
