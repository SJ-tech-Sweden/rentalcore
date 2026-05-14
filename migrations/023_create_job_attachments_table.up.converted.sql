-- AUTO-CONVERTED (heuristic)
-- Source: 023_create_job_attachments_table.up.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Create job_attachments table for file attachments to jobs
CREATE TABLE IF NOT EXISTS job_attachments (
    attachment_id BIGSERIAL PRIMARY KEY,
    job_id INT NOT NULL,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    uploaded_by BIGINT  DEFAULT NULL,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (job_id) REFERENCES jobs(jobid) ON DELETE CASCADE,
    FOREIGN KEY (uploaded_by) REFERENCES users(userid) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_job_attachments_job_id ON job_attachments(job_id);
CREATE INDEX IF NOT EXISTS idx_job_attachments_uploaded_at ON job_attachments(uploaded_at);