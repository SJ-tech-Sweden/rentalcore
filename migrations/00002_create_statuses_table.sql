-- Idempotent migration to ensure `statuses` table exists
CREATE TABLE IF NOT EXISTS statuses (
  id SERIAL PRIMARY KEY,
  name VARCHAR(150) NOT NULL UNIQUE,
  description TEXT,
  color VARCHAR(50),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ensure a simple index on name for lookups
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'idx_statuses_name') THEN
    CREATE INDEX idx_statuses_name ON statuses(name);
  END IF;
END;
$$ LANGUAGE plpgsql;

-- Optional: seed common default statuses if not present
INSERT INTO statuses (name, description, color)
VALUES
  ('Planning', 'Job is in planning phase', '#6c757d'),
  ('Active', 'Job is currently active', '#28a745'),
  ('Completed', 'Job has been completed', '#007bff'),
  ('Cancelled', 'Job has been cancelled', '#dc3545'),
  ('On Hold', 'Job is temporarily on hold', '#ffc107')
ON CONFLICT (name) DO NOTHING;
