package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aselahemantha/exoticsLanka/internal/common"
	"github.com/aselahemantha/exoticsLanka/internal/reviews/domain"
	"github.com/aselahemantha/exoticsLanka/internal/reviews/service"
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
	api := r.Group("/api")
	{
		// Basic endpoints (some with optional auth)
		api.GET("/reviews/seller/:sellerId", auth.OptionalHandle(), h.GetReviewsBySeller)
		api.GET("/reviews/seller/:sellerId/stats", h.GetSellerStats)

		// Protected endpoints
		protected := api.Group("/reviews")
		protected.Use(auth.Handle())
		{
			protected.POST("", h.CreateReview)
			protected.PUT("/:id", h.UpdateReview)
			protected.DELETE("/:id", h.DeleteReview)
			protected.POST("/:id/helpful", h.ToggleHelpful)
			protected.POST("/:id/response", h.AddSellerResponse)
			protected.POST("/:id/photos", h.AddPhoto)
			protected.DELETE("/:id/photos/:photoId", h.RemovePhoto)
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

func (h *Handler) getOptionalUserID(c *gin.Context) *uuid.UUID {
	val, exists := c.Get("userID")
	if !exists {
		return nil
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}

// GET /api/reviews/seller/:sellerId
func (h *Handler) GetReviewsBySeller(c *gin.Context) {
	sellerIDStr := c.Param("sellerId")
	sellerID, err := uuid.Parse(sellerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid seller ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userID := h.getOptionalUserID(c) // For checking 'hasVotedHelpful'

	reviews, pagination, err := h.service.GetReviewsBySeller(c.Request.Context(), sellerID, page, limit, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"reviews":    reviews,
			"pagination": pagination,
		},
	})
}

// GET /api/reviews/seller/:sellerId/stats
func (h *Handler) GetSellerStats(c *gin.Context) {
	sellerIDStr := c.Param("sellerId")
	sellerID, err := uuid.Parse(sellerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid seller ID"})
		return
	}

	stats, err := h.service.GetSellerStats(c.Request.Context(), sellerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

// POST /api/reviews
func (h *Handler) CreateReview(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req domain.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	review, err := h.service.CreateReview(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": review})
}

// PUT /api/reviews/:id
func (h *Handler) UpdateReview(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var req domain.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	review, err := h.service.UpdateReview(c.Request.Context(), reviewID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": review})
}

// DELETE /api/reviews/:id
func (h *Handler) DeleteReview(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	userRole, _ := c.Get("role")
	isAdmin := userRole == "admin"

	err = h.service.DeleteReview(c.Request.Context(), reviewID, userID, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Review deleted"})
}

// POST /api/reviews/:id/helpful
func (h *Handler) ToggleHelpful(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	count, hasVoted, err := h.service.ToggleHelpful(c.Request.Context(), reviewID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	msg := "Marked as helpful"
	if !hasVoted {
		msg = "Helpful vote removed"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": msg,
		"data": gin.H{
			"helpfulCount": count,
			"hasVoted":     hasVoted,
		},
	})
}

// POST /api/reviews/:id/response
func (h *Handler) AddSellerResponse(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var req domain.SellerResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	review, err := h.service.AddSellerResponse(c.Request.Context(), reviewID, userID, req.Comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Response added",
		"data":    review,
	})
}

// POST /api/reviews/:id/photos
func (h *Handler) AddPhoto(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	type PhotoRequest struct {
		URL string `json:"url" binding:"required"`
	}
	var req PhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.AddPhoto(c.Request.Context(), reviewID, userID, req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Photo added"})
}

// DELETE /api/reviews/:id/photos/:photoId
func (h *Handler) RemovePhoto(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	photoIDStr := c.Param("photoId")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo ID"})
		return
	}

	err = h.service.RemovePhoto(c.Request.Context(), reviewID, userID, photoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Photo removed"})
}
