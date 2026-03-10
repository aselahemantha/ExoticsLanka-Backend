-- 001_create_user_schema.sql

-- Enable uuid extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table (synchronizes/mirrors core elements with Auth Service over event bus / logic in real life)
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY,     -- Same UUID as in auth-service
    name        VARCHAR(255),
    email       VARCHAR(255) UNIQUE NOT NULL,
    role        VARCHAR(50) NOT NULL, -- buyer, seller, dealer
    phone       VARCHAR(50),
    avatar_url  TEXT,
    bio         TEXT,
    location    VARCHAR(255),
    verified    BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMP,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for fast lookup by email
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);


-- Create verification_requests table
CREATE TABLE IF NOT EXISTS verification_requests (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_type VARCHAR(50) NOT NULL, -- dealer
    status       VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),
    reviewed_at  TIMESTAMP,
    reviewer_id  UUID,
    notes        TEXT
);

-- Index for looking up verification requests by user 
CREATE INDEX IF NOT EXISTS idx_verification_requests_user ON verification_requests(user_id);
