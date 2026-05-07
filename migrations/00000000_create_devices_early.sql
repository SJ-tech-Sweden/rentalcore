-- Very-very-early migration: ensure `devices` table exists and has `deviceid`.
-- This runs before other migrations that may reference `devices`.
DO $$
BEGIN
  IF to_regclass('public.devices') IS NULL THEN
    CREATE TABLE devices (
      deviceid VARCHAR(255) PRIMARY KEY,
      productid INT,
      serialnumber TEXT,
      status VARCHAR(50) DEFAULT 'free',
      zone_id INT,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX idx_devices_productid ON devices(productid);
    CREATE INDEX idx_devices_serialnumber ON devices(serialnumber);
    CREATE INDEX idx_devices_zone ON devices(zone_id);
    RAISE NOTICE '00000000_create_devices_early: created devices table';
  ELSE
    RAISE NOTICE '00000000_create_devices_early: devices table already exists';
  END IF;
END$$;
