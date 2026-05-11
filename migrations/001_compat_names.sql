-- Compatibility migration: expose consistent table/column names across historical migrations
-- Ensures both `job_devices` / `jobdevices` and `devices` column name variants exist as views

DO $$
BEGIN
  -- If job_devices table exists and jobdevices view does not, create jobdevices view
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'job_devices') THEN
    IF NOT EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'jobdevices') THEN
      DECLARE
        sel_cols text := '';
        col_jobid text := 'jobid';
      BEGIN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'jobid') THEN
          col_jobid := 'jobid';
        ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'job_id') THEN
          col_jobid := 'job_id AS jobid';
        ELSE
          col_jobid := 'NULL::bigint AS jobid';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'deviceid') THEN
          sel_cols := sel_cols || ', deviceid';
        ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'device_id') THEN
          sel_cols := sel_cols || ', device_id AS deviceid';
        ELSE
          sel_cols := sel_cols || ', NULL::varchar AS deviceid';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'custom_price') THEN
          sel_cols := sel_cols || ', custom_price';
        ELSE
          sel_cols := sel_cols || ', NULL::numeric AS custom_price';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'package_id') THEN
          sel_cols := sel_cols || ', package_id';
        ELSE
          sel_cols := sel_cols || ', NULL::int AS package_id';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'is_package_item') THEN
          sel_cols := sel_cols || ', is_package_item';
        ELSE
          sel_cols := sel_cols || ', false AS is_package_item';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'pack_status') THEN
          sel_cols := sel_cols || ', pack_status';
        ELSE
          sel_cols := sel_cols || ', ''pending''::varchar AS pack_status';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'pack_ts') THEN
          sel_cols := sel_cols || ', pack_ts';
        ELSE
          sel_cols := sel_cols || ', NULL::timestamp AS pack_ts';
        END IF;

        EXECUTE format('CREATE VIEW jobdevices AS SELECT %s %s FROM job_devices', col_jobid, sel_cols);
      END;
    END IF;
  END IF;

  -- If jobdevices table exists and no real `job_devices` table exists, create job_devices view
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'jobdevices') THEN
    IF to_regclass('public.job_devices') IS NULL AND NOT EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'job_devices') THEN
      DECLARE
        sel_cols text := '';
        col_jobid text := 'jobid';
      BEGIN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'jobdevices' AND column_name = 'jobid') THEN
          col_jobid := 'jobid';
        ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'jobdevices' AND column_name = 'job_id') THEN
          col_jobid := 'job_id AS jobid';
        ELSE
          col_jobid := 'NULL::bigint AS jobid';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'jobdevices' AND column_name = 'deviceid') THEN
          sel_cols := sel_cols || ', deviceid';
        ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'jobdevices' AND column_name = 'device_id') THEN
          sel_cols := sel_cols || ', device_id AS deviceid';
        ELSE
          sel_cols := sel_cols || ', NULL::varchar AS deviceid';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'jobdevices' AND column_name = 'custom_price') THEN
          sel_cols := sel_cols || ', custom_price';
        ELSE
          sel_cols := sel_cols || ', NULL::numeric AS custom_price';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'jobdevices' AND column_name = 'package_id') THEN
          sel_cols := sel_cols || ', package_id';
        ELSE
          sel_cols := sel_cols || ', NULL::int AS package_id';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'jobdevices' AND column_name = 'is_package_item') THEN
          sel_cols := sel_cols || ', is_package_item';
        ELSE
          sel_cols := sel_cols || ', false AS is_package_item';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'jobdevices' AND column_name = 'pack_status') THEN
          sel_cols := sel_cols || ', pack_status';
        ELSE
          sel_cols := sel_cols || ', ''pending''::varchar AS pack_status';
        END IF;

        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'jobdevices' AND column_name = 'pack_ts') THEN
          sel_cols := sel_cols || ', pack_ts';
        ELSE
          sel_cols := sel_cols || ', NULL::timestamp AS pack_ts';
        END IF;

        EXECUTE format('CREATE VIEW job_devices AS SELECT %s %s FROM jobdevices', col_jobid, sel_cols);
      END;
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
