-- AUTO-CONVERTED (heuristic)
-- Source: 027_add_pdf_totals.down.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Remove total fields from pdf_extractions table
DO $$
BEGIN
    IF to_regclass('public.pdf_extractions') IS NOT NULL THEN
        ALTER TABLE pdf_extractions
            DROP COLUMN IF EXISTS parsed_total,
            DROP COLUMN IF EXISTS discount_percent;
    ELSE
        RAISE NOTICE 'pdf_extractions relation not found; skipping down migration 027_add_pdf_totals';
    END IF;
END $$;
