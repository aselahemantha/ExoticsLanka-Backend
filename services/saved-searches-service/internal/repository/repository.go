package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/saved-searches-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	// Saved Search CRUD
	CreateSavedSearch(ctx context.Context, ss *domain.SavedSearch) (*domain.SavedSearch, error)
	GetSavedSearchByID(ctx context.Context, id, userID uuid.UUID) (*domain.SavedSearch, error)
	GetUserSavedSearches(ctx context.Context, userID uuid.UUID) ([]domain.SavedSearch, error)
	UpdateSavedSearch(ctx context.Context, ss *domain.SavedSearch) error
	DeleteSavedSearch(ctx context.Context, id, userID uuid.UUID) error
	UpdateAlertSettings(ctx context.Context, id, userID uuid.UUID, enabled bool, frequency string) error

	// Matching Logic
	CountMatchingListings(ctx context.Context, filters domain.SearchFilters) (int, error)
	GetNewListingsSince(ctx context.Context, filters domain.SearchFilters, since time.Time) ([]domain.ListingSummary, error)
	RunSearch(ctx context.Context, filters domain.SearchFilters, page, limit int) ([]domain.ListingSummary, int64, error)

	// Internal / Background
	UpdateMatchStats(ctx context.Context, id uuid.UUID, newCount, totalCount int) error
	GetTotalNewMatches(ctx context.Context, userID uuid.UUID) (int, []domain.SearchStats, error)
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

// --- Query Builder Helper ---

func buildFilterQuery(baseQuery string, filters domain.SearchFilters, params []interface{}, paramIdx int) (string, []interface{}, int) {
	// Search Query (Full Text - simplified for MVP using ILIKE)
	if filters.SearchQuery != "" {
		baseQuery += fmt.Sprintf(" AND (title ILIKE $%d OR make ILIKE $%d OR model ILIKE $%d)", paramIdx, paramIdx, paramIdx)
		params = append(params, "%"+filters.SearchQuery+"%")
		paramIdx++
	}

	// Brands
	if len(filters.Brands) > 0 {
		baseQuery += fmt.Sprintf(" AND make = ANY($%d)", paramIdx)
		params = append(params, filters.Brands)
		paramIdx++
	}

	// Price Range
	if filters.PriceRange[1] > 0 {
		baseQuery += fmt.Sprintf(" AND price >= $%d AND price <= $%d", paramIdx, paramIdx+1)
		params = append(params, filters.PriceRange[0], filters.PriceRange[1])
		paramIdx += 2
	}

	// Year Range
	if filters.YearRange[1] > 0 {
		baseQuery += fmt.Sprintf(" AND year >= $%d AND year <= $%d", paramIdx, paramIdx+1)
		params = append(params, filters.YearRange[0], filters.YearRange[1])
		paramIdx += 2
	}

	// Mileage Range
	if filters.MileageRange[1] > 0 {
		baseQuery += fmt.Sprintf(" AND mileage >= $%d AND mileage <= $%d", paramIdx, paramIdx+1)
		params = append(params, filters.MileageRange[0], filters.MileageRange[1])
		paramIdx += 2
	}

	// Fuel Types
	if len(filters.FuelTypes) > 0 {
		baseQuery += fmt.Sprintf(" AND fuel_type = ANY($%d)", paramIdx)
		params = append(params, filters.FuelTypes)
		paramIdx++
	}

	// Transmissions
	if len(filters.Transmissions) > 0 {
		baseQuery += fmt.Sprintf(" AND transmission = ANY($%d)", paramIdx)
		params = append(params, filters.Transmissions)
		paramIdx++
	}

	// Body Types
	if len(filters.BodyTypes) > 0 {
		baseQuery += fmt.Sprintf(" AND body_type = ANY($%d)", paramIdx)
		params = append(params, filters.BodyTypes)
		paramIdx++
	}

	// Locations
	if len(filters.Locations) > 0 {
		baseQuery += fmt.Sprintf(" AND location = ANY($%d)", paramIdx)
		params = append(params, filters.Locations)
		paramIdx++
	}

	// Condition (Assuming 'condition' column exists or mapped)
	if len(filters.Condition) > 0 {
		baseQuery += fmt.Sprintf(" AND condition = ANY($%d)", paramIdx)
		params = append(params, filters.Condition)
		paramIdx++
	}

	return baseQuery, params, paramIdx
}

// --- CRUD ---

