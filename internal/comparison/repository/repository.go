package repository

import (
	"context"

	"github.com/aselahemantha/exoticsLanka/internal/comparison/domain"
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
		SELECT EXISTS(SELECT 1 FROM listings WHERE id = $1 AND status = 'active')
	`, listingID).Scan(&exists)
	return exists, err
}

func (r *postgresRepository) GetComparisonItems(ctx context.Context, userID uuid.UUID) ([]domain.VehicleComparison, error) {
	// Join with car_listings and other tables to fetch details
	query := `
		SELECT 
			cl.id, cl.title,
			(SELECT url FROM listing_images WHERE listing_id = cl.id AND is_primary = TRUE LIMIT 1),
			cl.make, cl.model, cl.year, cl.price, cl.mileage,
			COALESCE(cl.transmission, ''), COALESCE(cl.fuel_type, ''), COALESCE(cl.body_type, ''), 
			COALESCE(cl.color, ''), COALESCE(cl.engine_size, ''), 
			COALESCE(cl.doors, 0), COALESCE(cl.seats, 0),
			COALESCE(cl.health_score, 0), -- Handle NULL health score
			(SELECT AVG(rating) FROM reviews WHERE seller_id = cl.seller_id), -- Approximate seller rating
			cl.features
		FROM comparison_items ci
		JOIN listings cl ON ci.listing_id = cl.id
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
			&v.HealthScore, &rating, &v.Features,
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

		// Features are already loaded in the main query from JSONB

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
