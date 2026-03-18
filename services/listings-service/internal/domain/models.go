package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Condition constants
const (
	ConditionNew               = "new"
	ConditionUsed              = "used"
	ConditionCertifiedPreOwned = "certified_pre_owned"
)

// Status constants
const (
	StatusActive = "active"
	StatusSold   = "sold"
	StatusDraft  = "draft"
)

// Report Reasons
const (
	ReasonMisleading    = "misleading"
	ReasonDuplicate     = "duplicate"
	ReasonSold          = "sold"
	ReasonSpam          = "spam"
	ReasonInappropriate = "inappropriate"
	ReasonOther         = "other"
)

// Report Status
const (
	ReportPending   = "pending"
	ReportReviewed  = "reviewed"
	ReportDismissed = "dismissed"
)

// -- Entities --

type CarListing struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	SellerID     uuid.UUID       `json:"sellerId" db:"seller_id"`
	Title        string          `json:"title" db:"title"`
	Make         string          `json:"make" db:"make"`
	Model        string          `json:"model" db:"model"`
	Year         int             `json:"year" db:"year"`
	Price        float64         `json:"price" db:"price"`
	Mileage      int             `json:"mileage" db:"mileage"`
	Condition    string          `json:"condition" db:"condition"`
	BodyType     *string         `json:"bodyType,omitempty" db:"body_type"`
	FuelType     *string         `json:"fuelType,omitempty" db:"fuel_type"`
	Transmission *string         `json:"transmission,omitempty" db:"transmission"`
	Color        *string         `json:"color,omitempty" db:"color"`
	Doors        *int            `json:"doors,omitempty" db:"doors"`
	Seats        *int            `json:"seats,omitempty" db:"seats"`
	EngineSize   *string         `json:"engineSize,omitempty" db:"engine_size"`
	Drivetrain   *string         `json:"drivetrain,omitempty" db:"drivetrain"`
	Features     json.RawMessage `json:"features,omitempty" db:"features"` // stored as jsonb array
	Description  *string         `json:"description,omitempty" db:"description"`
	Location     *string         `json:"location,omitempty" db:"location"`
	Status       string          `json:"status" db:"status"`
	IsFeatured   bool            `json:"isFeatured" db:"is_featured"`
	IsTrending   bool            `json:"isTrending" db:"is_trending"`
	Views        int             `json:"views" db:"views"`
	Favorites    int             `json:"favorites" db:"favorites"`
	HealthScore  int             `json:"healthScore" db:"health_score"`
	Verified     bool            `json:"verified" db:"verified"`
	CreatedAt    time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time       `json:"updatedAt" db:"updated_at"`

	// Associations populated separately
	Images []ListingImage `json:"images,omitempty" db:"-"`
}

type ListingImage struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ListingID uuid.UUID `json:"listingId" db:"listing_id"`
	URL       string    `json:"url" db:"url"`
	PublicID  string    `json:"publicId" db:"public_id"`
	IsPrimary bool      `json:"isPrimary" db:"is_primary"`
	Position  int       `json:"position" db:"position"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type ListingReport struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ListingID  uuid.UUID `json:"listingId" db:"listing_id"`
	ReporterID uuid.UUID `json:"reporterId" db:"reporter_id"`
	Reason     string    `json:"reason" db:"reason"`
	Details    *string   `json:"details,omitempty" db:"details"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

// -- API Request/Response DTOs --

type CreateListingRequest struct {
	Title        string   `json:"title" binding:"required"`
	Make         string   `json:"make" binding:"required"`
	Model        string   `json:"model" binding:"required"`
	Year         int      `json:"year" binding:"required"`
	Price        float64  `json:"price" binding:"required"`
	Mileage      int      `json:"mileage" binding:"required"`
	Condition    string   `json:"condition" binding:"required"`
	BodyType     *string  `json:"bodyType"`
	FuelType     *string  `json:"fuelType"`
	Transmission *string  `json:"transmission"`
	Color        *string  `json:"color"`
	Doors        *int     `json:"doors"`
	Seats        *int     `json:"seats"`
	EngineSize   *string  `json:"engineSize"`
	Drivetrain   *string  `json:"drivetrain"`
	Features     []string `json:"features"` // Passed as array, stored as json payload
	Description  *string  `json:"description"`
	Location     *string  `json:"location"`
	Images       []string `json:"images"` // Initial image URLs to associate
	Status       *string  `json:"status"` // Optional, defaults to active/draft based on logic
}

