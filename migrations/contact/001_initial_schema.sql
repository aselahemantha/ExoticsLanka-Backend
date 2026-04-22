CREATE TABLE IF NOT EXISTS contact_inquiries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Contact Info
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    phone           VARCHAR(20),
    
    -- Inquiry Details
    subject         VARCHAR(100),
    message         TEXT NOT NULL,
    
    -- Status
    status          VARCHAR(20) DEFAULT 'pending' 
                    CHECK (status IN ('pending', 'in_progress', 'responded', 'closed')),
    priority        VARCHAR(20) DEFAULT 'normal'
                    CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    
    -- Response
    admin_response  TEXT,
    responded_by    UUID, -- Referenced from users table (shared DB assumption)
    responded_at    TIMESTAMP,
    
    -- Metadata
    user_id         UUID, -- Referenced from users table (shared DB assumption)
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_contact_inquiries_status ON contact_inquiries(status);
CREATE INDEX IF NOT EXISTS idx_contact_inquiries_email ON contact_inquiries(email);
CREATE INDEX IF NOT EXISTS idx_contact_inquiries_created_at ON contact_inquiries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_contact_inquiries_subject ON contact_inquiries(subject);
CREATE INDEX IF NOT EXISTS idx_contact_inquiries_priority ON contact_inquiries(priority);
