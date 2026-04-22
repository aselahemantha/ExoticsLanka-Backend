package handler

import (
	"net/http"

	"github.com/aselahemantha/exoticsLanka/internal/common"
	"github.com/aselahemantha/exoticsLanka/internal/notification/domain"
	"github.com/aselahemantha/exoticsLanka/internal/notification/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.Engine, auth common.AuthMiddleware) {
	api := r.Group("/api")
	{
		notifications := api.Group("/notifications")
		notifications.Use(auth.Handle())
		{
			notifications.GET("/preferences", h.GetPreferences)
			notifications.PATCH("/preferences", h.UpdatePreferences)
			notifications.POST("/send", h.SendNotification)
		}
	}
}

func (h *Handler) GetPreferences(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	prefs, err := h.service.GetPreferences(c.Request.Context(), userID.(uuid.UUID).String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": prefs})
}

func (h *Handler) UpdatePreferences(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var prefs domain.NotificationPreference
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enforce UserID from token
	prefs.UserID = userID.(uuid.UUID).String()

	err := h.service.UpdatePreferences(c.Request.Context(), &prefs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Preferences updated"})
}

func (h *Handler) SendNotification(c *gin.Context) {
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
