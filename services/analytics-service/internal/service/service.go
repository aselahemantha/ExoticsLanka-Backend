package service

import (
	"context"
	"fmt"

	"github.com/aselahemantha/exoticsLanka/services/analytics-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/analytics-service/internal/repository"
	"github.com/google/uuid"
)

type Service interface {
	TrackEvent(ctx context.Context, req domain.TrackEventRequest, metadata map[string]interface{}) error
	GetDashboard(ctx context.Context, dealerID uuid.UUID, period string) (*domain.DashboardStats, error)
	GenerateInsights(ctx context.Context, dealerID uuid.UUID) ([]domain.Insight, error)
	GetInventoryPerformance(ctx context.Context, dealerID uuid.UUID) (map[string]interface{}, error)
	RunDailyAggregation(ctx context.Context, dealerID uuid.UUID, date string) error
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) TrackEvent(ctx context.Context, req domain.TrackEventRequest, metadata map[string]interface{}) error {
	ip, _ := metadata["ipAddress"].(string)
	userAgent, _ := metadata["userAgent"].(string)
	// sessionID could be from req or metadata/cookie
	sessionID := req.SessionID

	view := &domain.ListingView{
		ListingID: req.ListingID,
		SessionID: sessionID,
		IPAddress: ip,
		UserAgent: userAgent,
		Referrer:  req.Referrer,
		EventType: req.EventType,
	}

	if uid, ok := metadata["userId"].(uuid.UUID); ok {
		view.UserID = &uid
	}

	return s.repo.TrackEvent(ctx, view)
}

func (s *service) GetDashboard(ctx context.Context, dealerID uuid.UUID, period string) (*domain.DashboardStats, error) {
	// Determine days
	days := 30
	if period == "7d" {
		days = 7
	}
	if period == "90d" {
		days = 90
	}

	// Fetch historical analytics
	history, err := s.repo.GetDealerAnalytics(ctx, dealerID.String(), days)
	if err != nil {
		return nil, err
	}

	// Fetch current inventory snapshot
	invStats, err := s.repo.GetInventoryMetrics(ctx, dealerID.String())
	if err != nil {
		return nil, err
	}

	// Aggregate History
	var totalViews, uniqueViewers, totalFavorites, totalShares int
	var totalLeads, totalMessages, phoneReveals, totalSales int
	// For change calculation, we'd need previous period using separate query or slice logic.
	// MVP: Simple sum.

	for _, day := range history {
		totalViews += day.TotalViews
		uniqueViewers += day.UniqueViewers // Summing uniques across days is technically wrong distinct count, but standard approximation for aggregate tables
		totalFavorites += day.TotalFavorites
		totalShares += day.TotalShares
		totalLeads += day.TotalLeads
		totalMessages += day.TotalMessages
		phoneReveals += day.PhoneReveals
		totalSales += day.TotalSales
	}

	stats := &domain.DashboardStats{
		Period:  period,
		Summary: *invStats, // Use real-time inventory stats for summary
		Engagement: domain.EngagementStats{
			TotalViews:     totalViews,
			UniqueViewers:  uniqueViewers,
			TotalFavorites: totalFavorites,
			TotalShares:    totalShares,
			// Change % would require querying previous period. Skipping for MVP.
		},
		Conversions: domain.ConversionStats{
			TotalLeads:    totalLeads,
			TotalMessages: totalMessages,
			PhoneReveals:  phoneReveals,
			TotalSales:    totalSales,
		},
		Performance: domain.PerformStats{
			AvgHealthScore: int(invStats.ReviewScore), // Check mapping in repo
			// Others stubbed
		},
	}

	return stats, nil
}

func (s *service) RunDailyAggregation(ctx context.Context, dealerID uuid.UUID, date string) error {
	// 1. Get Engagement
	eng, err := s.repo.GetDailyEngagement(ctx, dealerID.String(), date)
	if err != nil {
		return fmt.Errorf("engagement agg failed: %v", err)
	}

	// 2. Get Conversions
	conv, err := s.repo.GetDailyConversions(ctx, dealerID.String(), date)
	if err != nil {
		return fmt.Errorf("conversion agg failed: %v", err)
	}

	// 3. Get Inventory Snapshot
	inv, err := s.repo.GetInventoryMetrics(ctx, dealerID.String())
	if err != nil {
		return fmt.Errorf("inventory agg failed: %v", err)
	}

	// 4. Upsert
	analytics := &domain.DealerAnalytics{
		DealerID:       dealerID,
		Date:           date,
		TotalViews:     eng.TotalViews,
		UniqueViewers:  eng.UniqueViewers,
		TotalShares:    eng.TotalShares,
		TotalFavorites: eng.TotalFavorites,

		TotalLeads:    conv.TotalLeads,
		TotalMessages: conv.TotalMessages,
		PhoneReveals:  conv.PhoneReveals,

		InventoryCount: inv.TotalInventory,
		InventoryValue: inv.TotalValue,
		AvgDaysListed:  int(inv.AvgDaysToSell),
		// AvgHealthScore: ... mapping ?
	}

	return s.repo.UpsertDealerAnalytics(ctx, analytics)
}

func (s *service) GenerateInsights(ctx context.Context, dealerID uuid.UUID) ([]domain.Insight, error) {
	// Mock Rule Engine for MVP logic
	insights := []domain.Insight{}

	// Real implementation would query DB for specific conditions like "Overpriced Listings"
	// For example, calling a repo method GetOverpricedListings(dealerID).
	// Here we return the static examples or stub logic to prove architectural point.

	insights = append(insights, domain.Insight{
		Type:     "tip",
		Icon:     "photo",
		Title:    "Listing Quality",
		Message:  "2 listings are missing interior photos. Adding them increases views by 40%.",
		Priority: "low",
	})

	return insights, nil
}

func (s *service) GetInventoryPerformance(ctx context.Context, dealerID uuid.UUID) (map[string]interface{}, error) {
	// Return breakdown stats
	// MVP: Just reuse summary or mock breakdown
	metrics, err := s.repo.GetInventoryMetrics(ctx, dealerID.String())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"summary":          metrics,
		"status_breakdown": map[string]int{"active": metrics.TotalInventory, "sold": 5}, // stub
	}, nil
}
