CREATE TABLE IF NOT EXISTS listing_views (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id      UUID NOT NULL, -- References car_listings(id) in shared DB
    user_id         UUID,          -- References users(id) in shared DB
    
    -- Tracking Data
    session_id      VARCHAR(100),
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    referrer        TEXT,
    
    -- Event Type
    event_type      VARCHAR(20) DEFAULT 'view'
                    CHECK (event_type IN ('view', 'click', 'contact_view', 'phone_click', 'share')),
    
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_listing_views_listing_id ON listing_views(listing_id);
CREATE INDEX IF NOT EXISTS idx_listing_views_user_id ON listing_views(user_id);
CREATE INDEX IF NOT EXISTS idx_listing_views_created_at ON listing_views(created_at);
CREATE INDEX IF NOT EXISTS idx_listing_views_event ON listing_views(event_type);

-- Dealer Analytics (Aggregated Daily)
CREATE TABLE IF NOT EXISTS dealer_analytics (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dealer_id               UUID NOT NULL, -- References users(id)
    date                    DATE NOT NULL,
    
    -- Engagement Metrics
    total_views             INT DEFAULT 0,
    unique_viewers          INT DEFAULT 0,
    total_clicks            INT DEFAULT 0,
    total_favorites         INT DEFAULT 0,
    total_shares            INT DEFAULT 0,
    
    -- Conversion Metrics
    total_leads             INT DEFAULT 0,
    total_messages          INT DEFAULT 0,
    phone_reveals           INT DEFAULT 0,
    
    -- Sales Metrics
    total_sales             INT DEFAULT 0,
    total_revenue           DECIMAL(20, 2) DEFAULT 0,
    
    -- Performance Metrics
    avg_response_time_mins  INT,
    response_rate           DECIMAL(5, 2),
    
    -- Inventory Metrics
    inventory_count         INT DEFAULT 0,
    inventory_value         DECIMAL(20, 2) DEFAULT 0,
    avg_health_score        INT DEFAULT 0,
    avg_days_listed         INT DEFAULT 0,
    
    created_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(dealer_id, date)
);

CREATE INDEX IF NOT EXISTS idx_dealer_analytics_dealer_id ON dealer_analytics(dealer_id);
CREATE INDEX IF NOT EXISTS idx_dealer_analytics_date ON dealer_analytics(date DESC);

-- Market Trends (Aggregated)
CREATE TABLE IF NOT EXISTS market_trends (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date            DATE NOT NULL,
    
    -- Model Trends
    make            VARCHAR(100) NOT NULL,
    model           VARCHAR(100),
    
    -- Metrics
    search_count    INT DEFAULT 0,
    view_count      INT DEFAULT 0,
    listing_count   INT DEFAULT 0,
    avg_price       DECIMAL(15, 2),
    price_change    DECIMAL(5, 2),
    
    UNIQUE(date, make, model)
);

CREATE INDEX IF NOT EXISTS idx_market_trends_date ON market_trends(date DESC);
CREATE INDEX IF NOT EXISTS idx_market_trends_make ON market_trends(make);
