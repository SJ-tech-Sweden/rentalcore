-- Create minimal `jobcategory` compatibility table used by legacy queries
-- Idempotent: will not overwrite an existing table and seeds default categories if missing.
CREATE TABLE IF NOT EXISTS jobcategory (
  jobcategoryid SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  abbreviation VARCHAR(50),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed a default category if none exist
DO $$
BEGIN
  IF (SELECT COUNT(*) FROM jobcategory) = 0 THEN
    INSERT INTO jobcategory (name, abbreviation) VALUES
      ('General', 'GEN'),
      ('Installation', 'INST'),
      ('Maintenance', 'MAIN')
    ;
  END IF;
END$$ LANGUAGE plpgsql;
