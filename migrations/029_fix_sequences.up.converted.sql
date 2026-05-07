-- AUTO-CONVERTED (heuristic)
-- Source: 029_fix_sequences.up.sql.bak
-- Review this file for correctness before applying to Postgres.

-- Migration: Fix PostgreSQL sequences to match current max IDs
-- This prevents "duplicate key value violates unique constraint" errors
-- when inserting new records into tables with auto-increment primary keys

DO $$
DECLARE
    r RECORD;
    max_val BIGINT;
    table_name TEXT;
    column_name TEXT;
BEGIN
    -- Loop through all sequences in the public schema
    FOR r IN
        SELECT
            s.schemaname,
            s.sequencename,
            -- Extract table name from sequence name (e.g., "products_productID_seq" -> "products")
            SPLIT_PART(s.sequencename, '_', 1) as base_table
        FROM pg_sequences s
        WHERE s.schemaname = 'public'
    LOOP
        BEGIN
            -- Try to find the actual column name that uses this sequence
            SELECT
                t.table_name,
                c.column_name
            INTO table_name, column_name
            FROM information_schema.columns c
            JOIN information_schema.tables t ON c.table_name = t.table_name
            WHERE t.table_schema = 'public'
            AND c.table_schema = 'public'
            AND pg_get_serial_sequence(t.table_name, c.column_name) = r.schemaname || '."' || r.sequencename || '"';

            IF table_name IS NOT NULL AND column_name IS NOT NULL THEN
                -- Get the current max value from the table
                EXECUTE format(
                    'SELECT COALESCE(MAX(%I), 0) FROM %I',
                    column_name,
                    table_name
                ) INTO max_val;

                IF max_val > 0 THEN
                    -- Update the sequence to start from max + 1
                    EXECUTE format('SELECT setval(%L, %s)', r.schemaname || '."' || r.sequencename || '"', max_val);
                    RAISE NOTICE 'Fixed sequence %.% to % (table: %, column: %)',
                        r.schemaname, r.sequencename, max_val, table_name, column_name;
                END IF;
            END IF;

            -- Reset variables for next iteration
            table_name := NULL;
            column_name := NULL;

        EXCEPTION
            WHEN OTHERS THEN
                -- Skip if there's any error (table doesn't exist, column doesn't exist, etc.)
                RAISE NOTICE 'Skipping sequence % due to error: %', r.sequencename, SQLERRM;
        END;
    END LOOP;

    RAISE NOTICE 'Sequence synchronization completed successfully';
END$$;
