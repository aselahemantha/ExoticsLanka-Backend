package domain

import (
	"time"
)

type NotificationPreference struct {
	UserID          string    `json:"user_id"`
	EmailEnabled    bool      `json:"email_enabled"`
	SMSEnabled      bool      `json:"sms_enabled"`
	PushEnabled     bool      `json:"push_enabled"`
	MarketingEmails bool      `json:"marketing_emails"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type NotificationLog struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Type         string                 `json:"type"`
	Provider     string                 `json:"provider"`
	ExternalID   string                 `json:"external_id"`
	Status       string                 `json:"status"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

type NotificationRequest struct {
	UserID  string                 `json:"user_id"`
	Type    string                 `json:"type"`    // "welcome", "listing_approved", etc.
	Channel string                 `json:"channel"` // "email", "sms", "push" (optional, if empty send to all enabled)
	Data    map[string]interface{} `json:"data"`
}
