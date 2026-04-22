package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aselahemantha/exoticsLanka/internal/analytics/domain"
	"github.com/aselahemantha/exoticsLanka/internal/analytics/service"
	"github.com/aselahemantha/exoticsLanka/internal/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.Engine, auth common.AuthMiddleware) {
	api := r.Group("/api/analytics")
	{
		// Optional Auth for Tracking
		api.POST("/track", auth.OptionalHandle(), h.Track)

		// Protected Dashboard/Insights
		protected := api.Group("")
		protected.Use(auth.Handle())
		{
			protected.GET("/overview", h.GetOverview)
			protected.GET("/insights", h.GetInsights)
			protected.GET("/inventory", h.GetInventoryStats)
			protected.POST("/jobs/aggregate", h.TriggerAggregation)
		}
	}
}

func (h *Handler) getUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user ID type in context")
	}
	return id, nil
}

// POST /api/analytics/track
func (h *Handler) Track(c *gin.Context) {
	var req domain.TrackEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	meta := make(map[string]interface{})
	meta["ipAddress"] = c.ClientIP()
	meta["userAgent"] = c.Request.UserAgent()
	if uid, err := h.getUserID(c); err == nil {
		meta["userId"] = uid
	}

	if err := h.service.TrackEvent(c.Request.Context(), req, meta); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GET /api/analytics/overview
func (h *Handler) GetOverview(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	period := c.DefaultQuery("period", "30d")

	stats, err := h.service.GetDashboard(c.Request.Context(), userID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

// GET /api/analytics/insights
func (h *Handler) GetInsights(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	insights, err := h.service.GenerateInsights(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"insights": insights}})
}

// GET /api/analytics/inventory
func (h *Handler) GetInventoryStats(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	data, err := h.service.GetInventoryPerformance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// POST /api/analytics/jobs/aggregate (Admin/Manual Trigger)
func (h *Handler) TriggerAggregation(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	err = h.service.RunDailyAggregation(c.Request.Context(), userID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Aggregation complete for " + date})
}
