-- AUTO-CONVERTED (heuristic)
-- Source: 001_initial_schema.sql.bak
-- Review this file for correctness before applying to Postgres.

-- MySQL original initial schema moved to .bak to avoid syntax errors in Postgres migration runner
-- Original file contained CREATE DATABASE, USE, MySQL-specific index and engine clauses.
-- This repository contains a Postgres-friendly 000_initial.sql which should be used
-- for new Postgres deployments. Keep this backup for reference when needed.
-- JobScanner Pro - Initial Database Schema

-- Removed MySQL-specific database creation/selection statements (not valid in Postgres migration runner)
-- CREATE DATABASE IF NOT EXISTS jobscanner CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- USE jobscanner;

CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='customers' AND column_name='name') THEN
        EXECUTE 'CREATE INDEX IF NOT EXISTS idx_customers_name ON customers(name)';
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='customers' AND column_name='email') THEN
        EXECUTE 'CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email)';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Job-Device relationship table (idempotent creation)
DO $$
BEGIN
    IF to_regclass('public.job_devices') IS NULL THEN
        CREATE TABLE job_devices (
                id SERIAL PRIMARY KEY,
                job_id INT NOT NULL,
                device_id INT NOT NULL,
                assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                removed_at TIMESTAMP NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
                FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
        );
        IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'idx_job') THEN
            CREATE INDEX idx_job ON job_devices(job_id);
        END IF;
        IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'idx_device') THEN
            CREATE INDEX idx_device ON job_devices(device_id);
        END IF;
        IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'idx_assigned') THEN
            CREATE INDEX idx_assigned ON job_devices(assigned_at);
        END IF;
        IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'idx_removed') THEN
            CREATE INDEX idx_removed ON job_devices(removed_at);
        END IF;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Insert default statuses
INSERT INTO statuses (name, description, color) VALUES
('Planning', 'Job is in planning phase', '#6c757d'),
('Active', 'Job is currently active', '#28a745'),
('Completed', 'Job has been completed', '#007bff'),
('Cancelled', 'Job has been cancelled', '#dc3545'),
('On Hold', 'Job is temporarily on hold', '#ffc107')
ON CONFLICT (name) DO NOTHING;

-- Sample data removed for production use
-- To add initial data, use the application interface or create separate data import scripts

-- Create views for commonly used queries
DO $$
DECLARE
    cust_col TEXT;
    stat_col TEXT;
    cust_pk TEXT;
    stat_pk TEXT;
    stat_col_type TEXT;
    cust_name_col TEXT;
    stat_name_col TEXT;
    stat_color_col TEXT;
    jd_job_col TEXT;
    jd_dev_col TEXT;
    jd_removed_col TEXT;
    job_pk TEXT;
    jd_price_col TEXT;
    sql TEXT;
    -- optional job columns detection
    has_title BOOL;
    has_description BOOL;
    has_start_date BOOL;
    has_end_date BOOL;
    has_revenue BOOL;
    has_created_at BOOL;
    has_updated_at BOOL;
    select_fields TEXT;
    group_fields TEXT;
    has_deleted_at BOOL;
    where_clause TEXT;
