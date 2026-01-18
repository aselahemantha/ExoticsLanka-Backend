package domain

import (
	"time"

	"github.com/google/uuid"
)

// Review represents a buyer's review of a seller
type Review struct {
	ID               uuid.UUID  `json:"id"`
	ListingID        *uuid.UUID `json:"listingId,omitempty"` // Nullable if listing deleted
	SellerID         uuid.UUID  `json:"sellerId"`
	BuyerID          uuid.UUID  `json:"buyerId"`
	Rating           int        `json:"rating"`
	Title            string     `json:"title"`
	Comment          string     `json:"comment"`
	VerifiedPurchase bool       `json:"verifiedPurchase"`
	HelpfulCount     int        `json:"helpfulCount"`
	HasVotedHelpful  bool       `json:"hasVotedHelpful,omitempty"` // Field for response

	SellerResponse   *string    `json:"sellerResponse,omitempty"`
	SellerResponseAt *time.Time `json:"sellerResponseAt,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Joined/Hydrated fields
	Photos  []string        `json:"photos,omitempty"`
	Buyer   *UserSummary    `json:"buyer,omitempty"`
	Listing *ListingSummary `json:"listing,omitempty"`
}

type UserSummary struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"` // Placeholder, ideally fetched from user service
	Avatar *string   `json:"avatar,omitempty"`
}

type ListingSummary struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}

// Stats API Model
type SellerStats struct {
	AverageRating   float64            `json:"averageRating"`
	TotalReviews    int64              `json:"totalReviews"`
	VerifiedReviews int64              `json:"verifiedReviews"`
	Distribution    map[string]int64   `json:"distribution"`
	Percentages     map[string]float64 `json:"percentages"`
	RecentTrend     *TrendStats        `json:"recentTrend,omitempty"`
}

type TrendStats struct {
	Last30Days StatisticPeriod `json:"last30Days"`
	Last90Days StatisticPeriod `json:"last90Days"`
}

type StatisticPeriod struct {
	Average float64 `json:"average"`
	Count   int64   `json:"count"`
}

// Request Models
type CreateReviewRequest struct {
	ListingID uuid.UUID `json:"listingId" binding:"required"`
	SellerID  uuid.UUID `json:"sellerId" binding:"required"`
	Rating    int       `json:"rating" binding:"required,min=1,max=5"`
	Title     string    `json:"title" binding:"max=255"`
	Comment   string    `json:"comment" binding:"max=2000"`
	Photos    []string  `json:"photos" binding:"max=5"` // URLs
}

type UpdateReviewRequest struct {
	Rating  int    `json:"rating" binding:"min=1,max=5"`
	Title   string `json:"title" binding:"max=255"`
	Comment string `json:"comment" binding:"max=2000"`
}

type SellerResponseRequest struct {
	Comment string `json:"comment" binding:"required,min=10,max=1000"`
}

// Pagination
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}
