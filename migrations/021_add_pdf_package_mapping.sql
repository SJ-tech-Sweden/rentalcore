-- Migration 021: Add package mapping support for OCR extraction items (noop)

DO $$
BEGIN
    RAISE NOTICE 'Skipping MySQL-original migration 021_add_pdf_package_mapping.sql; converted migration should be used instead.';
END$$;