type UpdateListingRequest struct {
	Title        *string  `json:"title"`
	Make         *string  `json:"make"`
	Model        *string  `json:"model"`
	Year         *int     `json:"year"`
	Price        *float64 `json:"price"`
	Mileage      *int     `json:"mileage"`
	Condition    *string  `json:"condition"`
	BodyType     *string  `json:"bodyType"`
	FuelType     *string  `json:"fuelType"`
	Transmission *string  `json:"transmission"`
	Color        *string  `json:"color"`
	Doors        *int     `json:"doors"`
	Seats        *int     `json:"seats"`
	EngineSize   *string  `json:"engineSize"`
	Drivetrain   *string  `json:"drivetrain"`
	Features     []string `json:"features"` // Replace entirely if passed
	Description  *string  `json:"description"`
	Location     *string  `json:"location"`
	Status       *string  `json:"status"`
}

type ReportListingRequest struct {
	Reason  string  `json:"reason" binding:"required"`
	Details *string `json:"details"`
}

type ListingFilter struct {
	Query        string
	Brand        string
	Model        string
	Location     string
	PriceMin     float64
	PriceMax     float64
	YearMin      int
	YearMax      int
	MileageMin   int
	MileageMax   int
	FuelType     string
	Transmission string
	Condition    string
	IsFeatured   *bool
	IsTrending   *bool
	Sort         string
	Page         int
	Limit        int
	SellerID     *uuid.UUID // Used for /me/listings querying internally
}

type PaginatedListingsResponse struct {
	Listings   []*CarListing `json:"listings"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"totalPages"`
}

// -- Interfaces --

type ListingRepository interface {
	Create(ctx context.Context, listing *CarListing, images []string) error
	GetByID(ctx context.Context, id uuid.UUID) (*CarListing, error)
	GetListings(ctx context.Context, filter ListingFilter) ([]*CarListing, int, error)
	Update(ctx context.Context, listing *CarListing) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Stats
	IncrementViews(ctx context.Context, id uuid.UUID) error
	GetFeatured(ctx context.Context, limit int) ([]*CarListing, error)
	GetTrending(ctx context.Context, limit int) ([]*CarListing, error)

	// Discovery
	GetBrands(ctx context.Context) ([]string, error)
	GetModels(ctx context.Context, brand string) ([]string, error)

	// Favorites
	AddFavorite(ctx context.Context, userID, listingID uuid.UUID) error
	RemoveFavorite(ctx context.Context, userID, listingID uuid.UUID) error
	GetUserFavorites(ctx context.Context, userID uuid.UUID) ([]*CarListing, error)

	// Reports
	CreateReport(ctx context.Context, report *ListingReport) error
}

type ListingUseCase interface {
	CreateListing(ctx context.Context, userID uuid.UUID, req *CreateListingRequest) (*CarListing, error)
	GetListing(ctx context.Context, id uuid.UUID) (*CarListing, error)
	GetListings(ctx context.Context, filter ListingFilter) (*PaginatedListingsResponse, error)
	UpdateListing(ctx context.Context, userID, listingID uuid.UUID, req *UpdateListingRequest) (*CarListing, error)
	DeleteListing(ctx context.Context, userID, listingID uuid.UUID) error

	// Discovery
	GetFeaturedListings(ctx context.Context) ([]*CarListing, error)
	GetTrendingListings(ctx context.Context) ([]*CarListing, error)
	GetBrands(ctx context.Context) ([]string, error)
	GetModelsByBrand(ctx context.Context, brand string) ([]string, error)

	// User Interactions
	IncrementView(ctx context.Context, listingID uuid.UUID) error // Usually called async from GET
	ToggleFavorite(ctx context.Context, userID, listingID uuid.UUID, isFavorite bool) error
	GetUserFavorites(ctx context.Context, userID uuid.UUID) ([]*CarListing, error)
	ReportListing(ctx context.Context, userID, listingID uuid.UUID, req *ReportListingRequest) error
}
