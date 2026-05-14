-- Very-very-early migration: ensure `devices` table exists and has `deviceid`.
-- This runs before other migrations that may reference `devices`.
DO $$
BEGIN
  IF to_regclass('public.devices') IS NULL THEN
    CREATE TABLE devices (
      deviceid VARCHAR(255) PRIMARY KEY,
      productid INT,
      serialnumber TEXT,
      purchasedate DATE,
      lastmaintenance DATE,
      nextmaintenance DATE,
      insurancenumber TEXT,
      status VARCHAR(50) DEFAULT 'free',
      insuranceid INT,
      qr_code TEXT,
      current_location TEXT,
      gps_latitude NUMERIC,
      gps_longitude NUMERIC,
      condition_rating NUMERIC DEFAULT 5.0,
      usage_hours NUMERIC DEFAULT 0.0,
      total_revenue NUMERIC DEFAULT 0.0,
      last_maintenance_cost NUMERIC,
      notes TEXT,
      barcode TEXT,
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
