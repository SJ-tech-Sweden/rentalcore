-- Idempotent migration to create `user_preferences` table used by the app
CREATE TABLE IF NOT EXISTS user_preferences (
  preference_id SERIAL PRIMARY KEY,
  user_id INT NOT NULL UNIQUE,
  language VARCHAR(20) NOT NULL DEFAULT 'de',
  theme VARCHAR(50) NOT NULL DEFAULT 'dark',
  time_zone VARCHAR(100) NOT NULL DEFAULT 'Europe/Berlin',
  date_format VARCHAR(50) NOT NULL DEFAULT 'DD.MM.YYYY',
  time_format VARCHAR(20) NOT NULL DEFAULT '24h',

  email_notifications BOOLEAN NOT NULL DEFAULT TRUE,
  system_notifications BOOLEAN NOT NULL DEFAULT TRUE,
  job_status_notifications BOOLEAN NOT NULL DEFAULT TRUE,
  device_alert_notifications BOOLEAN NOT NULL DEFAULT TRUE,

  items_per_page INT NOT NULL DEFAULT 25,
  default_view VARCHAR(50) NOT NULL DEFAULT 'list',
  show_advanced_options BOOLEAN NOT NULL DEFAULT FALSE,
  auto_save_enabled BOOLEAN NOT NULL DEFAULT TRUE,

  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add FK to users if users table exists (converted migration 012 will adjust if needed)
DO $$
BEGIN
  IF to_regclass('public.users') IS NOT NULL THEN
    -- Add FK constraint if pref column exists and users PK present
    IF EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='user_preferences' AND column_name='user_id')
       AND EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='users' AND (column_name='userid' OR column_name='userID' OR column_name='id')) THEN
      -- Try to add FK named fk_user_preferences_user if it doesn't exist
      IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_preferences_user') THEN
        -- Use the converted migration logic to find the correct users column; keep simple here: reference users(userid) if available
        IF EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='userid') THEN
          EXECUTE 'ALTER TABLE user_preferences ADD CONSTRAINT fk_user_preferences_user FOREIGN KEY (user_id) REFERENCES users(userid) ON DELETE CASCADE ON UPDATE CASCADE';
        ELSIF EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='userID') THEN
          EXECUTE 'ALTER TABLE user_preferences ADD CONSTRAINT fk_user_preferences_user FOREIGN KEY (user_id) REFERENCES users("userID") ON DELETE CASCADE ON UPDATE CASCADE';
        ELSE
          -- fallback: no FK added
          RAISE NOTICE 'user_preferences: users primary key not found for FK, skipping';
        END IF;
      END IF;
    END IF;
  END IF;
END$$ LANGUAGE plpgsql;
