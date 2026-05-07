-- AUTO-CONVERTED (heuristic)
-- Source: 025_create_user_dashboard_widgets.down.sql.bak
-- Review this file for correctness before applying to Postgres.

-- ================================================================
-- Migration 025 (DOWN): Drop user dashboard widget preferences table
-- ================================================================

START TRANSACTION;

DROP TABLE IF EXISTS user_dashboard_widgets;

COMMIT;

