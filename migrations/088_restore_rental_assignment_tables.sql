-- 088_restore_rental_assignment_tables.sql
-- Recreate rental assignment tables for RentalCore job form persistence.
-- These tables may have been removed by legacy cleanup migrations.

CREATE TABLE IF NOT EXISTS rental_equipment (
    equipment_id BIGINT PRIMARY KEY,
    product_name TEXT NOT NULL,
    supplier_name TEXT NOT NULL,
    rental_price DOUBLE PRECISION NOT NULL DEFAULT 0,
    category TEXT,
    description TEXT,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    created_by INT
);

CREATE TABLE IF NOT EXISTS job_rental_equipment (
    job_id INT NOT NULL,
    equipment_id BIGINT NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    days_used INT NOT NULL DEFAULT 1,
    total_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    PRIMARY KEY (job_id, equipment_id),
    CONSTRAINT fk_jre_job FOREIGN KEY (job_id) REFERENCES jobs(jobid) ON DELETE CASCADE,
    CONSTRAINT fk_jre_equipment FOREIGN KEY (equipment_id) REFERENCES rental_equipment(equipment_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_job_rental_equipment_job_id ON job_rental_equipment(job_id);
CREATE INDEX IF NOT EXISTS idx_job_rental_equipment_equipment_id ON job_rental_equipment(equipment_id);
CREATE INDEX IF NOT EXISTS idx_rental_equipment_category ON rental_equipment(category);
