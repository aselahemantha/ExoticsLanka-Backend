package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aselahemantha/exoticsLanka/internal/common"
	"github.com/aselahemantha/exoticsLanka/internal/reports/domain"
	"github.com/aselahemantha/exoticsLanka/internal/reports/service"
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
	api := r.Group("/api/reports")
	api.Use(auth.Handle())
	{
		// Public (authenticated) submission
		api.POST("", h.SubmitReport)

		// Admin Routes
		api.GET("", h.GetReports)
		api.GET("/stats", h.GetStats)
		api.GET("/:id", h.GetReport)
		api.PUT("/:id", h.ResolveReport)
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

func (h *Handler) isAdmin(c *gin.Context) bool {
	role, exists := c.Get("userRole")
	return exists && role == "admin"
}

// POST /api/reports
func (h *Handler) SubmitReport(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req domain.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.SubmitReport(c.Request.Context(), req, userID)
	if err != nil {
		if err.Error() == "you have already reported this listing recently" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Report submitted successfully",
		"data":    report,
	})
}

// GET /api/reports (Admin)
func (h *Handler) GetReports(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	reason := c.Query("reason")

	reports, pagination, err := h.service.GetReports(c.Request.Context(), status, reason, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"reports":    reports,
			"pagination": pagination,
		},
	})
}

// GET /api/reports/:id (Admin)
func (h *Handler) GetReport(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	report, err := h.service.GetReport(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if report == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": report})
}

// PUT /api/reports/:id (Admin)
func (h *Handler) ResolveReport(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	adminID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var req domain.ResolveReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.ResolveReport(c.Request.Context(), id, adminID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Report updated successfully",
		"data":    report,
	})
}

// GET /api/reports/stats (Admin)
func (h *Handler) GetStats(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}
