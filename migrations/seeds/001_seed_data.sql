-- rentalcore seed data (idempotent)
BEGIN;

CREATE TABLE IF NOT EXISTS public.seed_marker (name text PRIMARY KEY, applied_at timestamptz DEFAULT now());

INSERT INTO public.seed_marker (name) VALUES ('initial_seed') ON CONFLICT DO NOTHING;

-- Roles
INSERT INTO roles (roleid, name) VALUES (1, 'admin') ON CONFLICT DO NOTHING;
INSERT INTO roles (roleid, name) VALUES (2, 'user') ON CONFLICT DO NOTHING;

-- Product samples
INSERT INTO products (productid, name)
VALUES (1, 'Widget A') ON CONFLICT DO NOTHING;
INSERT INTO products (productid, name)
VALUES (2, 'Cable 1m') ON CONFLICT DO NOTHING;

-- Customer sample
INSERT INTO customers (customerid, companyname, email, created_at)
VALUES (1, 'Acme Events', 'ops@acme.test', NOW()) ON CONFLICT DO NOTHING;

-- Job sample
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'jobs' AND column_name = 'job_code'
  ) THEN
    INSERT INTO jobs (jobid, job_code, customerid, name, status, created_at)
    VALUES (1, 'JOB000001', 1, 'Init job', 'open', NOW()) ON CONFLICT DO NOTHING;
  ELSE
    INSERT INTO jobs (jobid, customerid, name, status, created_at)
    VALUES (1, 1, 'Init job', 'open', NOW()) ON CONFLICT DO NOTHING;
  END IF;
END$$;

-- Keep sequences aligned after explicit ID inserts
SELECT setval(pg_get_serial_sequence('roles', 'roleid'), COALESCE((SELECT MAX(roleid) FROM roles), 1), true);
SELECT setval(pg_get_serial_sequence('products', 'productid'), COALESCE((SELECT MAX(productid) FROM products), 1), true);
SELECT setval(pg_get_serial_sequence('customers', 'customerid'), COALESCE((SELECT MAX(customerid) FROM customers), 1), true);
SELECT setval(pg_get_serial_sequence('jobs', 'jobid'), COALESCE((SELECT MAX(jobid) FROM jobs), 1), true);

COMMIT;
