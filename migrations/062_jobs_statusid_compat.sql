-- Add a compatibility column `statusid` to `jobs` when missing.
-- This migration is idempotent and will try to populate the new column
-- from any existing candidate column (statusID, statusId, status_id, status).
DO $$
DECLARE
  candidate text;
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'jobs' AND column_name = 'statusid'
  ) THEN
    SELECT column_name INTO candidate
    FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'jobs'
      AND column_name IN ('statusid','statusID','statusId','status_id','status')
    ORDER BY CASE column_name
      WHEN 'statusid' THEN 0
      WHEN 'statusID' THEN 1
      WHEN 'statusId' THEN 2
      WHEN 'status_id' THEN 3
      WHEN 'status' THEN 4
      ELSE 5 END
    LIMIT 1;

    -- Add the compatibility column (no FK by default to avoid failures).
    EXECUTE 'ALTER TABLE public.jobs ADD COLUMN IF NOT EXISTS statusid integer';

    -- If we found a candidate, copy values into the new column.
    IF candidate IS NOT NULL THEN
      EXECUTE format('UPDATE public.jobs SET statusid = (%I)::integer WHERE (%I) IS NOT NULL AND statusid IS NULL', candidate, candidate);
    END IF;
  END IF;
END$$;

-- Ensure a helpful comment exists so this migration is recognizable.
COMMENT ON TABLE public.jobs IS 'compat: statusid column added by migration 062_jobs_statusid_compat.sql when missing';