func (r *postgresRepository) CreateSavedSearch(ctx context.Context, ss *domain.SavedSearch) (*domain.SavedSearch, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO saved_searches (
			user_id, name, filters, alert_enabled, alert_frequency, 
			total_matches, last_checked, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW(), NOW())
		RETURNING id, last_checked, created_at, updated_at
	`, ss.UserID, ss.Name, ss.Filters, ss.AlertEnabled, ss.AlertFrequency, ss.TotalMatches).Scan(
		&ss.ID, &ss.LastChecked, &ss.CreatedAt, &ss.UpdatedAt,
	)
	return ss, err
}

func (r *postgresRepository) GetSavedSearchByID(ctx context.Context, id, userID uuid.UUID) (*domain.SavedSearch, error) {
	var ss domain.SavedSearch
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, name, filters, alert_enabled, alert_frequency, 
		       last_checked, last_notified, new_matches_count, total_matches, created_at, updated_at
		FROM saved_searches WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(
		&ss.ID, &ss.UserID, &ss.Name, &ss.Filters, &ss.AlertEnabled, &ss.AlertFrequency,
		&ss.LastChecked, &ss.LastNotified, &ss.NewMatchesCount, &ss.TotalMatches, &ss.CreatedAt, &ss.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &ss, nil
}

func (r *postgresRepository) GetUserSavedSearches(ctx context.Context, userID uuid.UUID) ([]domain.SavedSearch, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, filters, alert_enabled, alert_frequency, 
		       last_checked, last_notified, new_matches_count, total_matches, created_at, updated_at
		FROM saved_searches WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var searches []domain.SavedSearch
	for rows.Next() {
		var ss domain.SavedSearch
		err := rows.Scan(
			&ss.ID, &ss.UserID, &ss.Name, &ss.Filters, &ss.AlertEnabled, &ss.AlertFrequency,
			&ss.LastChecked, &ss.LastNotified, &ss.NewMatchesCount, &ss.TotalMatches, &ss.CreatedAt, &ss.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		searches = append(searches, ss)
	}
	return searches, nil
}

func (r *postgresRepository) UpdateSavedSearch(ctx context.Context, ss *domain.SavedSearch) error {
	_, err := r.db.Exec(ctx, `
		UPDATE saved_searches 
		SET name = $1, filters = $2, updated_at = NOW()
		WHERE id = $3 AND user_id = $4
	`, ss.Name, ss.Filters, ss.ID, ss.UserID)
	return err
}

func (r *postgresRepository) DeleteSavedSearch(ctx context.Context, id, userID uuid.UUID) error {
	ct, err := r.db.Exec(ctx, "DELETE FROM saved_searches WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("not found or unauthorized")
	}
	return nil
}

func (r *postgresRepository) UpdateAlertSettings(ctx context.Context, id, userID uuid.UUID, enabled bool, frequency string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE saved_searches 
		SET alert_enabled = $1, alert_frequency = $2, updated_at = NOW()
		WHERE id = $3 AND user_id = $4
	`, enabled, frequency, id, userID)
	return err
}

func (r *postgresRepository) UpdateMatchStats(ctx context.Context, id uuid.UUID, newCount, totalCount int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE saved_searches 
		SET new_matches_count = $1, total_matches = $2, last_checked = NOW()
		WHERE id = $3
	`, newCount, totalCount, id)
	return err
}

// --- Matching Logic ---

func (r *postgresRepository) CountMatchingListings(ctx context.Context, filters domain.SearchFilters) (int, error) {
	query := "SELECT COUNT(*) FROM car_listings WHERE status = 'active'" // Assuming 'status' column exists
	params := []interface{}{}
	query, params, _ = buildFilterQuery(query, filters, params, 1)

	var count int
	err := r.db.QueryRow(ctx, query, params...).Scan(&count)
	return count, err
}

func (r *postgresRepository) GetNewListingsSince(ctx context.Context, filters domain.SearchFilters, since time.Time) ([]domain.ListingSummary, error) {
	query := `
		SELECT id, title, make, model, year, price, location, created_at,
		(SELECT image_url FROM listing_images WHERE listing_id = car_listings.id AND is_cover = TRUE LIMIT 1) as cover_image
		FROM car_listings 
		WHERE status = 'active'
	`
	params := []interface{}{}

	// Add since filter
	query += " AND created_at > $1"
	params = append(params, since)
	paramIdx := 2

	query, params, _ = buildFilterQuery(query, filters, params, paramIdx)

	query += " ORDER BY created_at DESC LIMIT 10" // Limit new matches preview

	rows, err := r.db.Query(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []domain.ListingSummary
	for rows.Next() {
		var l domain.ListingSummary
		err := rows.Scan(&l.ID, &l.Title, &l.Make, &l.Model, &l.Year, &l.Price, &l.Location, &l.CreatedAt, &l.CoverImage)
		if err != nil {
			return nil, err
		}
		listings = append(listings, l)
	}
	return listings, nil
}

func (r *postgresRepository) RunSearch(ctx context.Context, filters domain.SearchFilters, page, limit int) ([]domain.ListingSummary, int64, error) {
	// Count first
	countQuery := "SELECT COUNT(*) FROM car_listings WHERE status = 'active'"
	cParams := []interface{}{}
	countQuery, cParams, _ = buildFilterQuery(countQuery, filters, cParams, 1)

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, cParams...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fetch Data
	query := `
		SELECT id, title, make, model, year, price, location, created_at,
		(SELECT image_url FROM listing_images WHERE listing_id = car_listings.id AND is_cover = TRUE LIMIT 1) as cover_image
		FROM car_listings 
		WHERE status = 'active'
	`
	params := []interface{}{}
	paramIdx := 1
	query, params, paramIdx = buildFilterQuery(query, filters, params, paramIdx)

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
	params = append(params, limit, (page-1)*limit)

	rows, err := r.db.Query(ctx, query, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var listings []domain.ListingSummary
	for rows.Next() {
		var l domain.ListingSummary
		err := rows.Scan(&l.ID, &l.Title, &l.Make, &l.Model, &l.Year, &l.Price, &l.Location, &l.CreatedAt, &l.CoverImage)
		if err != nil {
			return nil, 0, err
		}
		listings = append(listings, l)
	}

	return listings, total, nil
}

func (r *postgresRepository) GetTotalNewMatches(ctx context.Context, userID uuid.UUID) (int, []domain.SearchStats, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name, new_matches_count FROM saved_searches WHERE user_id = $1 AND new_matches_count > 0", userID)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var total int
	var stats []domain.SearchStats

	for rows.Next() {
		var s domain.SearchStats
		if err := rows.Scan(&s.ID, &s.Name, &s.NewMatches); err != nil {
			return 0, nil, err
		}
		total += s.NewMatches
		stats = append(stats, s)
	}
	return total, stats, nil
}
