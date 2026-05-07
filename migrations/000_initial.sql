-- Initial consolidated schema for RentalCore (minimal, idempotent)
-- Creates core tables and compatibility views/triggers needed by the app

-- Roles
CREATE TABLE IF NOT EXISTS roles (
  roleid SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  display_name VARCHAR(150),
  description TEXT,
  scope VARCHAR(50) DEFAULT 'rentalcore',
  is_system_role BOOLEAN DEFAULT FALSE,
  is_active BOOLEAN DEFAULT TRUE,
  permissions JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Users
CREATE TABLE IF NOT EXISTS users (
  userid SERIAL PRIMARY KEY,
  username VARCHAR(100) NOT NULL UNIQUE,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  is_admin BOOLEAN DEFAULT FALSE,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- User roles
CREATE TABLE IF NOT EXISTS user_roles (
  id SERIAL PRIMARY KEY,
  userid INT NOT NULL REFERENCES users(userid) ON DELETE CASCADE,
  roleid INT NOT NULL REFERENCES roles(roleid) ON DELETE CASCADE,
  assigned_by INT REFERENCES users(userid) ON DELETE SET NULL,
  assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  is_active BOOLEAN DEFAULT TRUE,
  UNIQUE(userid, roleid)
);

-- Sessions (simple)
CREATE TABLE IF NOT EXISTS sessions (
  session_id VARCHAR(255) PRIMARY KEY,
  user_id INT NOT NULL REFERENCES users(userid) ON DELETE CASCADE,
  data TEXT,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- App settings
CREATE TABLE IF NOT EXISTS app_settings (
  id SERIAL PRIMARY KEY,
  scope VARCHAR(50) NOT NULL DEFAULT 'rentalcore',
  key VARCHAR(100) NOT NULL,
  value TEXT,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(scope, key)
);

-- Products
CREATE TABLE IF NOT EXISTS products (
  productID SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  categoryID INT,
  manufacturerID INT,
  description TEXT,
  price_per_unit NUMERIC,
  stock_quantity NUMERIC,
  website_visible BOOLEAN DEFAULT FALSE
);

-- Customers
CREATE TABLE IF NOT EXISTS customers (
  customerID SERIAL PRIMARY KEY,
  companyname TEXT,
  firstname TEXT,
  lastname TEXT,
  email TEXT,
  phonenumber TEXT,
  billing_address TEXT,
  shipping_address TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Jobs (minimal)
CREATE TABLE IF NOT EXISTS jobs (
  jobid SERIAL PRIMARY KEY,
  name TEXT,
  customerID INT REFERENCES customers(customerID),
  start_date TIMESTAMP,
  end_date TIMESTAMP,
  status VARCHAR(50) DEFAULT 'open',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Job devices (associations) - canonical snake_case table
CREATE TABLE IF NOT EXISTS job_devices (
  id SERIAL PRIMARY KEY,
  jobid INT NOT NULL REFERENCES jobs(jobid) ON DELETE CASCADE,
  device_id VARCHAR(255) NOT NULL,
  productid INT,
  quantity INT DEFAULT 1
);

-- Create compatibility view `jobdevices` and INSTEAD OF triggers
-- Simple compatibility view `jobdevices` (read-only). Triggers omitted for now.
CREATE OR REPLACE VIEW jobdevices AS
  SELECT jobid, device_id AS deviceid, NULL::numeric AS custom_price, NULL::int AS package_id,
    false AS is_package_item, 'pending'::varchar AS pack_status, NULL::timestamp AS pack_ts
  FROM job_devices;

-- Job history and related tables
CREATE TABLE IF NOT EXISTS job_history (
  history_id SERIAL PRIMARY KEY,
  job_id INT NOT NULL REFERENCES jobs(jobid) ON DELETE CASCADE,
  user_id INT DEFAULT NULL REFERENCES users(userid) ON DELETE SET NULL,
  changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  change_type VARCHAR(100) NOT NULL,
  field_name VARCHAR(255) DEFAULT NULL,
  old_value TEXT,
  new_value TEXT,
  description TEXT,
  ip_address VARCHAR(45) DEFAULT NULL,
  user_agent TEXT DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS idx_job_history_job ON job_history(job_id);
CREATE INDEX IF NOT EXISTS idx_job_history_user ON job_history(user_id);
CREATE INDEX IF NOT EXISTS idx_job_history_changed_at ON job_history(changed_at);

CREATE TABLE IF NOT EXISTS job_edit_sessions (
  job_id INT NOT NULL REFERENCES jobs(jobid) ON DELETE CASCADE,
  user_id INT NOT NULL REFERENCES users(userid) ON DELETE CASCADE,
  username VARCHAR(255) NOT NULL,
  display_name VARCHAR(255) NOT NULL,
  started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (job_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_job_edit_sessions_last_seen ON job_edit_sessions(last_seen);

CREATE TABLE IF NOT EXISTS job_packages (
  job_package_id SERIAL PRIMARY KEY,
  job_id INT NOT NULL REFERENCES jobs(jobid) ON DELETE CASCADE,
  package_id INT NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  custom_price DECIMAL(12,2) DEFAULT NULL,
  added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  added_by INT DEFAULT NULL REFERENCES users(userid) ON DELETE SET NULL,
  notes TEXT
);
CREATE INDEX IF NOT EXISTS idx_job_packages_job ON job_packages(job_id);
CREATE INDEX IF NOT EXISTS idx_job_packages_added_at ON job_packages(added_at);

-- End of minimal initial schema
