-- AUTO-CONVERTED (heuristic)
-- Source: 029_fix_sequences.down.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Migration rollback: Fix PostgreSQL sequences
-- This is a data-fixing migration, so there's no meaningful rollback
-- The sequences will remain at their corrected values

DO $$
BEGIN
    RAISE NOTICE 'No rollback needed for sequence synchronization migration';
    RAISE NOTICE 'Sequences remain at their current values';
END$$;
