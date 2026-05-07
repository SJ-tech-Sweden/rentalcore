-- 081_create_subcategories_subbiercategories.sql
-- Idempotent creation of subcategories and subbiercategories tables used by the device/product tree

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'subcategories') THEN
        CREATE TABLE public.subcategories (
            subcategoryid varchar PRIMARY KEY,
            name varchar NOT NULL,
            abbreviation varchar,
            categoryid integer NOT NULL
        );
        CREATE INDEX IF NOT EXISTS idx_subcategories_categoryid ON public.subcategories(categoryid);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'subbiercategories') THEN
        CREATE TABLE public.subbiercategories (
            subbiercategoryid varchar PRIMARY KEY,
            name varchar NOT NULL,
            abbreviation varchar,
            subcategoryid varchar NOT NULL
        );
        CREATE INDEX IF NOT EXISTS idx_subbiercategories_subcategoryid ON public.subbiercategories(subcategoryid);
    END IF;
END$$;
