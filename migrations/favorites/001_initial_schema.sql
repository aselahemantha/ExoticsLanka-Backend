CREATE TABLE IF NOT EXISTS favorites (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL, -- Referenced from users table in auth-service (logically)
    listing_id      UUID NOT NULL REFERENCES car_listings(id) ON DELETE CASCADE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Prevent duplicate favorites
    UNIQUE(user_id, listing_id)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_listing_id ON favorites(listing_id);
CREATE INDEX IF NOT EXISTS idx_favorites_created_at ON favorites(user_id, created_at DESC);
