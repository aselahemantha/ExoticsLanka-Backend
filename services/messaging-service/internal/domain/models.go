package domain

import (
	"time"

	"github.com/google/uuid"
)

// Conversation represents a chat thread between buyer and seller
type Conversation struct {
	ID        uuid.UUID  `json:"id"`
	ListingID *uuid.UUID `json:"listingId,omitempty"`
	BuyerID   uuid.UUID  `json:"buyerId,omitempty"`  // Internal use mainly
	SellerID  uuid.UUID  `json:"sellerId,omitempty"` // Internal use mainly

	// Cached Listing Info
	ListingTitle string  `json:"listingTitle,omitempty"` // Mapped to nested structure in response usually
	ListingImage *string `json:"listingImage,omitempty"`
	ListingPrice float64 `json:"listingPrice,omitempty"`

	// Status
	LastMessageAt      time.Time `json:"lastMessageAt"`
	LastMessagePreview string    `json:"lastMessage"`

	// Calculated/Contextual fields
	UnreadCount int             `json:"unreadCount"`
	Participant *UserSummary    `json:"participant,omitempty"`
	Listing     *ListingSummary `json:"listing,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
}

type Message struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversationId"`
	SenderID       uuid.UUID  `json:"senderId"`
	Content        string     `json:"content"`
	IsRead         bool       `json:"isRead"`
	ReadAt         *time.Time `json:"readAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`

	SenderName string `json:"senderName,omitempty"` // Hydrated
}

type UserSummary struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Avatar *string   `json:"avatar,omitempty"`
	Role   string    `json:"role,omitempty"`
}

type ListingSummary struct {
	ID     uuid.UUID `json:"id"`
	Title  string    `json:"title"`
	Image  *string   `json:"image,omitempty"`
	Price  float64   `json:"price"`
	Status string    `json:"status,omitempty"` // Fetched from Real listing table if possible
}

// Request Models
type CreateConversationRequest struct {
	ListingID      uuid.UUID `json:"listingId" binding:"required"`
	SellerID       uuid.UUID `json:"sellerId" binding:"required"`
	InitialMessage string    `json:"initialMessage" binding:"required"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// Response Models
type ConversationResponse struct {
	ID    uuid.UUID `json:"id"`
	IsNew bool      `json:"isNew"`
}

type UnreadCountResponse struct {
	TotalUnread    int                  `json:"totalUnread"`
	ByConversation []ConversationUnread `json:"byConversation"`
}

type ConversationUnread struct {
	ConversationID uuid.UUID `json:"conversationId"`
	Unread         int       `json:"unread"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}
