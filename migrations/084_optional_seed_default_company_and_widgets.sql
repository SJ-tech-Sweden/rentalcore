-- 084_optional_seed_default_company_and_widgets.sql
-- Optional: seed default rows so startup "record not found" logs disappear.
-- Safe and idempotent across schema variants.

DO $$
DECLARE
    row_count bigint;
    cols text := '';
    vals text := '';
BEGIN
    IF to_regclass('public.company_settings') IS NULL THEN
        RAISE NOTICE 'company_settings table not found; skipping default seed';
        RETURN;
    END IF;

    EXECUTE 'SELECT COUNT(*) FROM company_settings' INTO row_count;
    IF row_count > 0 THEN
        RETURN;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'company_settings' AND column_name = 'company_name'
    ) THEN
        cols := cols || 'company_name';
        vals := vals || quote_literal('RentalCore');
    ELSIF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'company_settings' AND column_name = 'name'
    ) THEN
        cols := cols || 'name';
        vals := vals || quote_literal('RentalCore');
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'company_settings' AND column_name = 'updated_at'
    ) THEN
        IF cols <> '' THEN
            cols := cols || ',updated_at';
            vals := vals || ',NOW()';
        END IF;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'company_settings' AND column_name = 'created_at'
    ) THEN
        IF cols <> '' THEN
            cols := cols || ',created_at';
            vals := vals || ',NOW()';
        END IF;
    END IF;

    IF cols = '' THEN
        EXECUTE 'INSERT INTO company_settings DEFAULT VALUES';
    ELSE
        EXECUTE 'INSERT INTO company_settings (' || cols || ') VALUES (' || vals || ')';
    END IF;
END$$;

DO $$
DECLARE
    user_pk text;
    widget_col text;
    sql_stmt text;
    col_list text;
    select_list text;
    default_widgets text := '["total_customers","total_devices","active_jobs","total_cases"]';
BEGIN
    IF to_regclass('public.user_dashboard_widgets') IS NULL OR to_regclass('public.users') IS NULL THEN
        RAISE NOTICE 'user_dashboard_widgets or users table missing; skipping widget defaults seed';
        RETURN;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'user_dashboard_widgets' AND column_name = 'user_id'
    ) THEN
        RAISE NOTICE 'user_dashboard_widgets.user_id missing; skipping widget defaults seed';
        RETURN;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'userid'
    ) THEN
        user_pk := 'userid';
    ELSIF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'id'
    ) THEN
        user_pk := 'id';
    ELSIF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'user_id'
    ) THEN
        user_pk := 'user_id';
    ELSE
        RAISE NOTICE 'users PK column not detected; skipping widget defaults seed';
        RETURN;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'user_dashboard_widgets' AND column_name = 'widgets'
    ) THEN
        widget_col := 'widgets';
    ELSIF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'user_dashboard_widgets' AND column_name = 'widget'
    ) THEN
        widget_col := 'widget';
    ELSE
        RAISE NOTICE 'No widgets/widget JSON column found on user_dashboard_widgets; skipping';
        RETURN;
    END IF;

    col_list := format('user_id,%I', widget_col);
    select_list := format('u.%I,%L::jsonb', user_pk, default_widgets);

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'user_dashboard_widgets' AND column_name = 'created_at'
    ) THEN
        col_list := col_list || ',created_at';
        select_list := select_list || ',NOW()';
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'user_dashboard_widgets' AND column_name = 'updated_at'
    ) THEN
        col_list := col_list || ',updated_at';
        select_list := select_list || ',NOW()';
    END IF;

    sql_stmt := format(
        'INSERT INTO user_dashboard_widgets (%s) ' ||
        'SELECT %s FROM users u ' ||
        'WHERE NOT EXISTS (' ||
        '  SELECT 1 FROM user_dashboard_widgets w WHERE w.user_id = u.%I' ||
        ')',
        col_list,
        select_list,
        user_pk,
        user_pk
    );

    EXECUTE sql_stmt;
END$$;
