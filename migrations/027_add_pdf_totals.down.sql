-- Remove total fields from pdf_extractions table
ALTER TABLE pdf_extractions
    DROP COLUMN IF EXISTS parsed_total,
    DROP COLUMN IF EXISTS discount_percent;
