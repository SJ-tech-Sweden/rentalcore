-- ================================================================
-- Migration 025 (UP): Create user dashboard widget preferences table
-- ================================================================

BEGIN;

CREATE TABLE IF NOT EXISTS user_dashboard_widgets (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    widgets JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_dashboard_widgets_user FOREIGN KEY (user_id) REFERENCES users(userid) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_dashboard_widgets_user
    ON user_dashboard_widgets (user_id);

COMMIT;
