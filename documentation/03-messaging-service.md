# Messaging Service

> Handles buyer-seller communication through conversation threads.

---

## Overview

The Messaging Service enables direct communication between buyers and sellers about specific listings. It manages conversation threads, individual messages, read status tracking, and unread counts.

**Responsibilities:**
- Create and manage conversations
- Send and receive messages
- Track read/unread status
- Calculate unread counts per user
- Archive/delete conversations
- Real-time updates (optional WebSocket support)

---

## Database Tables

### 1. Conversations Table

```sql
CREATE TABLE conversations (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id              UUID REFERENCES car_listings(id) ON DELETE SET NULL,
    buyer_id                UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seller_id               UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
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

-- Indexes
CREATE INDEX idx_conversations_buyer_id ON conversations(buyer_id);
CREATE INDEX idx_conversations_seller_id ON conversations(seller_id);
CREATE INDEX idx_conversations_listing_id ON conversations(listing_id);
CREATE INDEX idx_conversations_last_message ON conversations(last_message_at DESC);
CREATE INDEX idx_conversations_buyer_unread ON conversations(buyer_id, buyer_unread_count) WHERE buyer_unread_count > 0;
CREATE INDEX idx_conversations_seller_unread ON conversations(seller_id, seller_unread_count) WHERE seller_unread_count > 0;
```

### 2. Messages Table

```sql
CREATE TABLE messages (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id     UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content             TEXT NOT NULL,
    is_read             BOOLEAN DEFAULT FALSE,
    read_at             TIMESTAMP,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_created_at ON messages(conversation_id, created_at);
CREATE INDEX idx_messages_unread ON messages(conversation_id, is_read) WHERE is_read = FALSE;
```

---

## Field Descriptions

### Conversations

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Conversation identifier |
| `listing_id` | UUID | Related listing (nullable if deleted) |
| `buyer_id` | UUID | Buyer in conversation |
| `seller_id` | UUID | Seller in conversation |
| `listing_title` | VARCHAR | Cached listing title |
| `listing_image` | TEXT | Cached listing cover image |
| `listing_price` | DECIMAL | Cached listing price |
| `last_message_at` | TIMESTAMP | Timestamp of last message |
| `last_message_preview` | VARCHAR | Preview of last message (truncated) |
| `buyer_unread_count` | INT | Unread messages for buyer |
| `seller_unread_count` | INT | Unread messages for seller |
| `is_archived_by_buyer` | BOOLEAN | Buyer archived this conversation |
| `is_archived_by_seller` | BOOLEAN | Seller archived this conversation |

### Messages

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Message identifier |
| `conversation_id` | UUID | Parent conversation |
| `sender_id` | UUID | User who sent the message |
| `content` | TEXT | Message content |
| `is_read` | BOOLEAN | Whether recipient has read |
| `read_at` | TIMESTAMP | When message was read |
| `created_at` | TIMESTAMP | When message was sent |

---

## API Endpoints

### Conversations

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/conversations` | Get all user's conversations | Yes |
| `GET` | `/api/conversations/:id` | Get conversation with messages | Yes |
| `POST` | `/api/conversations` | Create new conversation | Yes |
| `DELETE` | `/api/conversations/:id` | Archive conversation | Yes |
| `PUT` | `/api/conversations/:id/read` | Mark all messages as read | Yes |

### Messages

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/conversations/:id/messages` | Get messages (paginated) | Yes |
| `POST` | `/api/conversations/:id/messages` | Send message | Yes |

