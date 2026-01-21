package handler

import (
	"net/http"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/analytics-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/analytics-service/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
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
	if uid, err := GetUserID(c); err == nil {
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
	userID, err := GetUserID(c)
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
	userID, err := GetUserID(c)
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
	userID, err := GetUserID(c)
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
	userID, err := GetUserID(c) // Technically should check if admin or if user wants to agg their own data
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Default to today or provided date
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
