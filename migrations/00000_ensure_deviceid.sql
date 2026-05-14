-- Very-early migration: ensure `deviceid` exists on `devices` so subsequent migrations can rely on it.
DO $$
BEGIN
  IF to_regclass('public.devices') IS NULL THEN
    RAISE NOTICE 'early_ensure_deviceid: devices table missing; skipping';
    RETURN;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'deviceid') THEN
    ALTER TABLE devices ADD COLUMN deviceid VARCHAR(255);
    -- Try to populate from common sources
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'id') THEN
      EXECUTE 'UPDATE devices SET deviceid = id::text WHERE deviceid IS NULL';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'serialnumber') THEN
      EXECUTE 'UPDATE devices SET deviceid = serialnumber WHERE deviceid IS NULL';
    ELSE
      RAISE NOTICE 'early_ensure_deviceid: no id/serialnumber column to derive deviceid from';
    END IF;
  ELSE
    RAISE NOTICE 'early_ensure_deviceid: deviceid already present';
  END IF;
END$$;
