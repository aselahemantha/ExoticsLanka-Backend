-- 001_create_auth_tables.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create users table
CREATE TABLE IF NOT EXISTS users (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email                 VARCHAR(255) UNIQUE NOT NULL,
  password_hash         VARCHAR(255) NOT NULL,
  status                VARCHAR(20) NOT NULL DEFAULT 'pending',
  email_verified        BOOLEAN DEFAULT FALSE,
  email_verified_at     TIMESTAMP,
  role                  VARCHAR(50) NOT NULL DEFAULT 'buyer',
  two_factor_enabled    BOOLEAN DEFAULT FALSE,
  two_factor_secret     VARCHAR(255),
  failed_login_attempts INTEGER DEFAULT 0,
  locked_until          TIMESTAMP,
  oauth_provider        VARCHAR(50),
  oauth_id              VARCHAR(255),
  created_at            TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMP NOT NULL DEFAULT NOW(),
  last_login_at         TIMESTAMP,
  deleted_at            TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create sessions table (if we decide to store sessions in DB too, or for backup)
-- Note: We are primarily using Redis for sessions, but having a DB table is good for persistence
CREATE TABLE IF NOT EXISTS sessions (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token             VARCHAR(500) UNIQUE NOT NULL,
  refresh_token     VARCHAR(500),
  device_id         VARCHAR(255),
  device_name       VARCHAR(100),
  ip_address        VARCHAR(45),
  user_agent        TEXT,
  is_active         BOOLEAN DEFAULT TRUE,
  expires_at        TIMESTAMP NOT NULL,
  last_activity_at  TIMESTAMP NOT NULL DEFAULT NOW(),
  created_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
  id                BIGSERIAL PRIMARY KEY,
  user_id           UUID REFERENCES users(id) ON DELETE SET NULL,
  event_type        VARCHAR(100) NOT NULL,
  event_category    VARCHAR(50) NOT NULL,
  description       TEXT,
  metadata          JSONB,
  ip_address        VARCHAR(45),
  user_agent        TEXT,
  success           BOOLEAN NOT NULL,
  error_message     TEXT,
  created_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
