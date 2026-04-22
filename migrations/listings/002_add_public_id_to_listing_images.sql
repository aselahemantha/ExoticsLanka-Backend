-- 002_add_public_id_to_listing_images.sql

ALTER TABLE listing_images ADD COLUMN IF NOT EXISTS public_id TEXT;
