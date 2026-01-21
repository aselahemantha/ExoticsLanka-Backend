CREATE TABLE IF NOT EXISTS comparison_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL, -- References users(id) in shared DB
    listing_id      UUID NOT NULL, -- References car_listings(id) in shared DB
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, listing_id)
);

CREATE INDEX IF NOT EXISTS idx_comparison_items_user_id ON comparison_items(user_id);
CREATE INDEX IF NOT EXISTS idx_comparison_items_listing_id ON comparison_items(listing_id);
