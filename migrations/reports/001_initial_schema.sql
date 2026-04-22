CREATE TABLE IF NOT EXISTS listing_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id      UUID NOT NULL, -- Referenced from car_listings (shared DB assumption)
    reporter_id     UUID,          -- Referenced from users (shared DB assumption)
    
    -- Report Details
    reason          VARCHAR(50) NOT NULL,
    details         TEXT,
    
    -- Moderation
    status          VARCHAR(20) DEFAULT 'pending' 
                    CHECK (status IN ('pending', 'reviewing', 'resolved', 'dismissed')),
    admin_notes     TEXT,
    action_taken    VARCHAR(50),
    resolved_at     TIMESTAMP,
    resolved_by     UUID,
    
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_listing_reports_listing_id ON listing_reports(listing_id);
CREATE INDEX IF NOT EXISTS idx_listing_reports_status ON listing_reports(status);
CREATE INDEX IF NOT EXISTS idx_listing_reports_reason ON listing_reports(reason);
CREATE INDEX IF NOT EXISTS idx_listing_reports_created_at ON listing_reports(created_at DESC);
