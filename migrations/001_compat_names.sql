-- Compatibility migration: expose consistent table/column names across historical migrations
-- Ensures both `job_devices` / `jobdevices` and `devices` column name variants exist as views

DO $$
BEGIN
  -- If job_devices table exists and jobdevices view does not, create jobdevices view
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'job_devices') THEN
    IF NOT EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'jobdevices') THEN
      EXECUTE $q$
        CREATE VIEW jobdevices AS
        SELECT
          jobid,
          deviceid,
          custom_price,
          package_id,
          is_package_item,
          pack_status,
          pack_ts
        FROM job_devices;
      $q$;
    END IF;
  END IF;

  -- If jobdevices table exists and no real `job_devices` table exists, create job_devices view
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'jobdevices') THEN
    IF to_regclass('public.job_devices') IS NULL AND NOT EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'job_devices') THEN
      EXECUTE $q$
        CREATE VIEW job_devices AS
        SELECT
          jobid,
          deviceid,
          custom_price,
          package_id,
          is_package_item,
          pack_status,
          pack_ts
        FROM jobdevices;
      $q$;
    END IF;
  END IF;

  -- Devices compatibility: if devices table has deviceid (lowercase) or deviceID (mixed), create views mapping both names
  -- If devices table uses deviceid or deviceID column names, expose deviceID as a quoted alias
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'deviceid') THEN
    IF NOT EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'devices_camel') THEN
      EXECUTE $q$
        CREATE VIEW devices_camel AS
        SELECT *, deviceid AS "deviceID" FROM devices;
      $q$;
    END IF;
  END IF;

  -- If devices table uses numeric primary key `id`, expose it as deviceid to satisfy older migrations
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'id') THEN
    IF NOT EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'devices_as_deviceid') THEN
      EXECUTE $q$
        CREATE VIEW devices_as_deviceid AS
        SELECT id::text AS deviceid, id::text AS "deviceID", productid, serialnumber AS serial, zone_id AS zoneid, *
        FROM devices;
      $q$;
    END IF;
  END IF;

END$$;

-- Note: this migration is idempotent and safe to run multiple times.
