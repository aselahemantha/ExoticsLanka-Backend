package repository

import (
	"context"

	"github.com/aselahemantha/exoticsLanka/services/comparison-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	AddToComparison(ctx context.Context, userID, listingID uuid.UUID) error
	RemoveFromComparison(ctx context.Context, userID, listingID uuid.UUID) error
	ClearComparison(ctx context.Context, userID uuid.UUID) error
	GetComparisonCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetComparisonItems(ctx context.Context, userID uuid.UUID) ([]domain.VehicleComparison, error)
	IsListingInComparison(ctx context.Context, userID, listingID uuid.UUID) (bool, error)
	CheckListingExists(ctx context.Context, listingID uuid.UUID) (bool, error)
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) AddToComparison(ctx context.Context, userID, listingID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO comparison_items (user_id, listing_id) 
		VALUES ($1, $2)
		ON CONFLICT (user_id, listing_id) DO NOTHING
	`, userID, listingID)
	return err
}

func (r *postgresRepository) RemoveFromComparison(ctx context.Context, userID, listingID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM comparison_items 
		WHERE user_id = $1 AND listing_id = $2
	`, userID, listingID)
	return err
}

func (r *postgresRepository) ClearComparison(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM comparison_items WHERE user_id = $1`, userID)
	return err
}

func (r *postgresRepository) GetComparisonCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM comparison_items WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func (r *postgresRepository) IsListingInComparison(ctx context.Context, userID, listingID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM comparison_items WHERE user_id = $1 AND listing_id = $2)
	`, userID, listingID).Scan(&exists)
	return exists, err
}

func (r *postgresRepository) CheckListingExists(ctx context.Context, listingID uuid.UUID) (bool, error) {
	var exists bool
	// Checked against car_listings shared table
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM car_listings WHERE id = $1 AND status = 'active')
	`, listingID).Scan(&exists)
	return exists, err
}

func (r *postgresRepository) GetComparisonItems(ctx context.Context, userID uuid.UUID) ([]domain.VehicleComparison, error) {
	// Join with car_listings and other tables to fetch details
	query := `
		SELECT 
			cl.id, cl.title,
			(SELECT image_url FROM listing_images WHERE listing_id = cl.id AND is_cover = TRUE LIMIT 1),
			cl.make, cl.model, cl.year, cl.price, cl.mileage,
			cl.transmission, cl.fuel_type, cl.body_type, cl.color, cl.engine_size, cl.doors, cl.seats,
			COALESCE(cl.health_score, 0), -- Handle NULL health score
			(SELECT AVG(rating) FROM reviews WHERE seller_id = cl.user_id) -- Approximate seller rating
		FROM comparison_items ci
		JOIN car_listings cl ON ci.listing_id = cl.id
		WHERE ci.user_id = $1
		ORDER BY ci.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []domain.VehicleComparison
	for rows.Next() {
		var (
			v                                             domain.VehicleComparison
			make, model, trans, fuel, body, color, engine string
			year, mileage, doors, seats                   int
			price                                         float64
			rating                                        *float64
			image                                         *string
		)

		err := rows.Scan(
			&v.ID, &v.Title, &image,
			&make, &model, &year, &price, &mileage,
			&trans, &fuel, &body, &color, &engine, &doors, &seats,
			&v.HealthScore, &rating,
		)
		if err != nil {
			return nil, err
		}

		if image != nil {
			v.Image = *image
		}
		v.SellerRating = rating

		// Build specs map
		v.Specs = map[string]interface{}{
			"make": make, "model": model, "year": year, "price": price, "mileage": mileage,
			"transmission": trans, "fuelType": fuel, "bodyType": body, "color": color,
			"engineSize": engine, "doors": doors, "seats": seats,
		}

		// Fetch features separately or use array agg in main query (using separate here for simplicity/cleanliness in Go)
		// Or update main query to JSON agg features. Let's do secondary query for features for now or just array_agg in valid SQL.
		// Optimized approach: array_agg in main query.
		// Updating main query is better but let's do a sub-query loop for MVP simplicity or assume empty features if complex.
		// Let's execute feature query per item (N+1 but N<=4)
		features, _ := r.getFeatures(ctx, v.ID)
		v.Features = features

		vehicles = append(vehicles, v)
	}
	return vehicles, nil
}

func (r *postgresRepository) getFeatures(ctx context.Context, listingID uuid.UUID) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT feature_name FROM listing_features WHERE listing_id = $1`, listingID)
	if err != nil {
		return []string{}, nil
	}
	defer rows.Close()
	var feats []string
	for rows.Next() {
		var f string
		if err := rows.Scan(&f); err == nil {
			feats = append(feats, f)
		}
	}
	return feats, nil
}
