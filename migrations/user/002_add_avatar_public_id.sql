-- 002_add_avatar_public_id.sql

ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_public_id TEXT;
-- The code was using profile_image_url but migration had avatar_url. 
-- We'll stay with avatar_url and fix the code.
