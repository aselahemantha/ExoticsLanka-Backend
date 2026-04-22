package domain

import (
	"time"

	"github.com/google/uuid"
)

// Pagination meta data
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// Favorite represents a user's favorite listing
type Favorite struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	ListingID uuid.UUID `json:"listingId"`
	CreatedAt time.Time `json:"createdAt"`
	// Joined fields for response
	Listing *FavoriteListingDetails `json:"listing,omitempty"`
}

// FavoriteListingDetails contains a subset of listing data for the favorites list
type FavoriteListingDetails struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Make        string    `json:"make"`
	Model       string    `json:"model"`
	Year        int       `json:"year"`
	Price       float64   `json:"price"`
	Mileage     int       `json:"mileage"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	HealthScore int       `json:"healthScore"`
	Views       int       `json:"views"`
	DaysListed  int       `json:"daysListed"`
	CoverImage  string    `json:"coverImage"`
}

// API Models

type AddFavoriteResponse struct {
	ID        uuid.UUID `json:"id"`
	ListingID uuid.UUID `json:"listingId"`
	CreatedAt time.Time `json:"createdAt"`
}

type CheckFavoriteResponse struct {
	IsFavorited bool      `json:"isFavorited"`
	FavoritedAt time.Time `json:"favoritedAt,omitempty"`
}

type CountResponse struct {
	Count int64 `json:"count"`
}
