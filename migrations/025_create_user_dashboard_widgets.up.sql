-- ================================================================
-- Migration 025 (UP): Create user dashboard widget preferences table
-- ================================================================

START TRANSACTION;

CREATE TABLE IF NOT EXISTS user_dashboard_widgets (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    widgets JSON NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_dashboard_widgets_user
        FOREIGN KEY (user_id) REFERENCES users(userID)
        ON DELETE CASCADE ON UPDATE CASCADE,
    UNIQUE KEY idx_user_dashboard_widgets_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

COMMIT;

