-- AUTO-CONVERTED (heuristic)
-- Source: 006_performance_indexes.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Performance optimization indexes
-- Add indexes for commonly searched fields and join operations

-- Jobs table indexes
DO $$
BEGIN
	-- description
	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='description') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_description ON jobs(description)';
	ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='desc') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_description ON jobs("desc")';
	END IF;

	-- end date (support end_date, endDate)
	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='end_date') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_end_date ON jobs(end_date)';
	ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='endDate') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_end_date ON jobs("endDate")';
	END IF;

	-- customer id (support customer_id, customerID, customerid)
	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='customer_id') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_customer_id ON jobs(customer_id)';
	ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='customerID') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_customer_id ON jobs("customerID")';
	ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='customerid') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_customer_id ON jobs(customerid)';
	END IF;

	-- status id (support status_id, statusID, statusid)
	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='status_id') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_status_id ON jobs(status_id)';
	ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='statusID') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_status_id ON jobs("statusID")';
	ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='statusid') THEN
		EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_status_id ON jobs(statusid)';
	END IF;
END$$;

-- Customers table indexes  
CREATE INDEX IF NOT EXISTS idx_customers_search ON customers(companyname, firstname, lastname);
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);

-- Devices and jobdevices indexes (guarded)
DO $$
BEGIN
	IF to_regclass('public.devices') IS NOT NULL THEN
		-- create indexes only if the referenced columns exist
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='deviceID') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_devices_search ON devices("deviceID", serialnumber)';
		ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='deviceid') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_devices_search ON devices(deviceid, serialnumber)';
		END IF;

		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='productID') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_devices_product_id ON devices(productID)';
		ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='productid') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_devices_product_id ON devices(productid)';
		END IF;

		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='status') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status)';
		END IF;
	END IF;

	IF to_regclass('public.jobdevices') IS NOT NULL THEN
		-- only create indexes if jobdevices is a real table, not a view
		IF (SELECT relkind FROM pg_class WHERE relname = 'jobdevices' LIMIT 1) = 'v' THEN
			RAISE NOTICE 'jobdevices is a view; skipping index creation';
		ELSE
			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobdevices' AND column_name='jobID') THEN
				EXECUTE 'CREATE INDEX IF NOT EXISTS idx_job_devices_job_id ON jobdevices(jobID)';
			ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobdevices' AND column_name='jobid') THEN
				EXECUTE 'CREATE INDEX IF NOT EXISTS idx_job_devices_job_id ON jobdevices(jobid)';
			END IF;

			IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobdevices' AND column_name='deviceID') THEN
				EXECUTE 'CREATE INDEX IF NOT EXISTS idx_job_devices_device_id ON jobdevices("deviceID")';
				EXECUTE 'CREATE INDEX IF NOT EXISTS idx_job_devices_composite ON jobdevices(jobID, "deviceID")';
			ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobdevices' AND column_name='deviceid') THEN
				EXECUTE 'CREATE INDEX IF NOT EXISTS idx_job_devices_device_id ON jobdevices(deviceid)';
				EXECUTE 'CREATE INDEX IF NOT EXISTS idx_job_devices_composite ON jobdevices(jobID, deviceid)';
			ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobdevices' AND column_name='device_id') THEN
				EXECUTE 'CREATE INDEX IF NOT EXISTS idx_job_devices_device_id ON jobdevices(device_id)';
				EXECUTE 'CREATE INDEX IF NOT EXISTS idx_job_devices_composite ON jobdevices(jobID, device_id)';
			END IF;
		END IF;
	END IF;
END$$;

-- Financial transactions indexes (guarded)
DO $$
BEGIN
	IF to_regclass('public.financial_transactions') IS NOT NULL THEN
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='financial_transactions' AND column_name='status')
		   AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='financial_transactions' AND column_name='type') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_financial_transactions_status_type ON financial_transactions(status, type)';
		END IF;

		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='financial_transactions' AND column_name='transaction_date') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_financial_transactions_date ON financial_transactions(transaction_date)';
		END IF;

		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='financial_transactions' AND column_name='due_date') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_financial_transactions_due_date ON financial_transactions(due_date)';
		END IF;

		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='financial_transactions' AND column_name='customerID') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_financial_transactions_customer ON financial_transactions(customerID)';
		ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='financial_transactions' AND column_name='customer_id') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_financial_transactions_customer ON financial_transactions(customer_id)';
		END IF;
	ELSE
		RAISE NOTICE 'financial_transactions relation not found; skipping indexes';
	END IF;
END$$;

-- Invoices table indexes (guarded)
DO $$
BEGIN
	IF to_regclass('public.invoices') IS NOT NULL THEN
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='invoices' AND column_name='customerID') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_invoices_customer_id ON invoices(customerID)';
		ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='invoices' AND column_name='customer_id') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_invoices_customer_id ON invoices(customer_id)';
		END IF;

		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='invoices' AND column_name='status') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status)';
		END IF;

		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='invoices' AND column_name='invoice_date') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_invoices_date ON invoices(invoice_date)';
		ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='invoices' AND column_name='date') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_invoices_date ON invoices(date)';
		END IF;
	ELSE
		RAISE NOTICE 'invoices relation not found; skipping indexes';
	END IF;
END$$;

-- Sessions table cleanup index (guarded)
DO $$
BEGIN
	IF to_regclass('public.sessions') IS NOT NULL THEN
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='sessions' AND column_name='expires_at') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)';
		END IF;
	ELSE
		RAISE NOTICE 'sessions relation not found; skipping index';
	END IF;
END$$;

-- Composite indexes for common queries (guarded)
DO $$
BEGIN
	IF to_regclass('public.jobs') IS NOT NULL THEN
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='customerID')
		   AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='statusID') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_customer_status ON jobs(customerID, statusID)';
		ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='customer_id')
		   AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='status_id') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_customer_status ON jobs(customer_id, status_id)';
		END IF;
	ELSE
		RAISE NOTICE 'jobs relation not found; skipping composite index';
	END IF;

	IF to_regclass('public.devices') IS NOT NULL THEN
		IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='productID')
		   AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='status') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_devices_product_status ON devices(productID, status)';
		ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='productid')
		   AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='status') THEN
			EXECUTE 'CREATE INDEX IF NOT EXISTS idx_devices_product_status ON devices(productid, status)';
		END IF;
	ELSE
		RAISE NOTICE 'devices relation not found; skipping composite index';
	END IF;
END$$;