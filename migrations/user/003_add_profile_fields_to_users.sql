-- 003_add_profile_fields_to_users.sql

-- Add missing columns to users table if they don't exist
-- This handles the case where the auth module creates the users table first

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='name') THEN
        ALTER TABLE users ADD COLUMN name VARCHAR(255);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='phone') THEN
        ALTER TABLE users ADD COLUMN phone VARCHAR(50);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='avatar_url') THEN
        ALTER TABLE users ADD COLUMN avatar_url TEXT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='bio') THEN
        ALTER TABLE users ADD COLUMN bio TEXT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='location') THEN
        ALTER TABLE users ADD COLUMN location VARCHAR(255);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='verified') THEN
        ALTER TABLE users ADD COLUMN verified BOOLEAN NOT NULL DEFAULT FALSE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='verified_at') THEN
        ALTER TABLE users ADD COLUMN verified_at TIMESTAMP;
    END IF;
END $$;
