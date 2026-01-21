package domain

import (
	"time"

	"github.com/google/uuid"
)

type ComparisonItem struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	ListingID uuid.UUID `json:"listingId"`
	CreatedAt time.Time `json:"createdAt"`
}

type ComparisonData struct {
	Vehicles   []VehicleComparison `json:"vehicles"`
	Comparison ComparisonAnalysis  `json:"comparison"`
}

type VehicleComparison struct {
	ID           uuid.UUID              `json:"id"`
	Title        string                 `json:"title"`
	Image        string                 `json:"image"`
	Specs        map[string]interface{} `json:"specs"` // Flat map for flexibility specs
	Features     []string               `json:"features"`
	HealthScore  int                    `json:"healthScore"`
	SellerRating *float64               `json:"sellerRating"`
}

type ComparisonAnalysis struct {
	PriceRange     RangeData           `json:"priceRange"`
	YearRange      RangeData           `json:"yearRange"`
	MileageRange   RangeData           `json:"mileageRange"`
	CommonFeatures []string            `json:"commonFeatures"`
	UniqueFeatures map[string][]string `json:"uniqueFeatures"`
}

type RangeData struct {
	Lowest     float64 `json:"lowest"`
	Highest    float64 `json:"highest"`
	Difference float64 `json:"difference"`
}

type MinimalListing struct {
	ID      uuid.UUID `json:"id"`
	Title   string    `json:"title"`
	Make    string    `json:"make"`
	Model   string    `json:"model"`
	Year    int       `json:"year"`
	Price   float64   `json:"price"`
	Mileage int       `json:"mileage"`
	Image   string    `json:"image"`
}

type ComparisonListResponse struct {
	Items    []MinimalListing `json:"items"`
	Count    int              `json:"count"`
	MaxItems int              `json:"maxItems"`
}
