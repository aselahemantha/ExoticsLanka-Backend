-- 1. Car Listings (Primary Table)
CREATE TABLE IF NOT EXISTS car_listings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL, -- Referenced from user service, so no FK constraint here strictly speaking unless in same DB
    
    -- Basic Information
    title               VARCHAR(255) NOT NULL,
    make                VARCHAR(100) NOT NULL,
    model               VARCHAR(100) NOT NULL,
    year                INT NOT NULL CHECK (year >= 1900 AND year <= 2030),
    price               DECIMAL(15, 2) NOT NULL CHECK (price > 0),
    mileage             INT NOT NULL CHECK (mileage >= 0),
    condition           VARCHAR(50) NOT NULL,
    
    -- Specifications
    transmission        VARCHAR(50),
    fuel_type           VARCHAR(50),
    body_type           VARCHAR(50),
    color               VARCHAR(50),
    doors               INT CHECK (doors >= 1 AND doors <= 6),
    seats               INT CHECK (seats >= 1 AND seats <= 12),
    engine_size         VARCHAR(20),
    drivetrain          VARCHAR(20),
    
    -- Description & Contact
    description         TEXT,
    location            VARCHAR(255) NOT NULL,
    contact_phone       VARCHAR(20),
    contact_email       VARCHAR(255),
    
    -- Status & Metrics
    status              VARCHAR(20) DEFAULT 'pending' 
                        CHECK (status IN ('draft', 'pending', 'active', 'sold', 'expired', 'rejected')),
    health_score        INT DEFAULT 50 CHECK (health_score >= 0 AND health_score <= 100),
    views               INT DEFAULT 0,
    favorites_count     INT DEFAULT 0,
    days_listed         INT DEFAULT 0,
    
    -- Flags
    is_new              BOOLEAN DEFAULT FALSE,
    is_featured         BOOLEAN DEFAULT FALSE,
    is_verified         BOOLEAN DEFAULT FALSE,
    trending            BOOLEAN DEFAULT FALSE,
    
    -- Pricing Intelligence
    market_avg_price    DECIMAL(15, 2),
    price_alert         TEXT,
    
    -- Timestamps
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at        TIMESTAMP,
    expires_at          TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_listings_user_id ON car_listings(user_id);
CREATE INDEX IF NOT EXISTS idx_listings_status ON car_listings(status);
CREATE INDEX IF NOT EXISTS idx_listings_make ON car_listings(make);
CREATE INDEX IF NOT EXISTS idx_listings_year ON car_listings(year);
CREATE INDEX IF NOT EXISTS idx_listings_price ON car_listings(price);
CREATE INDEX IF NOT EXISTS idx_listings_location ON car_listings(location);
CREATE INDEX IF NOT EXISTS idx_listings_created_at ON car_listings(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_listings_health_score ON car_listings(health_score DESC);

-- Full-text search
CREATE INDEX IF NOT EXISTS idx_listings_search ON car_listings 
USING GIN (to_tsvector('english', title || ' ' || make || ' ' || model || ' ' || COALESCE(description, '')));

-- 2. Listing Images
CREATE TABLE IF NOT EXISTS listing_images (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id      UUID NOT NULL REFERENCES car_listings(id) ON DELETE CASCADE,
    image_url       TEXT NOT NULL,
    thumbnail_url   TEXT,
    is_cover        BOOLEAN DEFAULT FALSE,
    sort_order      INT DEFAULT 0,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_listing_images_listing_id ON listing_images(listing_id);
CREATE INDEX IF NOT EXISTS idx_listing_images_cover ON listing_images(listing_id, is_cover) WHERE is_cover = TRUE;

-- 3. Listing Features
CREATE TABLE IF NOT EXISTS listing_features (
    id              SERIAL PRIMARY KEY,
    listing_id      UUID NOT NULL REFERENCES car_listings(id) ON DELETE CASCADE,
    feature_name    VARCHAR(100) NOT NULL,
    
    UNIQUE(listing_id, feature_name)
);

CREATE INDEX IF NOT EXISTS idx_listing_features_listing_id ON listing_features(listing_id);

-- 4. Car Brands (Lookup Table)
CREATE TABLE IF NOT EXISTS car_brands (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL UNIQUE,
    logo_url        TEXT,
    listing_count   INT DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sample Data
INSERT INTO car_brands (name) VALUES 
('Mercedes-Benz'), ('BMW'), ('Porsche'), ('Land Rover'), 
('Toyota'), ('Honda'), ('Lexus'), ('Audi'), ('Ferrari'), ('Lamborghini')
ON CONFLICT (name) DO NOTHING;
