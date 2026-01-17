package domain

import (
	"time"

	"github.com/google/uuid"
)

// Condition constants
const (
	ConditionBrandNew      = "Brand New"
	ConditionUsedLikeNew   = "Used - Like New"
	ConditionUsedExcellent = "Used - Excellent"
	ConditionUsedGood      = "Used - Good"
	ConditionUsedFair      = "Used - Fair"
)

// Status constants
const (
	StatusDraft    = "draft"
	StatusPending  = "pending"
	StatusActive   = "active"
	StatusSold     = "sold"
	StatusExpired  = "expired"
	StatusRejected = "rejected"
)

// CarListing represents a car listing in the system
type CarListing struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"userId" db:"user_id"`
	Title          string     `json:"title" db:"title"`
	Make           string     `json:"make" db:"make"`
	Model          string     `json:"model" db:"model"`
	Year           int        `json:"year" db:"year"`
	Price          float64    `json:"price" db:"price"`
	Mileage        int        `json:"mileage" db:"mileage"`
	Condition      string     `json:"condition" db:"condition"`
	Transmission   *string    `json:"transmission,omitempty" db:"transmission"`
	FuelType       *string    `json:"fuelType,omitempty" db:"fuel_type"`
	BodyType       *string    `json:"bodyType,omitempty" db:"body_type"`
	Color          *string    `json:"color,omitempty" db:"color"`
	Doors          *int       `json:"doors,omitempty" db:"doors"`
	Seats          *int       `json:"seats,omitempty" db:"seats"`
	EngineSize     *string    `json:"engineSize,omitempty" db:"engine_size"`
	Drivetrain     *string    `json:"drivetrain,omitempty" db:"drivetrain"`
	Description    *string    `json:"description,omitempty" db:"description"`
	Location       string     `json:"location" db:"location"`
	ContactPhone   *string    `json:"contactPhone,omitempty" db:"contact_phone"`
	ContactEmail   *string    `json:"contactEmail,omitempty" db:"contact_email"`
	Status         string     `json:"status" db:"status"`
	HealthScore    int        `json:"healthScore" db:"health_score"`
	Views          int        `json:"views" db:"views"`
	FavoritesCount int        `json:"favoritesCount" db:"favorites_count"`
	DaysListed     int        `json:"daysListed" db:"days_listed"`
	IsNew          bool       `json:"isNew" db:"is_new"`
	IsFeatured     bool       `json:"isFeatured" db:"is_featured"`
	IsVerified     bool       `json:"isVerified" db:"is_verified"`
	Trending       bool       `json:"trending" db:"trending"`
	MarketAvgPrice *float64   `json:"marketAvgPrice,omitempty" db:"market_avg_price"`
	PriceAlert     *string    `json:"priceAlert,omitempty" db:"price_alert"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at"`
	PublishedAt    *time.Time `json:"publishedAt,omitempty" db:"published_at"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty" db:"expires_at"`

	// Associations - populated separately
	Images   []ListingImage   `json:"images,omitempty" db:"-"`
	Features []ListingFeature `json:"features,omitempty" db:"-"`
	User     *User            `json:"user,omitempty" db:"-"` // Basic user info
}

// ListingImage represents an image for a listing
type ListingImage struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ListingID    uuid.UUID `json:"listingId" db:"listing_id"`
	ImageURL     string    `json:"imageUrl" db:"image_url"`
	ThumbnailURL *string   `json:"thumbnailUrl,omitempty" db:"thumbnail_url"`
	IsCover      bool      `json:"isCover" db:"is_cover"`
	SortOrder    int       `json:"sortOrder" db:"sort_order"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

// ListingFeature represents a feature of a car
type ListingFeature struct {
	ID          int       `json:"id" db:"id"`
	ListingID   uuid.UUID `json:"listingId" db:"listing_id"`
	FeatureName string    `json:"featureName" db:"feature_name"`
}

// CarBrand represents a car brand
type CarBrand struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	LogoURL      *string   `json:"logoUrl,omitempty" db:"logo_url"`
	ListingCount int       `json:"listingCount" db:"listing_count"`
	IsActive     bool      `json:"isActive" db:"is_active"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// API Models

// CreateListingRequest
type CreateListingRequest struct {
	Title        string   `json:"title" binding:"required,min=10,max=255"`
	Make         string   `json:"make" binding:"required,max=100"`
	Model        string   `json:"model" binding:"required,max=100"`
	Year         int      `json:"year" binding:"required,min=1900,max=2030"`
	Price        float64  `json:"price" binding:"required,min=0"`
	Mileage      int      `json:"mileage" binding:"required,min=0"`
	Condition    string   `json:"condition" binding:"required"`
	Transmission *string  `json:"transmission"`
	FuelType     *string  `json:"fuelType"`
	BodyType     *string  `json:"bodyType"`
	Color        *string  `json:"color"`
	Doors        *int     `json:"doors"`
	Seats        *int     `json:"seats"`
	EngineSize   *string  `json:"engineSize"`
	Drivetrain   *string  `json:"drivetrain"`
	Description  *string  `json:"description" binding:"max=5000"`
	Location     string   `json:"location" binding:"required"`
	ContactPhone *string  `json:"contactPhone"`
	ContactEmail *string  `json:"contactEmail"`
	Features     []string `json:"features"`
}

// UpdateListingRequest
type UpdateListingRequest struct {
	Price        *float64 `json:"price"`
	Mileage      *int     `json:"mileage"`
	Description  *string  `json:"description"`
	Location     *string  `json:"location"`
	ContactPhone *string  `json:"contactPhone"`
	ContactEmail *string  `json:"contactEmail"`
	Condition    *string  `json:"condition"`
	Features     []string `json:"features"` // Replaces all features if provided
}

// User struct for embedding user info in response
type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Role      string    `json:"role"`
	Verified  bool      `json:"verified"`
	AvatarURL string    `json:"avatarUrl,omitempty"`
}
