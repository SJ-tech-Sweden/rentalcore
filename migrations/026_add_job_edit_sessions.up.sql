CREATE TABLE IF NOT EXISTS job_edit_sessions (
    session_id BIGSERIAL PRIMARY KEY,
    job_id INT NOT NULL,
    user_id BIGINT NOT NULL,
    username VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    last_seen TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_job_edit_sessions_job FOREIGN KEY (job_id) REFERENCES jobs(jobid) ON DELETE CASCADE,
    CONSTRAINT fk_job_edit_sessions_user FOREIGN KEY (user_id) REFERENCES users(userid) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_job_edit_sessions_job_user ON job_edit_sessions(job_id, user_id);
CREATE INDEX IF NOT EXISTS idx_job_edit_sessions_job ON job_edit_sessions(job_id);
CREATE INDEX IF NOT EXISTS idx_job_edit_sessions_user ON job_edit_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_job_edit_sessions_last_seen ON job_edit_sessions(last_seen);
