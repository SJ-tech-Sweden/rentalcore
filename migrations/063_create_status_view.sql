-- Ensure compatibility view `status` exists mapping to `statuses` table
-- Drop and recreate the view to avoid column-name conflicts.
DO $$
BEGIN
  IF to_regclass('public.statuses') IS NOT NULL THEN
    -- Remove any existing view (safe: it will be recreated immediately).
    EXECUTE 'DROP VIEW IF EXISTS public.status CASCADE';
    EXECUTE 'CREATE VIEW public.status AS SELECT id AS statusid, name AS status FROM public.statuses';
  END IF;
END
$$ LANGUAGE plpgsql;
