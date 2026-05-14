-- Migration 070: Create minimal `categories` table for compatibility
-- Purpose: Some pages query `categories` (ORDER BY name); ensure table exists.

DO $$
BEGIN
    IF to_regclass('categories') IS NOT NULL THEN
        RAISE NOTICE 'categories already exists; skipping migration 070_create_categories_table';
        RETURN;
    END IF;

    CREATE TABLE categories (
        categoryid SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        abbreviation VARCHAR(64)
    );

    RAISE NOTICE 'Created categories table (migration 070_create_categories_table)';
END
$$;
