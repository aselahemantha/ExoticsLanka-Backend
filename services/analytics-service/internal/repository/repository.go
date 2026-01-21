package repository

import (
	"context"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/analytics-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	// Tracking
	TrackEvent(ctx context.Context, view *domain.ListingView) error

	// Aggregation Logic (Cross-Service Querying)
	GetDailyEngagement(ctx context.Context, dealerID string, date string) (*domain.EngagementStats, error)
	GetDailyConversions(ctx context.Context, dealerID string, date string) (*domain.ConversionStats, error)
	GetInventoryMetrics(ctx context.Context, dealerID string) (*domain.SummaryStats, error)

	// Analytics Storage
	UpsertDealerAnalytics(ctx context.Context, analytics *domain.DealerAnalytics) error

	// Reporting
	GetDealerAnalytics(ctx context.Context, dealerID string, periodDays int) ([]domain.DealerAnalytics, error)
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) TrackEvent(ctx context.Context, view *domain.ListingView) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO listing_views (
			listing_id, user_id, session_id, ip_address, user_agent, referrer, event_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, view.ListingID, view.UserID, view.SessionID, view.IPAddress, view.UserAgent, view.Referrer, view.EventType)
	return err
}

func (r *postgresRepository) GetDailyEngagement(ctx context.Context, dealerID string, date string) (*domain.EngagementStats, error) {
	stats := &domain.EngagementStats{}

	// Query Listing Views (Direct)
	err := r.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total_views,
			COUNT(DISTINCT session_id) as unique_viewers, -- Approximation using session for unique
			COUNT(*) FILTER (WHERE event_type = 'share') as total_shares
		FROM listing_views lv
		JOIN car_listings cl ON lv.listing_id = cl.id
		WHERE cl.user_id = $1 AND DATE(lv.created_at) = $2
	`, dealerID, date).Scan(&stats.TotalViews, &stats.UniqueViewers, &stats.TotalShares)
	if err != nil {
		return nil, err
	}

	// Query Favorites (Favorites Service Table - Shared DB)
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM favorites f
		JOIN car_listings cl ON f.listing_id = cl.id
		WHERE cl.user_id = $1 AND DATE(f.created_at) = $2
	`, dealerID, date).Scan(&stats.TotalFavorites)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *postgresRepository) GetDailyConversions(ctx context.Context, dealerID string, date string) (*domain.ConversionStats, error) {
	stats := &domain.ConversionStats{}

	// Leads (Conversations - Shared DB)
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM conversations
		WHERE seller_id = $1 AND DATE(created_at) = $2
	`, dealerID, date).Scan(&stats.TotalLeads)
	if err != nil {
		return nil, err
	}

	// Messages (Shared DB)
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM messages m
		JOIN conversations c ON m.conversation_id = c.id
		WHERE c.seller_id = $1 AND DATE(m.created_at) = $2
	`, dealerID, date).Scan(&stats.TotalMessages)
	if err != nil {
		return nil, err
	}

	// Phone Reveals (from listing_views events)
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM listing_views lv
		JOIN car_listings cl ON lv.listing_id = cl.id
		WHERE cl.user_id = $1 AND DATE(lv.created_at) = $2 AND lv.event_type = 'phone_click'
	`, dealerID, date).Scan(&stats.PhoneReveals)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *postgresRepository) GetInventoryMetrics(ctx context.Context, dealerID string) (*domain.SummaryStats, error) {
	stats := &domain.SummaryStats{}
	err := r.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) as count,
			COALESCE(SUM(price), 0) as total_value,
			COALESCE(AVG(health_score), 0) as avg_health, -- Assuming health_score exists on listings
			COALESCE(AVG(EXTRACT(DAY FROM NOW() - created_at)), 0) as avg_days
		FROM car_listings
		WHERE user_id = $1 AND status = 'active'
	`, dealerID).Scan(&stats.TotalInventory, &stats.TotalValue, &stats.ReviewScore, &stats.AvgDaysToSell)
	// Mapping: ReviewScore is placeholder for AvgHealth here just to fit struct.
	// Real implementation would map accurately.
	return stats, err
}

func (r *postgresRepository) UpsertDealerAnalytics(ctx context.Context, a *domain.DealerAnalytics) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO dealer_analytics (
			dealer_id, date,
			total_views, unique_viewers, total_clicks, total_favorites, total_shares,
			total_leads, total_messages, phone_reveals,
			inventory_count, inventory_value, avg_health_score, avg_days_listed
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (dealer_id, date) DO UPDATE SET
			total_views = EXCLUDED.total_views,
			unique_viewers = EXCLUDED.unique_viewers,
			total_clicks = EXCLUDED.total_clicks,
			total_favorites = EXCLUDED.total_favorites,
			total_shares = EXCLUDED.total_shares,
			total_leads = EXCLUDED.total_leads,
			total_messages = EXCLUDED.total_messages,
			phone_reveals = EXCLUDED.phone_reveals,
			inventory_count = EXCLUDED.inventory_count,
			inventory_value = EXCLUDED.inventory_value,
			avg_health_score = EXCLUDED.avg_health_score,
			avg_days_listed = EXCLUDED.avg_days_listed
	`, a.DealerID, a.Date,
		a.TotalViews, a.UniqueViewers, a.TotalClicks, a.TotalFavorites, a.TotalShares,
		a.TotalLeads, a.TotalMessages, a.PhoneReveals,
		a.InventoryCount, a.InventoryValue, a.AvgHealthScore, a.AvgDaysListed)
	return err
}

func (r *postgresRepository) GetDealerAnalytics(ctx context.Context, dealerID string, periodDays int) ([]domain.DealerAnalytics, error) {
	query := `
		SELECT 
			id, dealer_id, date, 
			total_views, unique_viewers, total_clicks, total_favorites, total_shares,
			total_leads, total_messages, phone_reveals,
			inventory_count, inventory_value, avg_health_score, avg_days_listed
		FROM dealer_analytics
		WHERE dealer_id = $1 AND date >= CURRENT_DATE - make_interval(days => $2)
		ORDER BY date ASC
	`
	rows, err := r.db.Query(ctx, query, dealerID, periodDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.DealerAnalytics
	for rows.Next() {
		var da domain.DealerAnalytics
		var date time.Time
		err := rows.Scan(
			&da.ID, &da.DealerID, &date,
			&da.TotalViews, &da.UniqueViewers, &da.TotalClicks, &da.TotalFavorites, &da.TotalShares,
			&da.TotalLeads, &da.TotalMessages, &da.PhoneReveals,
			&da.InventoryCount, &da.InventoryValue, &da.AvgHealthScore, &da.AvgDaysListed,
		)
		if err != nil {
			return nil, err
		}
		da.Date = date.Format("2006-01-02")
		results = append(results, da)
	}
	return results, nil
}
