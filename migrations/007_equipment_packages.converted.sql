-- AUTO-CONVERTED (heuristic)
-- Source: 007_equipment_packages.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Enhanced Equipment Packages Migration (MySQL Compatible)
-- Add new fields for production-ready equipment packages

-- Add new columns to equipment_packages table
-- Note: Run each ALTER TABLE separately if you get syntax errors
DO $$
BEGIN
	IF to_regclass('public.equipment_packages') IS NOT NULL THEN
		-- Add columns if they don't already exist
		EXECUTE 'ALTER TABLE equipment_packages ADD COLUMN IF NOT EXISTS max_rental_days INT NULL';
		EXECUTE 'ALTER TABLE equipment_packages ADD COLUMN IF NOT EXISTS category VARCHAR(50) NULL';
		EXECUTE 'ALTER TABLE equipment_packages ADD COLUMN IF NOT EXISTS tags TEXT NULL';
		EXECUTE 'ALTER TABLE equipment_packages ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP NULL';
		EXECUTE 'ALTER TABLE equipment_packages ADD COLUMN IF NOT EXISTS total_revenue NUMERIC(12,2) DEFAULT 0.00';

		-- Create indexes only when target columns exist
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='equipment_packages' AND column_name='category') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_equipment_packages_category ON equipment_packages(category)';
		END IF;

		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='equipment_packages' AND column_name='is_active') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_equipment_packages_active ON equipment_packages(is_active)';
		END IF;

		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='equipment_packages' AND column_name='usage_count') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_equipment_packages_usage ON equipment_packages(usage_count)';
		END IF;

		-- Safely update package_items only if the column exists
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='equipment_packages' AND column_name='package_items') THEN
			EXECUTE 'UPDATE equipment_packages SET package_items = ''[]'' WHERE package_items IS NULL OR package_items = '''''';';
		END IF;
	ELSE
		RAISE NOTICE 'equipment_packages relation not found; skipping alterations and updates';
	END IF;

	-- Package_devices indexes (may be created by earlier migrations)
	IF to_regclass('public.package_devices') IS NOT NULL THEN
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='package_devices' AND column_name='package_id') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_package_devices_package_id ON package_devices(package_id)';
		END IF;
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='package_devices' AND column_name='device_id') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_package_devices_device_id ON package_devices(device_id)';
		END IF;
	ELSE
		RAISE NOTICE 'package_devices relation not found; skipping package_devices indexes';
	END IF;
END$$;

-- Sample equipment packages removed for production use
-- Create packages through the application interface as needed