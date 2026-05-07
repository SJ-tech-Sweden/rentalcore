-- AUTO-CONVERTED (heuristic)
-- Source: 021_create_rental_equipment_tables_corrected.up.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Migration: Create corrected tables for rental equipment system
-- This creates tables that match the Go models exactly

-- Table for rental equipment items (master list of available rental items)
CREATE TABLE IF NOT EXISTS rental_equipment (
    equipment_id SERIAL PRIMARY KEY,
    product_name VARCHAR(200) NOT NULL,
    supplier_name VARCHAR(100) NOT NULL,
    rental_price DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    category VARCHAR(50),
    description VARCHAR(1000),
    notes VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INT
);

CREATE INDEX IF NOT EXISTS idx_product_name ON rental_equipment(product_name);
CREATE INDEX IF NOT EXISTS idx_supplier_name ON rental_equipment(supplier_name);
CREATE INDEX IF NOT EXISTS idx_category ON rental_equipment(category);
CREATE INDEX IF NOT EXISTS idx_is_active ON rental_equipment(is_active);

-- Bridge table for job-rental equipment assignments
CREATE TABLE IF NOT EXISTS job_rental_equipment (
    job_id INT NOT NULL,
    equipment_id INT  NOT NULL,
    quantity INT  NOT NULL DEFAULT 1,
    days_used INT  NOT NULL DEFAULT 1,
    total_cost DECIMAL(12, 2) NOT NULL,
    notes VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (job_id, equipment_id),
    FOREIGN KEY (job_id) REFERENCES jobs(jobid) ON DELETE CASCADE,
    FOREIGN KEY (equipment_id) REFERENCES rental_equipment(equipment_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_job_id ON job_rental_equipment(job_id);
CREATE INDEX IF NOT EXISTS idx_equipment_id ON job_rental_equipment(equipment_id);
CREATE INDEX IF NOT EXISTS idx_created_at ON job_rental_equipment(created_at);

-- Insert some example rental equipment items
INSERT INTO rental_equipment (product_name, supplier_name, rental_price, description, category) VALUES
('LED Moving Head - Martin MAC Aura', 'Pro Rental GmbH', 45.00, 'Professional LED Moving Head Light', 'Lighting'),
('d&b V12 Line Array', 'Sound Solutions AG', 120.00, 'High-end Line Array Speaker System', 'Audio'),
('Truss System 3m Segment', 'Stage Tech Berlin', 15.00, '3 Meter Aluminum Truss Segment', 'Stage Equipment'),
('Haze Machine - Unique 2.1', 'Effect Masters', 35.00, 'Professional Haze Machine with DMX', 'Other'),
('LED Par 64 RGBW', 'Light Rental Pro', 12.00, 'RGBW LED Par with DMX Control', 'Lighting'),
('Wireless Microphone Shure ULXD2', 'Audio Rent Hamburg', 25.00, 'Professional Wireless Handheld Microphone', 'Audio');