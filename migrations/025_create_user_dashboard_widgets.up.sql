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

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_index i
        JOIN pg_class t ON t.oid = i.indrelid
        JOIN pg_namespace n ON n.oid = t.relnamespace
        JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(i.indkey)
        WHERE n.nspname = 'public'
          AND t.relname = 'user_dashboard_widgets'
          AND i.indisunique
          AND i.indnkeyatts = 1
          AND a.attname = 'user_id'
    ) THEN
        CREATE UNIQUE INDEX uq_user_dashboard_widgets_user
            ON user_dashboard_widgets (user_id);
    END IF;
END$$;

COMMIT;
