-- Migration 027 up (noop)
DO $$
BEGIN
    RAISE NOTICE 'Skipping MySQL-original migration 027_add_pdf_totals.up.sql; converted migration should be used instead.';
END $$;

-- Note: discount_amount and total_amount columns already exist