### Utility

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/messages/unread-count` | Get total unread count | Yes |

---

## Request/Response Examples

### GET /api/conversations

Get all conversations for the authenticated user.

**Query Parameters:**
```
?page=1
&limit=20
&archived=false
```

**Response:**
```json
{
  "success": true,
  "data": {
    "conversations": [
      {
        "id": "conv-uuid-1",
        "listing": {
          "id": "listing-uuid-1",
          "title": "Mercedes-Benz S-Class S500",
          "image": "https://storage.exotics.lk/listings/cover.jpg",
          "price": 85000000
        },
        "participant": {
          "id": "user-uuid",
          "name": "Premium Auto Gallery",
          "avatar": "https://storage.exotics.lk/avatars/dealer.jpg",
          "role": "dealer"
        },
        "lastMessage": "Hi, is this still available?",
        "lastMessageAt": "2024-01-16T10:30:00Z",
        "unreadCount": 2,
        "createdAt": "2024-01-15T08:00:00Z"
      },
      {
        "id": "conv-uuid-2",
        "listing": {
          "id": "listing-uuid-2",
          "title": "BMW 7 Series 740i",
          "image": "https://storage.exotics.lk/listings/cover2.jpg",
          "price": 72000000
        },
        "participant": {
          "id": "user-uuid-2",
          "name": "John Doe",
          "avatar": null,
          "role": "seller"
        },
        "lastMessage": "Yes, you can come see it tomorrow",
        "lastMessageAt": "2024-01-15T14:20:00Z",
        "unreadCount": 0,
        "createdAt": "2024-01-14T09:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 5,
      "totalPages": 1
    },
    "totalUnread": 2
  }
}
```

### GET /api/conversations/:id

Get a single conversation with its messages.

**Query Parameters:**
```
?messagesLimit=50
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "conv-uuid-1",
    "listing": {
      "id": "listing-uuid-1",
      "title": "Mercedes-Benz S-Class S500",
      "image": "https://storage.exotics.lk/listings/cover.jpg",
      "price": 85000000,
      "status": "active"
    },
    "participant": {
      "id": "user-uuid",
      "name": "Premium Auto Gallery",
      "avatar": "https://storage.exotics.lk/avatars/dealer.jpg",
      "role": "dealer",
      "phone": "+94771234567",
      "email": "dealer@example.com"
    },
    "messages": [
      {
        "id": "msg-uuid-1",
        "senderId": "current-user-id",
        "senderName": "You",
        "content": "Hi, I'm interested in this Mercedes. Is it still available?",
        "isRead": true,
        "readAt": "2024-01-15T08:05:00Z",
        "createdAt": "2024-01-15T08:00:00Z"
      },
      {
        "id": "msg-uuid-2",
        "senderId": "user-uuid",
        "senderName": "Premium Auto Gallery",
        "content": "Yes, it's available. Would you like to schedule a viewing?",
        "isRead": true,
        "readAt": "2024-01-15T10:00:00Z",
        "createdAt": "2024-01-15T09:30:00Z"
      },
      {
        "id": "msg-uuid-3",
        "senderId": "current-user-id",
        "senderName": "You",
        "content": "Great! Can I come see it this Saturday?",
        "isRead": false,
        "readAt": null,
        "createdAt": "2024-01-16T10:30:00Z"
      }
    ],
    "createdAt": "2024-01-15T08:00:00Z"
  }
}
```

### POST /api/conversations

Create a new conversation (or return existing one).

**Request Body:**
```json
{
  "listingId": "listing-uuid",
  "sellerId": "seller-uuid",
  "initialMessage": "Hi, I'm interested in this vehicle. Is it still available?"
}
```

**Response (New Conversation):**
```json
{
  "success": true,
  "message": "Conversation created",
  "data": {
    "id": "conv-uuid-new",
    "isNew": true
  }
}
```

**Response (Existing Conversation):**
```json
{
  "success": true,
  "message": "Message sent to existing conversation",
  "data": {
    "id": "conv-uuid-existing",
    "isNew": false
  }
}
```

### POST /api/conversations/:id/messages

Send a message in a conversation.

**Request Body:**
```json
{
  "content": "Can you provide more details about the service history?"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "msg-uuid-new",
    "conversationId": "conv-uuid",
    "senderId": "current-user-id",
    "content": "Can you provide more details about the service history?",
    "isRead": false,
    "createdAt": "2024-01-16T11:00:00Z"
  }
}
```

### PUT /api/conversations/:id/read

Mark all messages in conversation as read.

**Response:**
```json
{
  "success": true,
  "message": "Messages marked as read",
  "data": {
    "markedCount": 3
  }
}
```

### GET /api/messages/unread-count

Get total unread messages across all conversations.

**Response:**
```json
{
  "success": true,
  "data": {
    "totalUnread": 5,
    "byConversation": [
      { "conversationId": "conv-uuid-1", "unread": 3 },
      { "conversationId": "conv-uuid-2", "unread": 2 }
    ]
  }
}
```

---

## Business Logic

### Creating a Conversation

```javascript
async function createConversation(buyerId, listingId, sellerId, initialMessage) {
  // Verify listing exists
  const listing = await db.query(
    'SELECT id, title, user_id, price FROM car_listings WHERE id = $1',
    [listingId]
  );
  
  if (!listing.rows[0]) {
    throw new NotFoundError('Listing not found');
  }
  
  // Get cover image
  const coverImage = await db.query(
    'SELECT image_url FROM listing_images WHERE listing_id = $1 AND is_cover = TRUE LIMIT 1',
    [listingId]
  );
  
  // Check for existing conversation
  const existing = await db.query(
    'SELECT id FROM conversations WHERE listing_id = $1 AND buyer_id = $2 AND seller_id = $3',
    [listingId, buyerId, sellerId]
  );
  
  let conversationId;
  let isNew = false;
  
  if (existing.rows[0]) {
    conversationId = existing.rows[0].id;
  } else {
    // Create new conversation
    const conv = await db.query(`
      INSERT INTO conversations (
        listing_id, buyer_id, seller_id, 
        listing_title, listing_image, listing_price,
        last_message_preview
      ) VALUES ($1, $2, $3, $4, $5, $6, $7)
      RETURNING id
    `, [
      listingId, 
      buyerId, 
      sellerId, 
      listing.rows[0].title,
      coverImage.rows[0]?.image_url,
      listing.rows[0].price,
      initialMessage.substring(0, 100)
    ]);
    
    conversationId = conv.rows[0].id;
    isNew = true;
  }
  
  // Send initial message
  await sendMessage(conversationId, buyerId, initialMessage);
  
  return { id: conversationId, isNew };
}
```

### Sending a Message

```javascript
async function sendMessage(conversationId, senderId, content) {
  // Verify conversation exists and user is participant
  const conv = await db.query(
    'SELECT buyer_id, seller_id FROM conversations WHERE id = $1',
    [conversationId]
  );
  
  if (!conv.rows[0]) {
    throw new NotFoundError('Conversation not found');
  }
  
  const { buyer_id, seller_id } = conv.rows[0];
  
  if (senderId !== buyer_id && senderId !== seller_id) {
    throw new ForbiddenError('You are not a participant in this conversation');
  }
  
  // Insert message
  const message = await db.query(`
    INSERT INTO messages (conversation_id, sender_id, content)
    VALUES ($1, $2, $3)
    RETURNING *
  `, [conversationId, senderId, content]);
  
  // Update conversation
  const unreadColumn = senderId === buyer_id ? 'seller_unread_count' : 'buyer_unread_count';
  
  await db.query(`
    UPDATE conversations 
    SET 
      last_message_at = NOW(),
      last_message_preview = $1,
      ${unreadColumn} = ${unreadColumn} + 1
    WHERE id = $2
  `, [content.substring(0, 100), conversationId]);
  
  // Trigger notification to recipient
  const recipientId = senderId === buyer_id ? seller_id : buyer_id;
  await notificationService.sendNewMessageNotification(recipientId, conversationId);
  
  return message.rows[0];
}
```

### Marking Messages as Read

```javascript
async function markAsRead(conversationId, userId) {
  // Get conversation and verify user is participant
  const conv = await db.query(
    'SELECT buyer_id, seller_id FROM conversations WHERE id = $1',
    [conversationId]
  );
  
  if (!conv.rows[0]) {
    throw new NotFoundError('Conversation not found');
  }
  
  const { buyer_id, seller_id } = conv.rows[0];
  
  if (userId !== buyer_id && userId !== seller_id) {
    throw new ForbiddenError('You are not a participant in this conversation');
  }
  
  // Mark messages as read (only messages NOT sent by the user)
  const result = await db.query(`
    UPDATE messages 
    SET is_read = TRUE, read_at = NOW()
    WHERE conversation_id = $1 
      AND sender_id != $2 
      AND is_read = FALSE
  `, [conversationId, userId]);
  
  // Reset unread count
  const unreadColumn = userId === buyer_id ? 'buyer_unread_count' : 'seller_unread_count';
  
  await db.query(`
    UPDATE conversations SET ${unreadColumn} = 0 WHERE id = $1
  `, [conversationId]);
  
  return result.rowCount;
}
```

### Getting User's Conversations

```javascript
async function getUserConversations(userId, page = 1, limit = 20, archived = false) {
  const offset = (page - 1) * limit;
  
  const conversations = await db.query(`
    SELECT 
      c.id,
      c.listing_id,
      c.listing_title,
      c.listing_image,
      c.listing_price,
      c.last_message_at,
      c.last_message_preview,
      c.created_at,
      CASE 
        WHEN c.buyer_id = $1 THEN c.buyer_unread_count
        ELSE c.seller_unread_count
      END as unread_count,
      CASE 
        WHEN c.buyer_id = $1 THEN c.seller_id
        ELSE c.buyer_id
      END as participant_id,
      u.name as participant_name,
      u.avatar_url as participant_avatar,
      u.role as participant_role,
      cl.status as listing_status
    FROM conversations c
    JOIN users u ON u.id = CASE 
      WHEN c.buyer_id = $1 THEN c.seller_id
      ELSE c.buyer_id
    END
    LEFT JOIN car_listings cl ON cl.id = c.listing_id
    WHERE (c.buyer_id = $1 OR c.seller_id = $1)
      AND CASE 
        WHEN c.buyer_id = $1 THEN c.is_archived_by_buyer
        ELSE c.is_archived_by_seller
      END = $2
    ORDER BY c.last_message_at DESC
    LIMIT $3 OFFSET $4
  `, [userId, archived, limit, offset]);
  
  // Get total count
  const total = await db.query(`
    SELECT COUNT(*) FROM conversations
    WHERE (buyer_id = $1 OR seller_id = $1)
      AND CASE 
        WHEN buyer_id = $1 THEN is_archived_by_buyer
        ELSE is_archived_by_seller
      END = $2
  `, [userId, archived]);
  
  // Get total unread
  const unread = await db.query(`
    SELECT 
      SUM(CASE WHEN buyer_id = $1 THEN buyer_unread_count ELSE seller_unread_count END) as total
    FROM conversations
    WHERE buyer_id = $1 OR seller_id = $1
  `, [userId]);
  
  return {
    conversations: conversations.rows,
    pagination: {
      page,
      limit,
      total: parseInt(total.rows[0].count),
      totalPages: Math.ceil(parseInt(total.rows[0].count) / limit)
    },
    totalUnread: parseInt(unread.rows[0].total) || 0
  };
}
```

---

## WebSocket Events (Optional)

For real-time messaging, implement WebSocket events:

### Client → Server Events

```javascript
// Join a conversation room
socket.emit('join_conversation', { conversationId });

