-- AUTO-CONVERTED (heuristic)
-- Source: 027_add_pdf_totals.up.sql.bak
-- Review this file for correctness before applying to Postgres.

DO $$
BEGIN
    IF to_regclass('pdf_extractions') IS NULL THEN
        RAISE NOTICE 'pdf_extractions relation not found; skipping migration 027_add_pdf_totals.up';
    ELSE
        ALTER TABLE pdf_extractions
            ADD COLUMN IF NOT EXISTS parsed_total numeric(10,2),
            ADD COLUMN IF NOT EXISTS discount_percent numeric(5,2);
        RAISE NOTICE 'Applied migration 027_add_pdf_totals.up on pdf_extractions';
    END IF;
END$$;

-- Note: discount_amount and total_amount columns already exist
