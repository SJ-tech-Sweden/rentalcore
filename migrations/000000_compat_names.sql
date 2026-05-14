
DO $$
BEGIN
  -- Create jobdevices view if job_devices exists, mapping device_id -> deviceid when necessary
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'job_devices') THEN
    IF NOT EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'jobdevices') THEN
      -- Build select list dynamically to avoid referencing missing columns
      DECLARE
        sel_cols TEXT := '';
        col_jobid TEXT := NULL;
      BEGIN
        -- job column
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'jobid') THEN
          col_jobid := 'jobid';
        ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'job_id') THEN
          col_jobid := 'job_id';
        ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'jobID') THEN
          col_jobid := 'jobID';
        ELSE
          col_jobid := 'NULL::int AS jobid';
        END IF;

        -- device column
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'deviceid') THEN
          sel_cols := sel_cols || ', deviceid';
        ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'device_id') THEN
          sel_cols := sel_cols || ', device_id AS deviceid';
        ELSE
          sel_cols := sel_cols || ', NULL::varchar AS deviceid';
        END IF;

        -- custom_price
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'custom_price') THEN
          sel_cols := sel_cols || ', custom_price';
        ELSE
          sel_cols := sel_cols || ', NULL::numeric AS custom_price';
        END IF;

        -- package_id
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'package_id') THEN
          sel_cols := sel_cols || ', package_id';
        ELSE
          sel_cols := sel_cols || ', NULL::int AS package_id';
        END IF;

        -- is_package_item
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'is_package_item') THEN
          sel_cols := sel_cols || ', is_package_item';
        ELSE
          sel_cols := sel_cols || ', false AS is_package_item';
        END IF;

        -- pack_status
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'pack_status') THEN
          sel_cols := sel_cols || ', pack_status';
        ELSE
          sel_cols := sel_cols || ', ''pending''::varchar AS pack_status';
        END IF;

        -- pack_ts
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'job_devices' AND column_name = 'pack_ts') THEN
          sel_cols := sel_cols || ', pack_ts';
        ELSE
          sel_cols := sel_cols || ', NULL::timestamp AS pack_ts';
        END IF;

        EXECUTE format('CREATE OR REPLACE VIEW jobdevices AS SELECT %s %s FROM job_devices', col_jobid, sel_cols);
      END;
    END IF;
  END IF;

  -- Create job_devices view if `jobdevices` exists but only when no table/relation `job_devices` already exists
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'jobdevices') THEN
    IF to_regclass('public.job_devices') IS NULL THEN
      IF NOT EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'job_devices') THEN
        EXECUTE $q$
          CREATE OR REPLACE VIEW job_devices AS
          SELECT jobid, deviceid, custom_price, package_id, is_package_item, pack_status, pack_ts
          FROM jobdevices;
        $q$;
      END IF;
    END IF;
  END IF;

  -- Expose deviceid alias when devices have numeric id
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'id') THEN
    IF NOT EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'devices_as_deviceid') THEN
      DECLARE
        dev_cols TEXT := 'id::text AS deviceid, id::text AS "deviceID"';
      BEGIN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='productid') THEN
          dev_cols := dev_cols || ', productid';
        ELSE
          dev_cols := dev_cols || ', NULL::int AS productid';
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='serialnumber') THEN
          dev_cols := dev_cols || ', serialnumber AS serial';
        ELSE
          dev_cols := dev_cols || ', NULL::text AS serial';
        END IF;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='zone_id') THEN
          dev_cols := dev_cols || ', zone_id AS zoneid';
        ELSE
          dev_cols := dev_cols || ', NULL::int AS zoneid';
        END IF;
        EXECUTE format('CREATE OR REPLACE VIEW devices_as_deviceid AS SELECT %s FROM devices', dev_cols);
      END;
    END IF;
  END IF;

  -- If devices table uses deviceid column, expose quoted "deviceID" alias
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'deviceid') THEN
    IF NOT EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'devices_camel') THEN
      EXECUTE $q$
        CREATE OR REPLACE VIEW devices_camel AS
        SELECT *, deviceid AS "deviceID" FROM devices;
      $q$;
    END IF;
  END IF;

END$$;
