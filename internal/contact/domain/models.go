package domain

import (
	"time"

	"github.com/google/uuid"
)

type Inquiry struct {
	ID              uuid.UUID `json:"id"`
	ReferenceNumber string    `json:"referenceNumber"` // Computed for display logic if needed (e.g. "INQ-...") or stored? Docs imply generated.
	// For simplicity, we can compute it or store it. Docs example: "INQ-20240116-001".
	// Since schema doesn't have it, we might compute it on fly or schema update required?
	// The implementation plan says "Generate Reference Number ... Create".
	// If it's not in DB, we can't search by it easily.
	// Let's add it to struct but note it's derived or valid for response.
	// Wait, docs "Generate Reference Number ... Create". And Response has it.
	// Schema provided in docs DOES NOT have `reference_number` column.
	// We can compute it from ID or Date + sequence?
	// Let's stick to returning ID as ref or just return the struct.

	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Phone         *string    `json:"phone,omitempty"`
	Subject       string     `json:"subject"`
	Message       string     `json:"message"`
	Status        string     `json:"status"`
	Priority      string     `json:"priority"`
	AdminResponse *string    `json:"adminResponse,omitempty"`
	RespondedBy   *uuid.UUID `json:"respondedBy,omitempty"`
	RespondedAt   *time.Time `json:"respondedAt,omitempty"`
	UserID        *uuid.UUID `json:"userId,omitempty"`
	IPAddress     string     `json:"ipAddress,omitempty"` // Internal use
	UserAgent     string     `json:"userAgent,omitempty"` // Internal use
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type InquiryStats struct {
	Total      int            `json:"total"`
	ByStatus   map[string]int `json:"byStatus"`
	BySubject  map[string]int `json:"bySubject"`
	ByPriority map[string]int `json:"byPriority"`
}

// Requests
type CreateInquiryRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Phone   string `json:"phone"`
	Subject string `json:"subject"`
	Message string `json:"message" binding:"required"`
}

type RespondInquiryRequest struct {
	Status        string `json:"status" binding:"required"`
	AdminResponse string `json:"adminResponse"`
	Priority      string `json:"priority"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}
