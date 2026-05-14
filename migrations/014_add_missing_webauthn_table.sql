-- ================================================================
-- MIGRATION 014: ADD MISSING WEBAUTHN SESSION TABLE
-- Creates webauthn_sessions table if it doesn't exist properly
-- ================================================================

-- Drop and recreate the table to ensure proper structure
DROP TABLE IF EXISTS webauthn_sessions;

CREATE TABLE IF NOT EXISTS webauthn_sessions (
    session_id VARCHAR(191) PRIMARY KEY,
    user_id BIGINT NOT NULL DEFAULT 0,
    challenge VARCHAR(255) NOT NULL,
    session_type VARCHAR(50) NOT NULL,
    session_data TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webauthn_user_session ON webauthn_sessions(user_id, session_type);
CREATE INDEX IF NOT EXISTS idx_webauthn_expires ON webauthn_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_webauthn_session_type ON webauthn_sessions(session_type);