BEGIN
    -- detect jobs.customer column and customers PK column (support customer_id, customerid, customer)
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='customer_id') THEN
        cust_col := 'customer_id';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='customerid') THEN
        cust_col := 'customerid';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='customer') THEN
        cust_col := 'customer';
    ELSE
        cust_col := 'customerid';
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='customers' AND column_name='id') THEN
        cust_pk := 'id';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='customers' AND column_name='customerid') THEN
        cust_pk := 'customerid';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='customers' AND column_name='customer_id') THEN
        cust_pk := 'customer_id';
    ELSE
        cust_pk := 'id';
    END IF;

    -- detect jobs.status column and statuses PK column (support status_id, statusid, status)
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='status_id') THEN
        stat_col := 'status_id';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='statusid') THEN
        stat_col := 'statusid';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='status') THEN
        stat_col := 'status';
    ELSE
        stat_col := 'statusid';
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='statuses' AND column_name='id') THEN
        stat_pk := 'id';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='statuses' AND column_name='statusid') THEN
        stat_pk := 'statusid';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='statuses' AND column_name='status_id') THEN
        stat_pk := 'status_id';
    ELSE
        stat_pk := 'id';
    END IF;

    -- detect customers name column variants
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='customers' AND column_name='name') THEN
        cust_name_col := 'name';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='customers' AND column_name='customer_name') THEN
        cust_name_col := 'customer_name';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='customers' AND column_name='fullname') THEN
        cust_name_col := 'fullname';
    ELSE
        cust_name_col := '';
    END IF;

    -- detect statuses name/color column variants
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='statuses' AND column_name='name') THEN
        stat_name_col := 'name';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='statuses' AND column_name='status_name') THEN
        stat_name_col := 'status_name';
    ELSE
        stat_name_col := '';
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='statuses' AND column_name='color') THEN
        stat_color_col := 'color';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='statuses' AND column_name='colour') THEN
        stat_color_col := 'colour';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='statuses' AND column_name='status_color') THEN
        stat_color_col := 'status_color';
    ELSE
        stat_color_col := '';
    END IF;

    -- detect job_devices job and device columns
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='job_id') THEN
        jd_job_col := 'job_id';
    ELSE
        jd_job_col := 'jobid';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='device_id') THEN
        jd_dev_col := 'device_id';
    ELSE
        jd_dev_col := 'deviceid';
    END IF;

    -- detect job_devices price column name
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='price') THEN
        jd_price_col := 'price';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='custom_price') THEN
        jd_price_col := 'custom_price';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='customprice') THEN
        jd_price_col := 'customprice';
    ELSE
        jd_price_col := '';
    END IF;

    -- detect job_devices removed column name
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='removed_at') THEN
        jd_removed_col := 'removed_at';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='removedat') THEN
        jd_removed_col := 'removedat';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='removed') THEN
        jd_removed_col := 'removed';
    ELSE
        jd_removed_col := '';
    END IF;

    -- build job_summary view using detected names
    -- detect jobs primary key column
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='id') THEN
        job_pk := 'id';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='jobid') THEN
        job_pk := 'jobid';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='job_id') THEN
        job_pk := 'job_id';
    ELSE
        job_pk := 'id';
    END IF;

    -- detect data type of jobs.status column
    SELECT data_type INTO stat_col_type FROM information_schema.columns WHERE table_name='jobs' AND column_name=stat_col LIMIT 1;

    -- detect optional job columns
    has_title := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='title');
    has_description := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='description');
    has_start_date := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='start_date');
    has_end_date := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='end_date');
    has_revenue := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='revenue');
    has_created_at := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='created_at');
    has_updated_at := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='updated_at');
    select_fields := '';
    group_fields := '';

    -- build SQL using concatenation (dynamic optional fields)
    select_fields := 'j.' || job_pk || ' AS id';
    IF has_title THEN
        select_fields := select_fields || ', j.title';
        group_fields := group_fields || ', j.title';
    ELSE
        select_fields := select_fields || ', NULL::text as title';
    END IF;
    IF has_description THEN
        select_fields := select_fields || ', j.description';
        group_fields := group_fields || ', j.description';
    ELSE
        select_fields := select_fields || ', NULL::text as description';
    END IF;
    IF has_start_date THEN
        select_fields := select_fields || ', j.start_date';
        group_fields := group_fields || ', j.start_date';
    ELSE
        select_fields := select_fields || ', NULL::date as start_date';
    END IF;
    IF has_end_date THEN
        select_fields := select_fields || ', j.end_date';
        group_fields := group_fields || ', j.end_date';
    ELSE
        select_fields := select_fields || ', NULL::date as end_date';
    END IF;
    IF has_revenue THEN
        select_fields := select_fields || ', j.revenue';
        group_fields := group_fields || ', j.revenue';
    ELSE
        select_fields := select_fields || ', 0::numeric as revenue';
    END IF;
    IF has_created_at THEN
        select_fields := select_fields || ', j.created_at';
        group_fields := group_fields || ', j.created_at';
    ELSE
        select_fields := select_fields || ', NULL::timestamp as created_at';
    END IF;
    IF has_updated_at THEN
        select_fields := select_fields || ', j.updated_at';
        group_fields := group_fields || ', j.updated_at';
    ELSE
        select_fields := select_fields || ', NULL::timestamp as updated_at';
    END IF;

    -- always include customer/status and aggregates (use detected name/color columns)
    IF cust_name_col <> '' THEN
        select_fields := select_fields || ', c.' || cust_name_col || ' as customer_name';
    ELSE
        select_fields := select_fields || ', NULL::text as customer_name';
    END IF;
    IF stat_name_col <> '' THEN
        select_fields := select_fields || ', s.' || stat_name_col || ' as status_name';
    ELSE
        select_fields := select_fields || ', NULL::text as status_name';
    END IF;
    IF stat_color_col <> '' THEN
        select_fields := select_fields || ', s.' || stat_color_col || ' as status_color';
    ELSE
        select_fields := select_fields || ', NULL::text as status_color';
    END IF;
    select_fields := select_fields || ', COUNT(DISTINCT jd.' || jd_dev_col || ') as device_count';
    IF jd_price_col <> '' THEN
        select_fields := select_fields || ', COALESCE(SUM(jd.' || jd_price_col || '), 0) as total_device_revenue';
    ELSE
        select_fields := select_fields || ', 0::numeric as total_device_revenue';
    END IF;

    -- build group suffix for customer/status columns only when present
    group_fields := group_fields || '';
    IF cust_name_col <> '' THEN
        group_fields := group_fields || ', c.' || cust_name_col;
    END IF;
    IF stat_name_col <> '' THEN
        group_fields := group_fields || ', s.' || stat_name_col;
    END IF;
    IF stat_color_col <> '' THEN
        group_fields := group_fields || ', s.' || stat_color_col;
    END IF;

    sql := 'CREATE OR REPLACE VIEW job_summary AS SELECT ' || select_fields || ' FROM jobs j ' ||
           'LEFT JOIN customers c ON j.' || cust_col || ' = c.' || cust_pk || ' ';

    IF stat_col_type LIKE 'character%' OR stat_col_type IS NULL THEN
        IF stat_name_col <> '' THEN
            sql := sql || 'LEFT JOIN statuses s ON j.' || stat_col || ' = s.' || stat_name_col || ' ';
        ELSE
            sql := sql || 'LEFT JOIN statuses s ON j.' || stat_col || '::text = s.' || stat_pk || '::text ';
        END IF;
    ELSE
        sql := sql || 'LEFT JOIN statuses s ON j.' || stat_col || '::text = s.' || stat_pk || '::text ';
    END IF;

    has_deleted_at := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='deleted_at');
    IF has_deleted_at THEN
        where_clause := 'WHERE j.deleted_at IS NULL ';
    ELSE
        where_clause := '';
    END IF;

    IF jd_removed_col <> '' THEN
        sql := sql || 'LEFT JOIN job_devices jd ON j.' || job_pk || ' = jd.' || jd_job_col || ' AND (jd.' || jd_removed_col || ' IS NULL) ' ||
                     where_clause ||
                     'GROUP BY j.' || job_pk || group_fields || ';';
    ELSE
        sql := sql || 'LEFT JOIN job_devices jd ON j.' || job_pk || ' = jd.' || jd_job_col || ' ' ||
                     where_clause ||
                     'GROUP BY j.' || job_pk || group_fields || ';';
    END IF;

    EXECUTE sql;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    jd_job_col TEXT;
    jd_dev_col TEXT;
    jd_removed_col TEXT;
    job_pk TEXT;
    sql2 TEXT;
    has_deleted_at BOOL;
    device_pk TEXT;
    has_serial BOOL;
    has_dev_name BOOL;
    has_dev_description BOOL;
    has_dev_category BOOL;
    has_dev_price BOOL;
    has_dev_available BOOL;
    has_dev_created_at BOOL;
    has_dev_updated_at BOOL;
    has_job_title_local BOOL;
    select_fields_dev TEXT;
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='job_id') THEN
        jd_job_col := 'job_id';
    ELSE
        jd_job_col := 'jobid';
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='device_id') THEN
        jd_dev_col := 'device_id';
    ELSE
        jd_dev_col := 'deviceid';
    END IF;

    -- detect devices primary key column
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='id') THEN
        device_pk := 'id';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='deviceid') THEN
        device_pk := 'deviceid';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='device_id') THEN
        device_pk := 'device_id';
    ELSE
        device_pk := 'id';
    END IF;

    -- detect device columns
    has_serial := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='serial_no');
    has_dev_name := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='name');
    has_dev_description := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='description');
    has_dev_category := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='category');
    has_dev_price := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='price');
    has_dev_available := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='available');
    has_dev_created_at := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='created_at');
    has_dev_updated_at := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='devices' AND column_name='updated_at');

    -- detect jobs.title presence for current_job_title
    has_job_title_local := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='title');

    -- detect job_devices removed column name
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='removed_at') THEN
        jd_removed_col := 'removed_at';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='removedat') THEN
        jd_removed_col := 'removedat';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_devices' AND column_name='removed') THEN
        jd_removed_col := 'removed';
    ELSE
        jd_removed_col := '';
    END IF;

    -- detect jobs primary key column for joining
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='id') THEN
        job_pk := 'id';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='jobid') THEN
        job_pk := 'jobid';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='job_id') THEN
        job_pk := 'job_id';
    ELSE
        job_pk := 'id';
    END IF;

    -- build device_status view; include removed filter only if column detected
    has_deleted_at := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='jobs' AND column_name='deleted_at');

    -- build select list dynamically based on existing device columns
    select_fields_dev := 'd.' || device_pk || ' AS id';
    IF has_serial THEN
        select_fields_dev := select_fields_dev || ', d.serial_no';
    ELSE
        select_fields_dev := select_fields_dev || ', NULL::text as serial_no';
    END IF;
    IF has_dev_name THEN
        select_fields_dev := select_fields_dev || ', d.name';
    ELSE
        select_fields_dev := select_fields_dev || ', NULL::text as name';
    END IF;
    IF has_dev_description THEN
        select_fields_dev := select_fields_dev || ', d.description';
    ELSE
        select_fields_dev := select_fields_dev || ', NULL::text as description';
    END IF;
    IF has_dev_category THEN
        select_fields_dev := select_fields_dev || ', d.category';
    ELSE
        select_fields_dev := select_fields_dev || ', NULL::text as category';
    END IF;
    IF has_dev_price THEN
        select_fields_dev := select_fields_dev || ', d.price';
    ELSE
        select_fields_dev := select_fields_dev || ', 0::numeric as price';
    END IF;
    IF has_dev_available THEN
        select_fields_dev := select_fields_dev || ', d.available';
    ELSE
        select_fields_dev := select_fields_dev || ', TRUE as available';
    END IF;
    IF has_dev_created_at THEN
        select_fields_dev := select_fields_dev || ', d.created_at';
    ELSE
        select_fields_dev := select_fields_dev || ', NULL::timestamp as created_at';
    END IF;
    IF has_dev_updated_at THEN
        select_fields_dev := select_fields_dev || ', d.updated_at';
    ELSE
        select_fields_dev := select_fields_dev || ', NULL::timestamp as updated_at';
    END IF;

    -- always include is_free/current_job_id and job title if available
    select_fields_dev := select_fields_dev || ', CASE WHEN jd.' || jd_job_col || ' IS NOT NULL THEN FALSE ELSE TRUE END as is_free, jd.' || jd_job_col || ' as current_job_id';
    IF has_job_title_local THEN
        select_fields_dev := select_fields_dev || ', j.title as current_job_title';
    ELSE
        select_fields_dev := select_fields_dev || ', NULL::text as current_job_title';
    END IF;

    sql2 := 'CREATE OR REPLACE VIEW device_status AS SELECT ' || select_fields_dev || ' FROM devices d';
    IF jd_removed_col <> '' THEN
        sql2 := sql2 || format(' LEFT JOIN job_devices jd ON d.%I = jd.%I AND jd.%I IS NULL', device_pk, jd_dev_col, jd_removed_col);
    ELSE
        sql2 := sql2 || format(' LEFT JOIN job_devices jd ON d.%I = jd.%I', device_pk, jd_dev_col);
    END IF;

    IF has_deleted_at THEN
        sql2 := sql2 || format(' LEFT JOIN jobs j ON jd.%I = j.%I AND j.deleted_at IS NULL', jd_job_col, job_pk);
    ELSE
        sql2 := sql2 || format(' LEFT JOIN jobs j ON jd.%I = j.%I', jd_job_col, job_pk);
    END IF;

    EXECUTE sql2;
END;
$$ LANGUAGE plpgsql;