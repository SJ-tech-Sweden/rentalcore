-- AUTO-CONVERTED (heuristic)
-- Source: 037_accessories_consumables.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Converted from MySQL: accessories/consumables support

-- Accessories table
CREATE TABLE IF NOT EXISTS accessories (
	accessory_id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	sku VARCHAR(100),
	type VARCHAR(50) NOT NULL,
	quantity INT DEFAULT 0,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Consumables table
CREATE TABLE IF NOT EXISTS consumables (
	consumable_id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	sku VARCHAR(100),
	quantity INT DEFAULT 0,
	unit VARCHAR(50),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add check constraint for accessory type if needed
DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_accessory_type') THEN
		EXECUTE 'ALTER TABLE accessories ADD CONSTRAINT chk_accessory_type CHECK (type IN (''accessory'',''consumable'',''other''))';
	END IF;
END$$;

-- Example trigger: when consumable used, decrement quantity if column exists
-- Create or replace the trigger function (idempotent)
CREATE OR REPLACE FUNCTION trg_consumable_usage_decrement() RETURNS TRIGGER AS $$
BEGIN
	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'consumables' AND column_name = 'quantity') THEN
		UPDATE consumables SET quantity = GREATEST(COALESCE(quantity,0) - NEW.used_quantity, 0)
		WHERE consumable_id = NEW.consumable_id;
	END IF;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create the trigger only if the related table exists and trigger is absent
DO $$
BEGIN
	IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'consumable_usages') THEN
		IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_consumable_usage_decrement') THEN
			EXECUTE 'CREATE TRIGGER trg_consumable_usage_decrement AFTER INSERT ON consumable_usages FOR EACH ROW EXECUTE FUNCTION trg_consumable_usage_decrement();';
		END IF;
	END IF;
END$$;
