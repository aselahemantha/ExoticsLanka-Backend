package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SavedSearch struct {
	ID              uuid.UUID     `json:"id"`
	UserID          uuid.UUID     `json:"userId,omitempty"`
	Name            string        `json:"name"`
	Filters         SearchFilters `json:"filters"`
	AlertEnabled    bool          `json:"alertEnabled"`
	AlertFrequency  string        `json:"alertFrequency"` // 'instant', 'daily', 'weekly'
	LastChecked     time.Time     `json:"lastChecked"`
	LastNotified    *time.Time    `json:"lastNotified,omitempty"`
	NewMatchesCount int           `json:"newMatchesCount"`
	TotalMatches    int           `json:"totalMatches"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

type SearchFilters struct {
	SearchQuery   string   `json:"searchQuery"`
	Brands        []string `json:"brands"`
	PriceRange    [2]int   `json:"priceRange"`   // [min, max]
	YearRange     [2]int   `json:"yearRange"`    // [min, max]
	MileageRange  [2]int   `json:"mileageRange"` // [min, max]
	FuelTypes     []string `json:"fuelTypes"`
	Transmissions []string `json:"transmissions"`
	BodyTypes     []string `json:"bodyTypes"`
	Locations     []string `json:"locations"`
	Condition     []string `json:"condition"`
}

// Helper to scan JSONB
func (s *SearchFilters) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal(value.([]byte), s)
}

// Request Models
type CreateSavedSearchRequest struct {
	Name           string        `json:"name" binding:"required"`
	Filters        SearchFilters `json:"filters" binding:"required"`
	AlertEnabled   *bool         `json:"alertEnabled"` // Pointer to distinguish false from zero value
	AlertFrequency string        `json:"alertFrequency"`
}

type UpdateSavedSearchRequest struct {
	Name    string        `json:"name"`
	Filters SearchFilters `json:"filters"`
}

type UpdateAlertsRequest struct {
	AlertEnabled   bool   `json:"alertEnabled"`
	AlertFrequency string `json:"alertFrequency"`
}

// Check/Run Response Models
type CheckMatchesResponse struct {
	NewMatches   int              `json:"newMatches"`
	TotalMatches int              `json:"totalMatches"`
	NewListings  []ListingSummary `json:"newListings"`
}

type RunSearchResponse struct {
	Listings   []ListingSummary `json:"listings"`
	Pagination Pagination       `json:"pagination"`
}

type ListingSummary struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Make       string    `json:"make,omitempty"`
	Model      string    `json:"model,omitempty"`
	Year       int       `json:"year,omitempty"`
	Price      float64   `json:"price"`
	Location   string    `json:"location,omitempty"`
	CoverImage *string   `json:"coverImage,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type NewMatchesResponse struct {
	TotalNewMatches int           `json:"totalNewMatches"`
	BySearch        []SearchStats `json:"bySearch"`
}

type SearchStats struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	NewMatches int       `json:"newMatches"`
}
