package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aselahemantha/exoticsLanka/internal/common"
	"github.com/aselahemantha/exoticsLanka/internal/contact/domain"
	"github.com/aselahemantha/exoticsLanka/internal/contact/service"
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

	// Public Contact Form (Optional Auth)
	api.POST("/contact", auth.OptionalHandle(), h.SubmitInquiry)

	// Admin Routes (Protected)
	admin := api.Group("/contact")
	admin.Use(auth.Handle())
	{
		admin.GET("", h.GetInquiries)
		admin.GET("/stats", h.GetStats)
		admin.GET("/:id", h.GetInquiry)
		admin.PUT("/:id", h.RespondInquiry)
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

// POST /api/contact
func (h *Handler) SubmitInquiry(c *gin.Context) {
	var req domain.CreateInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	meta := make(map[string]interface{})
	meta["ipAddress"] = c.ClientIP()
	meta["userAgent"] = c.Request.UserAgent()
	if userID, err := h.getUserID(c); err == nil {
		meta["userId"] = userID
	}

	inq, err := h.service.SubmitInquiry(c.Request.Context(), req, meta)
	if err != nil {
		if err.Error() == "too many inquiries. please try again later" {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Message sent! We'll get back to you within 24 hours.",
		"data":    inq,
	})
}

// GET /api/contact (Admin)
func (h *Handler) GetInquiries(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	subject := c.Query("subject")
	priority := c.Query("priority")
	search := c.Query("search")

	inquiries, pagination, err := h.service.GetInquiries(c.Request.Context(), status, subject, priority, search, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"inquiries":  inquiries,
			"pagination": pagination,
		},
	})
}

// GET /api/contact/:id (Admin)
func (h *Handler) GetInquiry(c *gin.Context) {
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inquiry ID"})
		return
	}

	inq, err := h.service.GetInquiry(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if inq == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inquiry not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": inq})
}

// PUT /api/contact/:id (Admin)
func (h *Handler) RespondInquiry(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inquiry ID"})
		return
	}

	var req domain.RespondInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inq, err := h.service.RespondInquiry(c.Request.Context(), id, adminID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Inquiry updated successfully",
		"data":    inq,
	})
}

// GET /api/contact/stats (Admin)
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
