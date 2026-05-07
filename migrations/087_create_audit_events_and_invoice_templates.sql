-- 087_create_audit_events_and_invoice_templates.sql
-- Compatibility migration for missing compliance and invoicing tables.

DO $$
BEGIN
    -- ------------------------------------------------------------------
    -- audit_events (GoBD compliance)
    -- ------------------------------------------------------------------
    IF to_regclass('public.audit_events') IS NULL THEN
        CREATE TABLE public.audit_events (
            id BIGSERIAL PRIMARY KEY,
            event_type TEXT NOT NULL,
            object_type TEXT NOT NULL,
            object_id TEXT NOT NULL,
            user_id BIGINT NOT NULL,
            username TEXT NOT NULL,
            action TEXT NOT NULL,
            old_values TEXT,
            new_values TEXT,
            ip_address TEXT NOT NULL DEFAULT '',
            user_agent TEXT,
            session_id TEXT,
            context TEXT,
            event_hash TEXT NOT NULL,
            previous_hash TEXT,
            is_compliant BOOLEAN NOT NULL DEFAULT TRUE,
            retention_date TIMESTAMP NOT NULL,
            timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );

        CREATE UNIQUE INDEX IF NOT EXISTS uq_audit_events_event_hash ON public.audit_events(event_hash);
        CREATE INDEX IF NOT EXISTS idx_audit_events_event_type ON public.audit_events(event_type);
        CREATE INDEX IF NOT EXISTS idx_audit_events_object_type ON public.audit_events(object_type);
        CREATE INDEX IF NOT EXISTS idx_audit_events_object_id ON public.audit_events(object_id);
        CREATE INDEX IF NOT EXISTS idx_audit_events_user_id ON public.audit_events(user_id);
        CREATE INDEX IF NOT EXISTS idx_audit_events_session_id ON public.audit_events(session_id);
        CREATE INDEX IF NOT EXISTS idx_audit_events_previous_hash ON public.audit_events(previous_hash);
        CREATE INDEX IF NOT EXISTS idx_audit_events_retention_date ON public.audit_events(retention_date);
        CREATE INDEX IF NOT EXISTS idx_audit_events_timestamp ON public.audit_events(timestamp);
    ELSE
        -- Patch partially-created table shapes
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS event_type TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS object_type TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS object_id TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS user_id BIGINT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS username TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS action TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS old_values TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS new_values TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS ip_address TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS user_agent TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS session_id TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS context TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS event_hash TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS previous_hash TEXT;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS is_compliant BOOLEAN DEFAULT TRUE;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS retention_date TIMESTAMP;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
        ALTER TABLE public.audit_events ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

        UPDATE public.audit_events SET ip_address = COALESCE(ip_address, '');
        UPDATE public.audit_events SET timestamp = COALESCE(timestamp, CURRENT_TIMESTAMP);

        CREATE UNIQUE INDEX IF NOT EXISTS uq_audit_events_event_hash ON public.audit_events(event_hash);
        CREATE INDEX IF NOT EXISTS idx_audit_events_event_type ON public.audit_events(event_type);
        CREATE INDEX IF NOT EXISTS idx_audit_events_object_type ON public.audit_events(object_type);
        CREATE INDEX IF NOT EXISTS idx_audit_events_object_id ON public.audit_events(object_id);
        CREATE INDEX IF NOT EXISTS idx_audit_events_user_id ON public.audit_events(user_id);
        CREATE INDEX IF NOT EXISTS idx_audit_events_session_id ON public.audit_events(session_id);
        CREATE INDEX IF NOT EXISTS idx_audit_events_previous_hash ON public.audit_events(previous_hash);
        CREATE INDEX IF NOT EXISTS idx_audit_events_retention_date ON public.audit_events(retention_date);
        CREATE INDEX IF NOT EXISTS idx_audit_events_timestamp ON public.audit_events(timestamp);
    END IF;

    -- ------------------------------------------------------------------
    -- invoice_templates (required by startup default-template check)
    -- ------------------------------------------------------------------
    IF to_regclass('public.invoice_templates') IS NULL THEN
        CREATE TABLE public.invoice_templates (
            template_id BIGSERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            html_template TEXT NOT NULL,
            css_styles TEXT,
            is_default BOOLEAN NOT NULL DEFAULT FALSE,
            is_active BOOLEAN NOT NULL DEFAULT TRUE,
            created_by BIGINT,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );

        CREATE INDEX IF NOT EXISTS idx_invoice_templates_default ON public.invoice_templates(is_default);
        CREATE INDEX IF NOT EXISTS idx_invoice_templates_active ON public.invoice_templates(is_active);
        CREATE INDEX IF NOT EXISTS idx_invoice_templates_name ON public.invoice_templates(name);
    ELSE
        ALTER TABLE public.invoice_templates ADD COLUMN IF NOT EXISTS template_id BIGSERIAL;
        ALTER TABLE public.invoice_templates ADD COLUMN IF NOT EXISTS name TEXT;
        ALTER TABLE public.invoice_templates ADD COLUMN IF NOT EXISTS description TEXT;
        ALTER TABLE public.invoice_templates ADD COLUMN IF NOT EXISTS html_template TEXT;
        ALTER TABLE public.invoice_templates ADD COLUMN IF NOT EXISTS css_styles TEXT;
        ALTER TABLE public.invoice_templates ADD COLUMN IF NOT EXISTS is_default BOOLEAN DEFAULT FALSE;
        ALTER TABLE public.invoice_templates ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE;
        ALTER TABLE public.invoice_templates ADD COLUMN IF NOT EXISTS created_by BIGINT;
        ALTER TABLE public.invoice_templates ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
        ALTER TABLE public.invoice_templates ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

        CREATE INDEX IF NOT EXISTS idx_invoice_templates_default ON public.invoice_templates(is_default);
        CREATE INDEX IF NOT EXISTS idx_invoice_templates_active ON public.invoice_templates(is_active);
        CREATE INDEX IF NOT EXISTS idx_invoice_templates_name ON public.invoice_templates(name);
    END IF;
END$$;