// Leave a conversation room
socket.emit('leave_conversation', { conversationId });

// Send a message
socket.emit('send_message', { 
  conversationId, 
  content: "Message text" 
});

// User is typing
socket.emit('typing', { conversationId });

// User stopped typing
socket.emit('stop_typing', { conversationId });

// Mark messages as read
socket.emit('mark_read', { conversationId });
```

### Server → Client Events

```javascript
// New message received
socket.on('new_message', {
  id: "msg-uuid",
  conversationId: "conv-uuid",
  senderId: "user-uuid",
  senderName: "John Doe",
  content: "Message text",
  createdAt: "2024-01-16T11:00:00Z"
});

// Message was read by recipient
socket.on('message_read', {
  conversationId: "conv-uuid",
  readBy: "user-uuid",
  readAt: "2024-01-16T11:05:00Z"
});

// User is typing
socket.on('user_typing', {
  conversationId: "conv-uuid",
  userId: "user-uuid",
  userName: "John Doe"
});

// User stopped typing
socket.on('user_stopped_typing', {
  conversationId: "conv-uuid",
  userId: "user-uuid"
});

// Unread count updated
socket.on('unread_count_updated', {
  totalUnread: 5
});
```

---

## Validation Rules

```javascript
const messageValidation = {
  content: {
    required: true,
    minLength: 1,
    maxLength: 2000
  }
};

const conversationValidation = {
  listingId: {
    required: true,
    format: 'uuid'
  },
  sellerId: {
    required: true,
    format: 'uuid'
  },
  initialMessage: {
    required: true,
    minLength: 1,
    maxLength: 2000
  }
};
```

---

## Error Responses

```json
// 404 - Conversation Not Found
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Conversation not found"
  }
}

// 403 - Not a Participant
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "You are not a participant in this conversation"
  }
}

// 400 - Cannot Message Self
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "You cannot send messages to yourself"
  }
}

// 400 - Empty Message
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Message content is required"
  }
}
```

---

## Related Services

- **Listings Service** - Provides listing context for conversations
- **Notification Service** - Sends email/push notifications for new messages
- **Analytics Service** - Tracks messaging metrics for dealers (response time)
- **User Service** - Provides participant information

