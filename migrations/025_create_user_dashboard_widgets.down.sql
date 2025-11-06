-- ================================================================
-- Migration 025 (DOWN): Drop user dashboard widget preferences table
-- ================================================================

START TRANSACTION;

DROP TABLE IF EXISTS user_dashboard_widgets;

COMMIT;

