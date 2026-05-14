-- AUTO-CONVERTED (heuristic)
-- Source: 030_fix_job_edit_sessions_postgres.down.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Rollback PostgreSQL-specific fix for job_edit_sessions table

ALTER TABLE job_edit_sessions
  DROP CONSTRAINT IF EXISTS job_edit_sessions_job_user_unique;
