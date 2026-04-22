CREATE TABLE IF NOT EXISTS reviews (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id          UUID REFERENCES car_listings(id) ON DELETE SET NULL, -- Nullable if listing deleted
    seller_id           UUID NOT NULL, -- Referenced from users table
    buyer_id            UUID NOT NULL, -- Referenced from users table
    
    -- Review Content
    rating              INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title               VARCHAR(255),
    comment             TEXT,
    
    -- Metadata
    verified_purchase   BOOLEAN DEFAULT FALSE,
    helpful_count       INT DEFAULT 0,
    
    -- Seller Response
    seller_response     TEXT,
    seller_response_at  TIMESTAMP,
    
    -- Timestamps
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Prevent duplicate reviews from same buyer for same listing
    UNIQUE(listing_id, buyer_id)
);

CREATE TABLE IF NOT EXISTS review_helpful_votes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id       UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL, -- Referenced from users table
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- One vote per user per review
    UNIQUE(review_id, user_id)
);

CREATE TABLE IF NOT EXISTS review_photos (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id       UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    photo_url       TEXT NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_reviews_seller_id ON reviews(seller_id);
CREATE INDEX IF NOT EXISTS idx_reviews_buyer_id ON reviews(buyer_id);
CREATE INDEX IF NOT EXISTS idx_reviews_listing_id ON reviews(listing_id);
CREATE INDEX IF NOT EXISTS idx_reviews_rating ON reviews(seller_id, rating);
CREATE INDEX IF NOT EXISTS idx_reviews_created_at ON reviews(seller_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reviews_helpful ON reviews(helpful_count DESC);

CREATE INDEX IF NOT EXISTS idx_review_votes_review_id ON review_helpful_votes(review_id);
CREATE INDEX IF NOT EXISTS idx_review_votes_user_id ON review_helpful_votes(user_id);

CREATE INDEX IF NOT EXISTS idx_review_photos_review_id ON review_photos(review_id);
