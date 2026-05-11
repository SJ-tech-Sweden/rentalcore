-- Ensure `deviceid` column exists on `devices` and populate from known identifiers
DO $$
BEGIN
  IF to_regclass('public.devices') IS NULL THEN
    RAISE NOTICE 'ensure_deviceid: devices table not present, skipping';
    RETURN;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'deviceid') THEN
    ALTER TABLE devices ADD COLUMN deviceid VARCHAR(255);
    -- Populate deviceid from best available source
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'id') THEN
      EXECUTE 'UPDATE devices SET deviceid = id::text WHERE deviceid IS NULL';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'deviceID') THEN
      EXECUTE 'UPDATE devices SET deviceid = "deviceID" WHERE deviceid IS NULL';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'serialnumber') THEN
      EXECUTE 'UPDATE devices SET deviceid = serialnumber WHERE deviceid IS NULL';
    ELSE
      RAISE NOTICE 'ensure_deviceid: no source found to populate deviceid; left NULL';
    END IF;
  ELSE
    RAISE NOTICE 'ensure_deviceid: deviceid column already exists';
  END IF;
END$$;
