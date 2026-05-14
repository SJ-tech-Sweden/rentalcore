-- Ensure `devices` table exists for RentalCore (idempotent)
BEGIN;

CREATE TABLE IF NOT EXISTS devices (
  deviceid VARCHAR(255) PRIMARY KEY,
  productid INT,
  serialnumber TEXT,
  status VARCHAR(50) DEFAULT 'free',
  zone_id INT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_devices_productid ON devices(productid);
CREATE INDEX IF NOT EXISTS idx_devices_serialnumber ON devices(serialnumber);
CREATE INDEX IF NOT EXISTS idx_devices_zone ON devices(zone_id);

COMMIT;
