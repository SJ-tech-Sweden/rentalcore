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
INSERT INTO jobs (jobid, customerid, name, status, created_at)
VALUES (1, 1, 'Init job', 'open', NOW()) ON CONFLICT DO NOTHING;

COMMIT;
