-- AUTO-CONVERTED (heuristic)
-- Source: 026_add_job_edit_sessions.up.sql.bak
-- Review this file for correctness before applying to Postgres.

CREATE TABLE IF NOT EXISTS job_edit_sessions (
    session_id BIGSERIAL PRIMARY KEY,
    job_id INT NOT NULL,
    user_id BIGINT NOT NULL,
    username VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_job_edit_sessions_job FOREIGN KEY (job_id) REFERENCES jobs(jobid) ON DELETE CASCADE,
    CONSTRAINT fk_job_edit_sessions_user FOREIGN KEY (user_id) REFERENCES users(userid) ON DELETE CASCADE,
    CONSTRAINT uk_job_edit_sessions_job_user UNIQUE (job_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_job_edit_sessions_job ON job_edit_sessions(job_id);
CREATE INDEX IF NOT EXISTS idx_job_edit_sessions_user ON job_edit_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_job_edit_sessions_last_seen ON job_edit_sessions(last_seen);
