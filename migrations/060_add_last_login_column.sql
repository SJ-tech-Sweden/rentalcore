-- 060_add_last_login_column.sql
-- Add missing `last_login` column to users table (idempotent)
BEGIN;

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS last_login timestamp;

COMMIT;
