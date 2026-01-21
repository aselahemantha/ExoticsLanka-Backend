package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/aselahemantha/exoticsLanka/services/notification-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/notification-service/internal/provider"
	"github.com/aselahemantha/exoticsLanka/services/notification-service/internal/repository"
)

type Service struct {
	repo          *repository.Repository
	emailProvider provider.EmailProvider
	smsProvider   provider.SMSProvider
}

func NewService(repo *repository.Repository, email provider.EmailProvider, sms provider.SMSProvider) *Service {
	return &Service{
		repo:          repo,
		emailProvider: email,
		smsProvider:   sms,
	}
}

func (s *Service) GetPreferences(ctx context.Context, userID string) (*domain.NotificationPreference, error) {
	return s.repo.GetPreferences(ctx, userID)
}

func (s *Service) UpdatePreferences(ctx context.Context, prefs *domain.NotificationPreference) error {
	return s.repo.UpsertPreferences(ctx, prefs)
}

func (s *Service) SendNotification(ctx context.Context, req *domain.NotificationRequest) error {
	// 1. Get User Preferences
	prefs, err := s.repo.GetPreferences(ctx, req.UserID)
	if err != nil {
		return err
	}

	// 2. Determine Channels to send to
	sendEmail := prefs.EmailEnabled && (req.Channel == "" || req.Channel == "email")
	sendSMS := prefs.SMSEnabled && (req.Channel == "" || req.Channel == "sms")

	// Override if explicit channel requested but disabled in prefs?
	// Usually system notifications (password reset) ignore prefs, but marketing ones don't.
	// For MVP, lets assume "marketing" flag controls marketing, others control 'transactional'
	// Ref: Documentation says "Manage notification preferences".

	if req.Type == "marketing" && !prefs.MarketingEmails {
		sendEmail = false
	}

	var errs []string

	// 3. Send Email
	if sendEmail {
		email, err := s.repo.GetUserEmail(ctx, req.UserID)
		if err == nil && email != "" {
			subject, body := s.renderEmailTemplate(req.Type, req.Data)
			msgID, err := s.emailProvider.SendEmail(email, subject, body)
			status := "sent"
			errMsg := ""
			if err != nil {
				status = "failed"
				errMsg = err.Error()
				errs = append(errs, fmt.Sprintf("email failed: %v", err))
			}

			// Log
			_ = s.repo.LogNotification(ctx, &domain.NotificationLog{
				UserID:       req.UserID,
				Type:         req.Type,
				Provider:     "sendgrid", // or dynamic
				ExternalID:   msgID,
				Status:       status,
				ErrorMessage: errMsg,
				Metadata:     req.Data,
			})
		}
	}

	// 4. Send SMS
	if sendSMS {
		phone, err := s.repo.GetUserPhone(ctx, req.UserID)
		if err == nil && phone != "" {
			body := s.renderSMSTemplate(req.Type, req.Data)
			sid, err := s.smsProvider.SendSMS(phone, body)
			status := "sent"
			errMsg := ""
			if err != nil {
				status = "failed"
				errMsg = err.Error()
				errs = append(errs, fmt.Sprintf("sms failed: %v", err))
			}

			// Log
			_ = s.repo.LogNotification(ctx, &domain.NotificationLog{
				UserID:       req.UserID,
				Type:         req.Type,
				Provider:     "twilio",
				ExternalID:   sid,
				Status:       status,
				ErrorMessage: errMsg,
				Metadata:     req.Data,
			})
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// Simple Template Engine
func (s *Service) renderEmailTemplate(templateType string, data map[string]interface{}) (string, string) {
	subject := "Notification from Exotics Lanka"
	body := "You have a new notification."

	switch templateType {
	case "welcome":
		subject = "Welcome to Exotics Lanka!"
		body = fmt.Sprintf("<h1>Welcome %v!</h1><p>We are glad to have you.</p>", data["name"])
	case "listing_approved":
		subject = "Your listing has been approved"
		body = fmt.Sprintf("<p>Your listing <b>%v</b> is now live.</p>", data["listing_title"])
	case "listing_rejected":
		subject = "Action required on your listing"
		body = fmt.Sprintf("<p>Your listing <b>%v</b> was rejected. Reason: %v</p>", data["listing_title"], data["reason"])
	case "new_message":
		subject = "New Message Received"
		body = fmt.Sprintf("<p>You have a new message from %v about %v.</p>", data["sender_name"], data["listing_title"])
	case "new_lead":
		subject = "New Lead!"
		body = fmt.Sprintf("<p>New inquiry for %v from %v.</p>", data["listing_title"], data["buyer_name"])
	}

	return subject, body
}

func (s *Service) renderSMSTemplate(templateType string, data map[string]interface{}) string {
	body := "Exotics Lanka Notification"

	switch templateType {
	case "welcome":
		body = fmt.Sprintf("Welcome to Exotics Lanka, %v!", data["name"])
	case "listing_approved":
		body = fmt.Sprintf("Your listing %v is now live.", data["listing_title"])
	case "new_message":
		body = fmt.Sprintf("New message from %v about %v.", data["sender_name"], data["listing_title"])
	case "new_lead":
		body = fmt.Sprintf("New lead for %v from %v.", data["listing_title"], data["buyer_name"])
	}

	return body
}
