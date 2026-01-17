package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for data access
type Repository interface {
	// Listings
	CreateListing(ctx context.Context, listing *domain.CarListing) error
	GetListingByID(ctx context.Context, id uuid.UUID) (*domain.CarListing, error)
	UpdateListing(ctx context.Context, listing *domain.CarListing) error
	DeleteListing(ctx context.Context, id uuid.UUID) error
	GetListings(ctx context.Context, filter ListingFilter) ([]*domain.CarListing, int, error)
	UpdateListingStatus(ctx context.Context, id uuid.UUID, status string) error
	IncrementViews(ctx context.Context, id uuid.UUID) error

	// Specialized Getters
	GetFeaturedListings(ctx context.Context, limit int) ([]*domain.CarListing, error)
	GetTrendingListings(ctx context.Context, limit int) ([]*domain.CarListing, error)
	GetListingsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.CarListing, error)

	// Brands
	GetBrands(ctx context.Context) ([]*domain.CarBrand, error)
	CreateBrand(ctx context.Context, brand *domain.CarBrand) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository creates a new postgres repository
func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

// ListingFilter represents the filters for querying listings
type ListingFilter struct {
	Query         string
	Brands        []string
	MinPrice      float64
	MaxPrice      float64
	MinYear       int
	MaxYear       int
	MinMileage    int
	MaxMileage    int
	FuelTypes     []string
	Transmissions []string
	BodyTypes     []string
	Locations     []string
	Conditions    []string
	Status        string
	SortBy        string
	Page          int
	Limit         int
}

// CreateListing inserts a new listing into the database
func (r *postgresRepository) CreateListing(ctx context.Context, listing *domain.CarListing) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert listing
	query := `
		INSERT INTO car_listings (
			id, user_id, title, make, model, year, price, mileage, condition,
			transmission, fuel_type, body_type, color, doors, seats, engine_size, drivetrain,
			description, location, contact_phone, contact_email, status,
			health_score, is_new, is_featured, is_verified, trending,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22,
			$23, $24, $25, $26, $27,
			$28, $29
		)
	`
	_, err = tx.Exec(ctx, query,
		listing.ID, listing.UserID, listing.Title, listing.Make, listing.Model, listing.Year, listing.Price, listing.Mileage, listing.Condition,
		listing.Transmission, listing.FuelType, listing.BodyType, listing.Color, listing.Doors, listing.Seats, listing.EngineSize, listing.Drivetrain,
		listing.Description, listing.Location, listing.ContactPhone, listing.ContactEmail, listing.Status,
		listing.HealthScore, listing.IsNew, listing.IsFeatured, listing.IsVerified, listing.Trending,
		listing.CreatedAt, listing.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert listing: %w", err)
	}

	// Insert features
	if len(listing.Features) > 0 {
		for _, f := range listing.Features {
			_, err := tx.Exec(ctx, "INSERT INTO listing_features (listing_id, feature_name) VALUES ($1, $2)", listing.ID, f.FeatureName)
			if err != nil {
				return fmt.Errorf("failed to insert feature %s: %w", f.FeatureName, err)
			}
		}
	}

	return tx.Commit(ctx)
}

// GetListingByID retrieves a listing by its ID
func (r *postgresRepository) GetListingByID(ctx context.Context, id uuid.UUID) (*domain.CarListing, error) {
	query := `
		SELECT id, user_id, title, make, model, year, price, mileage, condition,
			   transmission, fuel_type, body_type, color, doors, seats, engine_size, drivetrain,
			   description, location, contact_phone, contact_email, status,
			   health_score, views, favorites_count, days_listed,
			   is_new, is_featured, is_verified, trending,
			   market_avg_price, price_alert,
			   created_at, updated_at, published_at, expires_at
		FROM car_listings
		WHERE id = $1
	`

	var l domain.CarListing
	err := r.db.QueryRow(ctx, query, id).Scan(
		&l.ID, &l.UserID, &l.Title, &l.Make, &l.Model, &l.Year, &l.Price, &l.Mileage, &l.Condition,
		&l.Transmission, &l.FuelType, &l.BodyType, &l.Color, &l.Doors, &l.Seats, &l.EngineSize, &l.Drivetrain,
		&l.Description, &l.Location, &l.ContactPhone, &l.ContactEmail, &l.Status,
		&l.HealthScore, &l.Views, &l.FavoritesCount, &l.DaysListed,
		&l.IsNew, &l.IsFeatured, &l.IsVerified, &l.Trending,
		&l.MarketAvgPrice, &l.PriceAlert,
		&l.CreatedAt, &l.UpdatedAt, &l.PublishedAt, &l.ExpiresAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Or custom ErrNotFound
		}
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}

	// Fetch features
	rows, err := r.db.Query(ctx, "SELECT id, feature_name FROM listing_features WHERE listing_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get features: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var f domain.ListingFeature
		f.ListingID = id
		if err := rows.Scan(&f.ID, &f.FeatureName); err != nil {
			return nil, err
		}
		l.Features = append(l.Features, f)
	}

	// Fetch images
	imgRows, err := r.db.Query(ctx, "SELECT id, image_url, thumbnail_url, is_cover, sort_order, created_at FROM listing_images WHERE listing_id = $1 ORDER BY sort_order ASC", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}
	defer imgRows.Close()

	for imgRows.Next() {
		var img domain.ListingImage
		img.ListingID = id
		if err := imgRows.Scan(&img.ID, &img.ImageURL, &img.ThumbnailURL, &img.IsCover, &img.SortOrder, &img.CreatedAt); err != nil {
			return nil, err
		}
		l.Images = append(l.Images, img)
	}

	return &l, nil
}

