-- Use a PL/pgSQL DO block to perform conditional DDL in Postgres
DO $$
BEGIN
    -- If sessions table exists and has no primary key, add primary key on session_id
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'sessions') THEN
        IF NOT EXISTS (
            SELECT 1 FROM pg_constraint c
            JOIN pg_class t ON c.conrelid = t.oid
            WHERE t.relname = 'sessions' AND c.contype = 'p'
        ) THEN
            EXECUTE 'ALTER TABLE sessions ADD PRIMARY KEY (session_id)';
        END IF;
    END IF;

    -- If constraint fk_sessions_user exists on users table, drop it
    IF EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON c.conrelid = t.oid
        WHERE c.conname = 'fk_sessions_user' AND t.relname = 'users'
    ) THEN
        EXECUTE 'ALTER TABLE users DROP CONSTRAINT fk_sessions_user';
    END IF;
END$$;
