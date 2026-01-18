CREATE TABLE IF NOT EXISTS conversations (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id              UUID REFERENCES car_listings(id) ON DELETE SET NULL,
    buyer_id                UUID NOT NULL, -- Referenced from users table
    seller_id               UUID NOT NULL, -- Referenced from users table
    
    -- Cached listing info (in case listing is deleted)
    listing_title           VARCHAR(255),
    listing_image           TEXT,
    listing_price           DECIMAL(15, 2),
    
    -- Status tracking
    last_message_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_message_preview    VARCHAR(100),
    buyer_unread_count      INT DEFAULT 0,
    seller_unread_count     INT DEFAULT 0,
    
    -- Archive flags
    is_archived_by_buyer    BOOLEAN DEFAULT FALSE,
    is_archived_by_seller   BOOLEAN DEFAULT FALSE,
    
    created_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Prevent duplicate conversations for same listing between same users
    UNIQUE(listing_id, buyer_id, seller_id)
);

CREATE TABLE IF NOT EXISTS messages (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id     UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id           UUID NOT NULL, -- Referenced from users table
    content             TEXT NOT NULL,
    is_read             BOOLEAN DEFAULT FALSE,
    read_at             TIMESTAMP,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_conversations_buyer_id ON conversations(buyer_id);
CREATE INDEX IF NOT EXISTS idx_conversations_seller_id ON conversations(seller_id);
CREATE INDEX IF NOT EXISTS idx_conversations_listing_id ON conversations(listing_id);
CREATE INDEX IF NOT EXISTS idx_conversations_last_message ON conversations(last_message_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversations_buyer_unread ON conversations(buyer_id, buyer_unread_count) WHERE buyer_unread_count > 0;
CREATE INDEX IF NOT EXISTS idx_conversations_seller_unread ON conversations(seller_id, seller_unread_count) WHERE seller_unread_count > 0;

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(conversation_id, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_unread ON messages(conversation_id, is_read) WHERE is_read = FALSE;
