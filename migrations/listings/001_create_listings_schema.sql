-- 001_create_listings_schema.sql

-- Enable uuid extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create listings table
CREATE TABLE IF NOT EXISTS listings (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id    UUID NOT NULL, -- references users.id ideally, but microservices so soft reference
    title        VARCHAR(255) NOT NULL,
    make         VARCHAR(100) NOT NULL,
    model        VARCHAR(100) NOT NULL,
    year         INTEGER NOT NULL,
    price        DECIMAL(15, 2) NOT NULL, -- Keep numeric pricing standard
    mileage      INTEGER NOT NULL,
    condition    VARCHAR(50) NOT NULL,
    body_type    VARCHAR(100),
    fuel_type    VARCHAR(100),
    transmission VARCHAR(100),
    color        VARCHAR(50),
    doors        INTEGER,
    seats        INTEGER,
    engine_size  VARCHAR(50),
    drivetrain   VARCHAR(50),
    features     JSONB, -- We'll store features as a JSON array
    description  TEXT,
    location     VARCHAR(255),
    status       VARCHAR(50) NOT NULL DEFAULT 'draft', -- active, sold, draft
    is_featured  BOOLEAN NOT NULL DEFAULT FALSE,
    is_trending  BOOLEAN NOT NULL DEFAULT FALSE,
    views        INTEGER NOT NULL DEFAULT 0,
    favorites    INTEGER NOT NULL DEFAULT 0,
    health_score INTEGER NOT NULL DEFAULT 0,
    verified     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for common search operations
CREATE INDEX IF NOT EXISTS idx_listings_make_model ON listings(make, model);
CREATE INDEX IF NOT EXISTS idx_listings_status ON listings(status);
CREATE INDEX IF NOT EXISTS idx_listings_seller ON listings(seller_id);
CREATE INDEX IF NOT EXISTS idx_listings_price ON listings(price);
CREATE INDEX IF NOT EXISTS idx_listings_year ON listings(year);

-- Create listing_images table
CREATE TABLE IF NOT EXISTS listing_images (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id  UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    is_primary  BOOLEAN NOT NULL DEFAULT FALSE,
    position    INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_listing_images_listing ON listing_images(listing_id);

-- Create listing_reports table
CREATE TABLE IF NOT EXISTS listing_reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id  UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    reporter_id UUID NOT NULL, -- soft reference to user
    reason      VARCHAR(100) NOT NULL,
    details     TEXT,
    status      VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, reviewed, dismissed
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create user_favorites table (Joining table for Tracking User Favorites)
CREATE TABLE IF NOT EXISTS user_favorites (
    user_id     UUID NOT NULL,
    listing_id  UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, listing_id)
);

-- Optional listing_stats could be a view or table, but since views/favorites are
-- tracked directly on the listing table as standard integers per the spec,
-- we'll skip creating a separate stats table and update the columns instead.
