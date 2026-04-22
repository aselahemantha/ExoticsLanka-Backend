package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aselahemantha/exoticsLanka/internal/common"
	"github.com/aselahemantha/exoticsLanka/internal/messaging/domain"
	"github.com/aselahemantha/exoticsLanka/internal/messaging/service"
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
	api.Use(auth.Handle())
	{
		api.GET("/conversations", h.GetUserConversations)
		api.POST("/conversations", h.CreateConversation)
		api.GET("/conversations/:id", h.GetConversationByID)
		api.POST("/conversations/:id/messages", h.SendMessage)
		api.PUT("/conversations/:id/read", h.MarkRead)
		api.GET("/messages/unread-count", h.GetUnreadCount)
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

// GET /api/conversations
func (h *Handler) GetUserConversations(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	archived := c.Query("archived") == "true"

	conversations, pagination, err := h.service.GetUserConversations(c.Request.Context(), userID, page, limit, archived)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"conversations": conversations,
			"pagination":    pagination,
		},
	})
}

// POST /api/conversations
func (h *Handler) CreateConversation(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req domain.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateConversation(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	msg := "Conversation created"
	if !resp.IsNew {
		msg = "Message sent to existing conversation"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": msg,
		"data":    resp,
	})
}

// GET /api/conversations/:id
func (h *Handler) GetConversationByID(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	conv, messages, err := h.service.GetConversationByID(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": conv.ID,
			"listing": gin.H{
				"id":    conv.ListingID,
				"title": conv.ListingTitle,
				"image": conv.ListingImage,
				"price": conv.ListingPrice,
			},
			"participant":   conv.Participant,
			"messages":      messages,
			"createdAt":     conv.CreatedAt,
			"lastMessageAt": conv.LastMessageAt,
		},
	})
}

// POST /api/conversations/:id/messages
func (h *Handler) SendMessage(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req domain.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.service.SendMessage(c.Request.Context(), id, userID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    msg,
	})
}

// PUT /api/conversations/:id/read
func (h *Handler) MarkRead(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	err = h.service.MarkConversationRead(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Messages marked as read",
	})
}

// GET /api/messages/unread-count
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	stats, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
