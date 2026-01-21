package handler

import (
	"net/http"
	"strconv"

	"github.com/aselahemantha/exoticsLanka/services/contact-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/contact-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

// POST /api/contact
func (h *Handler) SubmitInquiry(c *gin.Context) {
	var req domain.CreateInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Gather metadata
	meta := make(map[string]interface{})
	meta["ipAddress"] = c.ClientIP()
	meta["userAgent"] = c.Request.UserAgent()
	if userID, err := GetUserID(c); err == nil {
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
	adminID, err := GetUserID(c)
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
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}
