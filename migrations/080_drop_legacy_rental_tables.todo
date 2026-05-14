-- 080_drop_legacy_rental_tables.sql
-- Idempotent migration to drop legacy rental_equipment and job_rental_equipment tables.
-- This should only be applied after ensuring WarehouseCore provides the required data
-- and the application is configured to use WarehouseCore reads.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'job_rental_equipment') THEN
        RAISE NOTICE 'Dropping table job_rental_equipment';
        EXECUTE 'DROP TABLE IF EXISTS job_rental_equipment CASCADE';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'rental_equipment') THEN
        RAISE NOTICE 'Dropping table rental_equipment';
        EXECUTE 'DROP TABLE IF EXISTS rental_equipment CASCADE';
    END IF;

    -- Optionally drop any legacy analytics or views related to rental equipment
    IF EXISTS (SELECT 1 FROM information_schema.views WHERE table_schema = 'public' AND table_name = 'rental_equipment_analytics') THEN
        RAISE NOTICE 'Dropping view rental_equipment_analytics';
        EXECUTE 'DROP VIEW IF EXISTS rental_equipment_analytics CASCADE';
    END IF;
END$$;

-- Safe no-op if tables are already removed.
