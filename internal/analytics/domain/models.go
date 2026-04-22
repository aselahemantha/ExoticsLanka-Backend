package domain

import (
	"time"

	"github.com/google/uuid"
)

type ListingView struct {
	ID        uuid.UUID  `json:"id"`
	ListingID uuid.UUID  `json:"listingId"`
	UserID    *uuid.UUID `json:"userId,omitempty"`
	SessionID string     `json:"sessionId"`
	IPAddress string     `json:"ipAddress"`
	UserAgent string     `json:"userAgent"`
	Referrer  string     `json:"referrer"`
	EventType string     `json:"eventType"`
	CreatedAt time.Time  `json:"createdAt"`
}

type DealerAnalytics struct {
	ID                  uuid.UUID `json:"id"`
	DealerID            uuid.UUID `json:"dealerId"`
	Date                string    `json:"date"` // YYYY-MM-DD
	TotalViews          int       `json:"totalViews"`
	UniqueViewers       int       `json:"uniqueViewers"`
	TotalClicks         int       `json:"totalClicks"`
	TotalFavorites      int       `json:"totalFavorites"`
	TotalShares         int       `json:"totalShares"`
	TotalLeads          int       `json:"totalLeads"`
	TotalMessages       int       `json:"totalMessages"`
	PhoneReveals        int       `json:"phoneReveals"`
	TotalSales          int       `json:"totalSales"`
	TotalRevenue        float64   `json:"totalRevenue"`
	AvgResponseTimeMins *int      `json:"avgResponseTimeMins,omitempty"`
	ResponseRate        *float64  `json:"responseRate,omitempty"`
	InventoryCount      int       `json:"inventoryCount"`
	InventoryValue      float64   `json:"inventoryValue"`
	AvgHealthScore      int       `json:"avgHealthScore"`
	AvgDaysListed       int       `json:"avgDaysListed"`
	CreatedAt           time.Time `json:"createdAt"`
}

// Stats Objects for Dashboard API
type DashboardStats struct {
	Period      string          `json:"period"`
	Summary     SummaryStats    `json:"summary"`
	Engagement  EngagementStats `json:"engagement"`
	Conversions ConversionStats `json:"conversions"`
	Performance PerformStats    `json:"performance"`
}

type SummaryStats struct {
	TotalInventory   int     `json:"totalInventory"`
	TotalValue       float64 `json:"totalValue"`
	AvgDaysToSell    int     `json:"avgDaysToSell"`
	ResponseTimeAvg  int     `json:"responseTimeAvg"`
	ReviewScore      float64 `json:"reviewScore"`
	DepreciationRate float64 `json:"inventoryDepreciation"`
}

type EngagementStats struct {
	TotalViews      int     `json:"totalViews"`
	UniqueViewers   int     `json:"uniqueViewers"`
	TotalFavorites  int     `json:"totalFavorites"`
	TotalShares     int     `json:"totalShares"`
	ViewsChange     float64 `json:"viewsChange"`
	FavoritesChange float64 `json:"favoritesChange"`
}

type ConversionStats struct {
	TotalLeads     int     `json:"totalLeads"`
	TotalMessages  int     `json:"totalMessages"`
	PhoneReveals   int     `json:"phoneReveals"`
	TotalSales     int     `json:"totalSales"`
	ConversionRate float64 `json:"conversionRate"`
	LeadsChange    float64 `json:"leadsChange"`
}

type PerformStats struct {
	AvgHealthScore       int `json:"avgHealthScore"`
	ListingsAbove80      int `json:"listingsAbove80"`
	ListingsBelow60      int `json:"listingsBelow60"`
	PriceCompetitiveness int `json:"priceCompetitiveness"`
}

type Insight struct {
	Type     string         `json:"type"` // prediction, warning, opportunity, alert, tip
	Icon     string         `json:"icon"`
	Title    string         `json:"title"`
	Message  string         `json:"message"`
	Priority string         `json:"priority"` // high, warning, medium, urgent, low
	Action   *InsightAction `json:"action,omitempty"`
}

type InsightAction struct {
	Label string `json:"label"`
	Link  string `json:"link"`
}

// Request Objects
type TrackEventRequest struct {
	ListingID uuid.UUID `json:"listingId" binding:"required"`
	EventType string    `json:"eventType" binding:"required,oneof=view click contact_view phone_click share"`
	SessionID string    `json:"sessionId"`
	Referrer  string    `json:"referrer"`
}
