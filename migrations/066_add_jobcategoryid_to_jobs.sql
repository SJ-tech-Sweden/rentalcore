-- Add compatibility column `jobcategoryid` to `jobs` when missing.
-- Idempotent: will add the column and try to populate from common candidates.
DO $$
DECLARE
  candidate text;
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'jobs' AND column_name = 'jobcategoryid'
  ) THEN
    SELECT column_name INTO candidate
    FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'jobs'
      AND column_name IN ('jobcategoryid','jobCategoryID','jobCategoryId','job_categoryid','job_category_id')
    ORDER BY CASE column_name
      WHEN 'jobcategoryid' THEN 0
      WHEN 'jobCategoryID' THEN 1
      WHEN 'jobCategoryId' THEN 2
      WHEN 'job_categoryid' THEN 3
      WHEN 'job_category_id' THEN 4
      ELSE 5 END
    LIMIT 1;

    EXECUTE 'ALTER TABLE public.jobs ADD COLUMN IF NOT EXISTS jobcategoryid integer';

    IF candidate IS NOT NULL THEN
      EXECUTE format('UPDATE public.jobs SET jobcategoryid = (%I)::integer WHERE (%I) IS NOT NULL AND jobcategoryid IS NULL', candidate, candidate);
    END IF;

    -- Optionally, if jobcategory table exists, do not add FK automatically to avoid startup failures
  END IF;
END$$ LANGUAGE plpgsql;
