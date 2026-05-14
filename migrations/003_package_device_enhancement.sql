-- AUTO-CONVERTED (heuristic)
-- Source: 003_package_device_enhancement.sql.bak
-- Review this file for correctness before applying to Postgres.

-- MySQL-heavy migration 003_package_device_enhancement.sql moved to .bak for manual conversion
-- Converted from MySQL. Review triggers and business logic.

-- Add package_devices bridge table if missing
CREATE TABLE IF NOT EXISTS package_devices (
	package_id INT NOT NULL,
	device_id VARCHAR(255) NOT NULL,
	quantity INT NOT NULL DEFAULT 1,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (package_id, device_id)
);

-- Ensure indexes
CREATE INDEX IF NOT EXISTS idx_package_devices_package_id ON package_devices(package_id);
CREATE INDEX IF NOT EXISTS idx_package_devices_device_id ON package_devices(device_id);

-- Add helper trigger: when package_devices inserted, decrement device stock (if such column exists)
-- Create the trigger function unconditionally (idempotent via CREATE OR REPLACE)
CREATE OR REPLACE FUNCTION trg_package_devices_adjust() RETURNS TRIGGER AS $$
BEGIN
	-- If devices table has stock_quantity, decrement it when package_devices inserted
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'stock_quantity') THEN
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'deviceid') THEN
				EXECUTE 'UPDATE devices SET stock_quantity = GREATEST(COALESCE(stock_quantity,0) - $1, 0) WHERE deviceid = $2' USING NEW.quantity, NEW.device_id;
			ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = '"deviceID"') THEN
				EXECUTE 'UPDATE devices SET stock_quantity = GREATEST(COALESCE(stock_quantity,0) - $1, 0) WHERE "deviceID" = $2' USING NEW.quantity, NEW.device_id;
			ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'id') THEN
				EXECUTE 'UPDATE devices SET stock_quantity = GREATEST(COALESCE(stock_quantity,0) - $1, 0) WHERE id::text = $2' USING NEW.quantity, NEW.device_id;
			ELSE
				-- No known device identifier column; skip update
				RAISE NOTICE 'trg_package_devices_adjust: no device identifier column found, skipping stock update';
			END IF;
		END IF;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create the trigger only if the table exists and the trigger is not already present
DO $$
BEGIN
	IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'package_devices') THEN
		IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_package_devices_adjust') THEN
			EXECUTE 'CREATE TRIGGER trg_package_devices_adjust AFTER INSERT ON package_devices FOR EACH ROW EXECUTE FUNCTION trg_package_devices_adjust();';
		END IF;
	END IF;
END$$;