func (r *postgresRepository) UpdateListing(ctx context.Context, listing *domain.CarListing) error {
	query := `
        UPDATE car_listings SET
            price = $2, mileage = $3, description = $4, location = $5, 
            contact_phone = $6, contact_email = $7, condition = $8, updated_at = $9
        WHERE id = $1
    `
	_, err := r.db.Exec(ctx, query,
		listing.ID, listing.Price, listing.Mileage, listing.Description, listing.Location,
		listing.ContactPhone, listing.ContactEmail, listing.Condition, listing.UpdatedAt,
	)
	return err
}

func (r *postgresRepository) DeleteListing(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM car_listings WHERE id = $1", id)
	return err
}

func (r *postgresRepository) UpdateListingStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, "UPDATE car_listings SET status = $1 WHERE id = $2", status, id)
	return err
}

func (r *postgresRepository) IncrementViews(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE car_listings SET views = views + 1 WHERE id = $1", id)
	return err
}

func (r *postgresRepository) GetFeaturedListings(ctx context.Context, limit int) ([]*domain.CarListing, error) {
	if limit <= 0 {
		limit = 10
	}
	return r.getSimpleListings(ctx, "SELECT id, title, make, model, year, price, mileage, condition, location, status, health_score, created_at, is_new, is_featured, is_verified, trending FROM car_listings WHERE status = 'active' AND is_featured = TRUE ORDER BY created_at DESC LIMIT $1", limit)
}

func (r *postgresRepository) GetTrendingListings(ctx context.Context, limit int) ([]*domain.CarListing, error) {
	if limit <= 0 {
		limit = 10
	}
	return r.getSimpleListings(ctx, "SELECT id, title, make, model, year, price, mileage, condition, location, status, health_score, created_at, is_new, is_featured, is_verified, trending FROM car_listings WHERE status = 'active' AND trending = TRUE ORDER BY views DESC LIMIT $1", limit)
}

func (r *postgresRepository) GetListingsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.CarListing, error) {
	return r.getSimpleListings(ctx, "SELECT id, title, make, model, year, price, mileage, condition, location, status, health_score, created_at, is_new, is_featured, is_verified, trending FROM car_listings WHERE user_id = $1 ORDER BY created_at DESC", userID)
}

// Helper for simple listing queries (without deep associated data like features/images unless needed)
// For home screens etc we usually just need the cover image
func (r *postgresRepository) getSimpleListings(ctx context.Context, query string, args ...interface{}) ([]*domain.CarListing, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []*domain.CarListing
	for rows.Next() {
		var l domain.CarListing
		if err := rows.Scan(
			&l.ID, &l.Title, &l.Make, &l.Model, &l.Year, &l.Price, &l.Mileage, &l.Condition,
			&l.Location, &l.Status, &l.HealthScore, &l.CreatedAt, &l.IsNew, &l.IsFeatured, &l.IsVerified, &l.Trending,
		); err != nil {
			return nil, err
		}

		// Fetch cover image efficiently
		var coverImg string
		err = r.db.QueryRow(ctx, "SELECT image_url FROM listing_images WHERE listing_id = $1 AND is_cover = TRUE LIMIT 1", l.ID).Scan(&coverImg)
		if err == nil {
			l.Images = []domain.ListingImage{{ImageURL: coverImg, IsCover: true}}
		} else if err != pgx.ErrNoRows {
			// Log error but don't fail entire request?
		}

		listings = append(listings, &l)
	}
	return listings, nil
}

// GetListings implements advanced filtering and pagination
func (r *postgresRepository) GetListings(ctx context.Context, filter ListingFilter) ([]*domain.CarListing, int, error) {
	baseQuery := `
		SELECT id, title, make, model, year, price, mileage, condition, 
		       location, status, health_score, created_at, is_new, is_featured, is_verified, trending
		FROM car_listings
		WHERE status = 'active'
	`
	countQuery := `SELECT COUNT(*) FROM car_listings WHERE status = 'active'`

	var args []interface{}
	argCounter := 1
	var conditions []string

	if filter.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR make ILIKE $%d OR model ILIKE $%d)", argCounter, argCounter, argCounter))
		args = append(args, "%"+filter.Query+"%")
		argCounter++
	}

	if len(filter.Brands) > 0 {
		conditions = append(conditions, fmt.Sprintf("make = ANY($%d)", argCounter))
		args = append(args, filter.Brands)
		argCounter++
	}

	if filter.MinPrice > 0 {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", argCounter))
		args = append(args, filter.MinPrice)
		argCounter++
	}

	if filter.MaxPrice > 0 {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", argCounter))
		args = append(args, filter.MaxPrice)
		argCounter++
	}

	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Calculate total count first
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Sorting
	switch filter.SortBy {
	case "price_asc":
		baseQuery += " ORDER BY price ASC"
	case "price_desc":
		baseQuery += " ORDER BY price DESC"
	default:
		baseQuery += " ORDER BY created_at DESC"
	}

	// Pagination
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (filter.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	listings, err := r.getSimpleListings(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	return listings, total, nil
}

// Brands
func (r *postgresRepository) GetBrands(ctx context.Context) ([]*domain.CarBrand, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name, logo_url, listing_count FROM car_brands ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []*domain.CarBrand
	for rows.Next() {
		var b domain.CarBrand
		if err := rows.Scan(&b.ID, &b.Name, &b.LogoURL, &b.ListingCount); err != nil {
			return nil, err
		}
		brands = append(brands, &b)
	}
	return brands, nil
}

func (r *postgresRepository) CreateBrand(ctx context.Context, brand *domain.CarBrand) error {
	_, err := r.db.Exec(ctx, "INSERT INTO car_brands (name, logo_url) VALUES ($1, $2)", brand.Name, brand.LogoURL)
	return err
}
