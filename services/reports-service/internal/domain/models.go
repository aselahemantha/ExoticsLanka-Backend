package domain

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID          uuid.UUID       `json:"id"`
	ListingID   uuid.UUID       `json:"listingId"`
	Listing     *ListingSummary `json:"listing,omitempty"` // For expanded view
	ReporterID  *uuid.UUID      `json:"reporterId,omitempty"`
	Reporter    *UserSummary    `json:"reporter,omitempty"`
	Reason      string          `json:"reason"`
	Details     string          `json:"details"`
	Status      string          `json:"status"`
	AdminNotes  string          `json:"adminNotes,omitempty"`
	ActionTaken string          `json:"actionTaken,omitempty"`
	ResolvedAt  *time.Time      `json:"resolvedAt,omitempty"`
	ResolvedBy  *uuid.UUID      `json:"resolvedBy,omitempty"`
	Resolver    *UserSummary    `json:"resolver,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type ListingSummary struct {
	ID         uuid.UUID   `json:"id"`
	Title      string      `json:"title"`
	Price      float64     `json:"price"`
	Status     string      `json:"status"`
	CoverImage *string     `json:"coverImage,omitempty"`
	User       UserSummary `json:"user"`
}

type UserSummary struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

// Stats
type ReportStats struct {
	TotalReports           int            `json:"totalReports"`
	ByStatus               map[string]int `json:"byStatus"`
	ByReason               map[string]int `json:"byReason"`
	AvgResolutionTimeHours float64        `json:"avgResolutionTimeHours"`
}

// Requests
type CreateReportRequest struct {
	ListingID uuid.UUID `json:"listingId" binding:"required"`
	Reason    string    `json:"reason" binding:"required"`
	Details   string    `json:"details"`
}

type ResolveReportRequest struct {
	Status      string `json:"status" binding:"required,oneof=pending reviewing resolved dismissed"`
	AdminNotes  string `json:"adminNotes"`
	ActionTaken string `json:"actionTaken"` // 'listing_removed', 'user_suspended', etc.
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}
