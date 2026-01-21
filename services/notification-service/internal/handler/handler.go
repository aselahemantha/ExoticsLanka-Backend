package handler

import (
	"net/http"

	"github.com/aselahemantha/exoticsLanka/services/notification-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/notification-service/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetPreferences(c *gin.Context) {
	userID := c.GetString("userID")
	prefs, err := h.service.GetPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": prefs})
}

func (h *Handler) UpdatePreferences(c *gin.Context) {
	userID := c.GetString("userID")
	var prefs domain.NotificationPreference
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enforce UserID from token
	prefs.UserID = userID

	err := h.service.UpdatePreferences(c.Request.Context(), &prefs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Preferences updated"})
}

func (h *Handler) SendNotification(c *gin.Context) {
	// This endpoint is internal/system triggered.
	// Ideally it should be protected by a service-key or specific admin role.
	// For now, we assume any authenticated user can trigger for DEMO purposes,
	// OR we can check for a special header/claim.
	// Let's assume it's protected by AuthMiddleware as normal.

	var req domain.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Attempt to send
	err := h.service.SendNotification(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Notification dispatched"})
}
