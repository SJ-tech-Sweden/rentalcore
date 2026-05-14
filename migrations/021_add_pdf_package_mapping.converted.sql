-- AUTO-CONVERTED (heuristic)
-- Source: 021_add_pdf_package_mapping.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Migration 021: Add package mapping support for OCR extraction items

DO $$
BEGIN
    -- add column if table exists
    IF to_regclass('public.pdf_extraction_items') IS NOT NULL THEN
        EXECUTE 'ALTER TABLE pdf_extraction_items ADD COLUMN IF NOT EXISTS mapped_package_id INT';
    ELSE
        RAISE NOTICE 'pdf_extraction_items not found; skipping mapped_package_id addition';
    END IF;

    -- create index only if table exists
    IF to_regclass('public.pdf_extraction_items') IS NOT NULL THEN
        EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pdf_items_package ON pdf_extraction_items(mapped_package_id)';
    ELSE
        RAISE NOTICE 'pdf_extraction_items not found; skipping index creation';
    END IF;

    -- add foreign key if both tables exist
    IF to_regclass('public.pdf_extraction_items') IS NOT NULL
       AND to_regclass('public.product_packages') IS NOT NULL THEN
        BEGIN
            EXECUTE 'ALTER TABLE pdf_extraction_items ADD CONSTRAINT fk_pdf_items_package FOREIGN KEY (mapped_package_id) REFERENCES product_packages(id) ON DELETE SET NULL';
        EXCEPTION
            WHEN duplicate_object THEN NULL;
        END;
    ELSE
        RAISE NOTICE 'Skipping FK creation; required relations missing';
    END IF;
END $$;
