CREATE TABLE IF NOT EXISTS user_favorites (
    user_id         UUID NOT NULL,
    listing_id      UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (user_id, listing_id)
);

CREATE INDEX IF NOT EXISTS idx_user_favorites_user_id ON user_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_user_favorites_listing_id ON user_favorites(listing_id);
CREATE INDEX IF NOT EXISTS idx_user_favorites_created_at ON user_favorites(user_id, created_at DESC);
