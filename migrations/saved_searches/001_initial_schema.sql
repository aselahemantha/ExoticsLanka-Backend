CREATE TABLE IF NOT EXISTS saved_searches (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL, -- Referenced from users table (shared DB assumption)
    name                VARCHAR(100) NOT NULL,
    
    -- Search Filters (stored as JSON)
    filters             JSONB NOT NULL,
    
    -- Alert Settings
    alert_enabled       BOOLEAN DEFAULT TRUE,
    alert_frequency     VARCHAR(20) DEFAULT 'daily' 
                        CHECK (alert_frequency IN ('instant', 'daily', 'weekly')),
    
    -- Match Tracking
    last_checked        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_notified       TIMESTAMP,
    new_matches_count   INT DEFAULT 0,
    total_matches       INT DEFAULT 0,
    
    -- Timestamps
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_saved_searches_user_id ON saved_searches(user_id);
CREATE INDEX IF NOT EXISTS idx_saved_searches_alert ON saved_searches(alert_enabled, alert_frequency);
CREATE INDEX IF NOT EXISTS idx_saved_searches_last_checked ON saved_searches(last_checked);
