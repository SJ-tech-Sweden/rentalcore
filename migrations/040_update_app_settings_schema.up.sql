-- Migration: update app_settings schema
-- Description: Add id (auto-increment PK) and scope columns to app_settings.
--              The scope column allows RentalCore and WarehouseCore to store
--              settings under different scopes while sharing the same table.
--              A unique constraint on (scope, key) replaces the old key-only PK.

-- Add id column as the new primary key
ALTER TABLE app_settings ADD COLUMN IF NOT EXISTS id SERIAL;

-- Add scope column (default 'global' so existing rows are migrated automatically)
ALTER TABLE app_settings ADD COLUMN IF NOT EXISTS scope VARCHAR(50) NOT NULL DEFAULT 'global';

-- Drop the old primary key constraint on key
ALTER TABLE app_settings DROP CONSTRAINT IF EXISTS app_settings_pkey;

-- Make id the new primary key
ALTER TABLE app_settings ADD PRIMARY KEY (id);

-- Add unique constraint on (scope, key) to prevent duplicates and support ON CONFLICT upserts
ALTER TABLE app_settings DROP CONSTRAINT IF EXISTS idx_app_settings_scope_key;
ALTER TABLE app_settings ADD CONSTRAINT idx_app_settings_scope_key UNIQUE (scope, key);
