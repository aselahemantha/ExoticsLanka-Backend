package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aselahemantha/exoticsLanka/internal/common"
	"github.com/aselahemantha/exoticsLanka/internal/saved_searches/domain"
	"github.com/aselahemantha/exoticsLanka/internal/saved_searches/service"
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
	api := r.Group("/api/searches")
	api.Use(auth.Handle())
	{
		api.GET("", h.GetUserSearches)
		api.POST("", h.CreateSavedSearch)
		api.GET("/new-matches", h.GetNewMatchesOverview)
		api.GET("/:id", h.GetDetailedSearch)
		api.PUT("/:id", h.UpdateSavedSearch)
		api.DELETE("/:id", h.DeleteSavedSearch)
		api.POST("/:id/check", h.CheckMatches)
		api.POST("/:id/run", h.RunSearch)
		api.PUT("/:id/alerts", h.UpdateAlerts)
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

// GET /api/searches
func (h *Handler) GetUserSearches(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	searches, _, err := h.service.GetUserSavedSearches(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	total, err := h.service.GetTotalNewMatches(c.Request.Context(), userID)
	totalCount := 0
	if err == nil && total != nil {
		totalCount = total.TotalNewMatches
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"searches":        searches,
			"totalNewMatches": totalCount,
		},
	})
}

// POST /api/searches
func (h *Handler) CreateSavedSearch(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req domain.CreateSavedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ss, err := h.service.CreateSavedSearch(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Search saved successfully",
		"data":    ss,
	})
}

// GET /api/searches/:id
func (h *Handler) GetDetailedSearch(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search ID"})
		return
	}

	ss, err := h.service.GetSavedSearch(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": ss})
}

// PUT /api/searches/:id
func (h *Handler) UpdateSavedSearch(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search ID"})
		return
	}

	var req domain.UpdateSavedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ss, err := h.service.UpdateSavedSearch(c.Request.Context(), id, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Search updated",
		"data":    ss,
	})
}

// DELETE /api/searches/:id
func (h *Handler) DeleteSavedSearch(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search ID"})
		return
	}

	err = h.service.DeleteSavedSearch(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Saved search deleted"})
}

// POST /api/searches/:id/check
func (h *Handler) CheckMatches(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search ID"})
		return
	}

	res, err := h.service.CheckForNewMatches(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

// POST /api/searches/:id/run
func (h *Handler) RunSearch(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.service.RunSearch(c.Request.Context(), id, userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

// PUT /api/searches/:id/alerts
func (h *Handler) UpdateAlerts(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search ID"})
		return
	}

	var req domain.UpdateAlertsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateAlertSettings(c.Request.Context(), id, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Alert settings updated"})
}

// GET /api/searches/new-matches
func (h *Handler) GetNewMatchesOverview(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	res, err := h.service.GetTotalNewMatches(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}
