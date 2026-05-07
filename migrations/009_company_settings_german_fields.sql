-- Migration: Add German Business Fields to Company Settings
-- Description: Adds banking, legal, and invoice text fields for German compliance (GoBD)
-- Date: 2025-06-20
DO $$
BEGIN
    RAISE NOTICE 'Skipping MySQL-original migration 009_company_settings_german_fields.sql; converted migration should be used instead.';
END$$;