-- 002_add_public_id_to_listing_images.sql

ALTER TABLE listing_images ADD COLUMN IF NOT EXISTS public_id TEXT;
ALTER TABLE listing_images RENAME COLUMN image_url TO url;
ALTER TABLE listing_images RENAME COLUMN display_order TO position;
