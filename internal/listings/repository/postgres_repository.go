package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aselahemantha/exoticsLanka/internal/listings/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresListingRepository struct {
	db *pgxpool.Pool
}

func NewPostgresListingRepository(db *pgxpool.Pool) domain.ListingRepository {
	return &postgresListingRepository{db: db}
}

// ----------------------------------------------------
// 1) BASE LISTING CRUD
// ----------------------------------------------------

func (r *postgresListingRepository) Create(ctx context.Context, l *domain.CarListing, images []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Create listing
	query := `
		INSERT INTO listings (
			id, seller_id, title, make, model, year, price, mileage, condition,
			body_type, fuel_type, transmission, color, doors, seats, engine_size,
			drivetrain, features, description, location, status, is_featured, is_trending,
			views, favorites, health_score, verified, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29
		)
	`
	_, err = tx.Exec(ctx, query,
		l.ID, l.SellerID, l.Title, l.Make, l.Model, l.Year, l.Price, l.Mileage, l.Condition,
		l.BodyType, l.FuelType, l.Transmission, l.Color, l.Doors, l.Seats, l.EngineSize,
		l.Drivetrain, l.Features, l.Description, l.Location, l.Status, l.IsFeatured, l.IsTrending,
		l.Views, l.Favorites, l.HealthScore, l.Verified, l.CreatedAt, l.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Insert associated initial images
	for i, url := range images {
		imgID := uuid.New()
		imgQuery := `INSERT INTO listing_images (id, listing_id, url, public_id, is_primary, position) VALUES ($1, $2, $3, $4, $5, $6)`
		isPrimary := i == 0
		_, err = tx.Exec(ctx, imgQuery, imgID, l.ID, url, "", isPrimary, i) // PublicID empty for initial image URLs
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *postgresListingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CarListing, error) {
	query := `
		SELECT 	id, seller_id, title, make, model, year, price, mileage, condition,
				body_type, fuel_type, transmission, color, doors, seats, engine_size,
				drivetrain, features, description, location, status, is_featured, is_trending,
				views, favorites, health_score, verified, created_at, updated_at
		FROM listings WHERE id = $1
	`
	var l domain.CarListing
	err := r.db.QueryRow(ctx, query, id).Scan(
		&l.ID, &l.SellerID, &l.Title, &l.Make, &l.Model, &l.Year, &l.Price, &l.Mileage, &l.Condition,
		&l.BodyType, &l.FuelType, &l.Transmission, &l.Color, &l.Doors, &l.Seats, &l.EngineSize,
		&l.Drivetrain, &l.Features, &l.Description, &l.Location, &l.Status, &l.IsFeatured, &l.IsTrending,
		&l.Views, &l.Favorites, &l.HealthScore, &l.Verified, &l.CreatedAt, &l.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Load images
	imgs, err := r.getImagesForListing(ctx, l.ID)
	if err != nil {
		return nil, err
	}
	l.Images = imgs

	return &l, nil
}

func (r *postgresListingRepository) Update(ctx context.Context, l *domain.CarListing) error {
	query := `
		UPDATE listings SET 
			title=$1, make=$2, model=$3, year=$4, price=$5, mileage=$6, condition=$7,
			body_type=$8, fuel_type=$9, transmission=$10, color=$11, doors=$12, seats=$13, engine_size=$14,
			drivetrain=$15, features=$16, description=$17, location=$18, status=$19, updated_at=NOW()
		WHERE id = $20
	`
	_, err := r.db.Exec(ctx, query,
		l.Title, l.Make, l.Model, l.Year, l.Price, l.Mileage, l.Condition,
		l.BodyType, l.FuelType, l.Transmission, l.Color, l.Doors, l.Seats, l.EngineSize,
		l.Drivetrain, l.Features, l.Description, l.Location, l.Status, l.ID,
	)
	return err
}

func (r *postgresListingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM listings WHERE id = $1", id)
	return err
}

// ----------------------------------------------------
// 2) DISCOVERY & SEARCH
// ----------------------------------------------------

func (r *postgresListingRepository) GetListings(ctx context.Context, f domain.ListingFilter) ([]*domain.CarListing, int, error) {
	// Base query
	baseQuery := "FROM listings WHERE 1=1"
	args := []interface{}{}
	argIdx := 1
	var conditions []string

	if f.SellerID != nil {
		conditions = append(conditions, fmt.Sprintf("seller_id = $%d", argIdx))
		args = append(args, *f.SellerID)
		argIdx++
	}
	if f.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR make ILIKE $%d OR model ILIKE $%d OR location ILIKE $%d)", argIdx, argIdx, argIdx, argIdx))
		args = append(args, "%"+f.Query+"%")
		argIdx++
	}
	if f.Brand != "" {
		conditions = append(conditions, fmt.Sprintf("make ILIKE $%d", argIdx))
		args = append(args, f.Brand)
		argIdx++
	}
	if f.Model != "" {
		conditions = append(conditions, fmt.Sprintf("model ILIKE $%d", argIdx))
		args = append(args, f.Model)
		argIdx++
	}
	if f.Location != "" {
		conditions = append(conditions, fmt.Sprintf("location ILIKE $%d", argIdx))
		args = append(args, "%"+f.Location+"%")
		argIdx++
	}
	if f.Condition != "" {
		conditions = append(conditions, fmt.Sprintf("condition = $%d", argIdx))
		args = append(args, f.Condition)
		argIdx++
	}
	if f.FuelType != "" {
		conditions = append(conditions, fmt.Sprintf("fuel_type ILIKE $%d", argIdx))
		args = append(args, f.FuelType)
		argIdx++
	}
	if f.Transmission != "" {
		conditions = append(conditions, fmt.Sprintf("transmission ILIKE $%d", argIdx))
		args = append(args, f.Transmission)
		argIdx++
	}
	if f.PriceMin > 0 {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", argIdx))
		args = append(args, f.PriceMin)
		argIdx++
	}
	if f.PriceMax > 0 {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", argIdx))
		args = append(args, f.PriceMax)
		argIdx++
	}
	if f.YearMin > 0 {
		conditions = append(conditions, fmt.Sprintf("year >= $%d", argIdx))
		args = append(args, f.YearMin)
		argIdx++
	}
	if f.YearMax > 0 {
		conditions = append(conditions, fmt.Sprintf("year <= $%d", argIdx))
		args = append(args, f.YearMax)
		argIdx++
	}
	if f.IsFeatured != nil {
		conditions = append(conditions, fmt.Sprintf("is_featured = $%d", argIdx))
		args = append(args, *f.IsFeatured)
		argIdx++
	}
	if f.IsTrending != nil {
		conditions = append(conditions, fmt.Sprintf("is_trending = $%d", argIdx))
		args = append(args, *f.IsTrending)
		argIdx++
	}

	// Always hide drafts unless specifically asking for seller's own listings
	if f.SellerID == nil {
		conditions = append(conditions, "status = 'active'")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// 1. Get total count
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 2. Fetch paginated models
	sortBy := "created_at DESC"
	switch f.Sort {
	case "price-asc":
		sortBy = "price ASC"
	case "price-desc":
		sortBy = "price DESC"
	case "year-asc":
		sortBy = "year ASC"
	case "year-desc":
		sortBy = "year DESC"
	case "mileage-asc":
		sortBy = "mileage ASC"
	case "mileage-desc":
		sortBy = "mileage DESC"
	}

	limit := f.Limit
	if limit == 0 {
		limit = 20
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	fetchQuery := fmt.Sprintf(`
		SELECT 	id, seller_id, title, make, model, year, price, mileage, condition,
				body_type, fuel_type, transmission, color, doors, seats, engine_size,
				drivetrain, features, description, location, status, is_featured, is_trending,
				views, favorites, health_score, verified, created_at, updated_at
		%s %s ORDER BY %s LIMIT %d OFFSET %d
	`, baseQuery, whereClause, sortBy, limit, offset)

	rows, err := r.db.Query(ctx, fetchQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var listings []*domain.CarListing
	for rows.Next() {
		var l domain.CarListing
		if err := rows.Scan(
			&l.ID, &l.SellerID, &l.Title, &l.Make, &l.Model, &l.Year, &l.Price, &l.Mileage, &l.Condition,
			&l.BodyType, &l.FuelType, &l.Transmission, &l.Color, &l.Doors, &l.Seats, &l.EngineSize,
			&l.Drivetrain, &l.Features, &l.Description, &l.Location, &l.Status, &l.IsFeatured, &l.IsTrending,
			&l.Views, &l.Favorites, &l.HealthScore, &l.Verified, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		// Typically we don't fully hydrate images arrays for list endpoints due to N+1 queries.
		// For a full spec impl we would either join lateral, or fetch them in one massive query post-loop.
		// For simplicity right now, we will leave Images empty unles queried individually by `GetByID`.

		listings = append(listings, &l)
	}

	return listings, total, nil
}

func (r *postgresListingRepository) GetBrands(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, "SELECT DISTINCT make FROM listings WHERE status = 'active' ORDER BY make")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []string
	for rows.Next() {
		var brand string
		if err := rows.Scan(&brand); err != nil {
			return nil, err
		}
		brands = append(brands, brand)
	}
	return brands, nil
}

func (r *postgresListingRepository) GetModels(ctx context.Context, brand string) ([]string, error) {
	rows, err := r.db.Query(ctx, "SELECT DISTINCT model FROM listings WHERE make ILIKE $1 AND status = 'active' ORDER BY model", brand)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []string
	for rows.Next() {
		var mod string
		if err := rows.Scan(&mod); err != nil {
			return nil, err
		}
		models = append(models, mod)
	}
	return models, nil
}

func (r *postgresListingRepository) GetFeatured(ctx context.Context, limit int) ([]*domain.CarListing, error) {
	f := true
	return r.getListingsShort(ctx, domain.ListingFilter{Limit: limit, IsFeatured: &f})
}

func (r *postgresListingRepository) GetTrending(ctx context.Context, limit int) ([]*domain.CarListing, error) {
	f := true
	return r.getListingsShort(ctx, domain.ListingFilter{Limit: limit, IsTrending: &f})
}

func (r *postgresListingRepository) getListingsShort(ctx context.Context, f domain.ListingFilter) ([]*domain.CarListing, error) {
	listings, _, err := r.GetListings(ctx, f)
	return listings, err
}

// ----------------------------------------------------
// 3) INTERACTIONS (Views, Favorites, Reports)
// ----------------------------------------------------

func (r *postgresListingRepository) IncrementViews(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE listings SET views = views + 1 WHERE id = $1", id)
	return err
}

func (r *postgresListingRepository) AddFavorite(ctx context.Context, userID, listingID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "INSERT INTO user_favorites (user_id, listing_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, listingID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE listings SET favorites = favorites + 1 WHERE id = $1", listingID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *postgresListingRepository) RemoveFavorite(ctx context.Context, userID, listingID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, "DELETE FROM user_favorites WHERE user_id = $1 AND listing_id = $2", userID, listingID)
	if err != nil {
		return err
	}

	if res.RowsAffected() > 0 {
		_, err = tx.Exec(ctx, "UPDATE listings SET favorites = GREATEST(favorites - 1, 0) WHERE id = $1", listingID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *postgresListingRepository) GetUserFavorites(ctx context.Context, userID uuid.UUID) ([]*domain.CarListing, error) {
	query := `
		SELECT 	l.id, l.seller_id, l.title, l.make, l.model, l.year, l.price, l.mileage, l.condition,
				l.body_type, l.fuel_type, l.transmission, l.color, l.doors, l.seats, l.engine_size,
				l.drivetrain, l.features, l.description, l.location, l.status, l.is_featured, l.is_trending,
				l.views, l.favorites, l.health_score, l.verified, l.created_at, l.updated_at
		FROM listings l
		JOIN user_favorites uf ON l.id = uf.listing_id
		WHERE uf.user_id = $1
		ORDER BY uf.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []*domain.CarListing
	for rows.Next() {
		var l domain.CarListing
		if err := rows.Scan(
			&l.ID, &l.SellerID, &l.Title, &l.Make, &l.Model, &l.Year, &l.Price, &l.Mileage, &l.Condition,
			&l.BodyType, &l.FuelType, &l.Transmission, &l.Color, &l.Doors, &l.Seats, &l.EngineSize,
			&l.Drivetrain, &l.Features, &l.Description, &l.Location, &l.Status, &l.IsFeatured, &l.IsTrending,
			&l.Views, &l.Favorites, &l.HealthScore, &l.Verified, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		listings = append(listings, &l)
	}

	return listings, nil
}

func (r *postgresListingRepository) CreateReport(ctx context.Context, report *domain.ListingReport) error {
	query := `
		INSERT INTO listing_reports (id, listing_id, reporter_id, reason, details, status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		report.ID, report.ListingID, report.ReporterID, report.Reason, report.Details, report.Status, report.CreatedAt, report.UpdatedAt,
	)
	return err
}

// ----------------------------------------------------
// 4) HELPER EXTRACIONS
// ----------------------------------------------------

func (r *postgresListingRepository) getImagesForListing(ctx context.Context, listingID uuid.UUID) ([]domain.ListingImage, error) {
	query := `SELECT id, listing_id, url, public_id, is_primary, position, created_at FROM listing_images WHERE listing_id = $1 ORDER BY position ASC`
	rows, err := r.db.Query(ctx, query, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []domain.ListingImage
	for rows.Next() {
		var img domain.ListingImage
		if err := rows.Scan(&img.ID, &img.ListingID, &img.URL, &img.PublicID, &img.IsPrimary, &img.Position, &img.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}
