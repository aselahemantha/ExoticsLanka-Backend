package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/notification-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetPreferences(ctx context.Context, userID string) (*domain.NotificationPreference, error) {
	var prefs domain.NotificationPreference
	query := `
		SELECT user_id, email_enabled, sms_enabled, push_enabled, marketing_emails, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&prefs.UserID,
		&prefs.EmailEnabled,
		&prefs.SMSEnabled,
		&prefs.PushEnabled,
		&prefs.MarketingEmails,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)
	if err != nil {
		// If not found, return default preferences
		return &domain.NotificationPreference{
			UserID:          userID,
			EmailEnabled:    true,
			SMSEnabled:      false,
			PushEnabled:     true,
			MarketingEmails: false,
		}, nil
	}
	return &prefs, nil
}

func (r *Repository) UpsertPreferences(ctx context.Context, prefs *domain.NotificationPreference) error {
	query := `
		INSERT INTO notification_preferences (user_id, email_enabled, sms_enabled, push_enabled, marketing_emails, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			email_enabled = EXCLUDED.email_enabled,
			sms_enabled = EXCLUDED.sms_enabled,
			push_enabled = EXCLUDED.push_enabled,
			marketing_emails = EXCLUDED.marketing_emails,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(ctx, query, prefs.UserID, prefs.EmailEnabled, prefs.SMSEnabled, prefs.PushEnabled, prefs.MarketingEmails, time.Now())
	if err != nil {
		return fmt.Errorf("failed to upsert preferences: %w", err)
	}
	return nil
}

func (r *Repository) LogNotification(ctx context.Context, log *domain.NotificationLog) error {
	query := `
		INSERT INTO notification_logs (user_id, type, provider, external_id, status, error_message, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query, log.UserID, log.Type, log.Provider, log.ExternalID, log.Status, log.ErrorMessage, log.Metadata).Scan(&log.ID, &log.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to log notification: %w", err)
	}
	return nil
}

// GetUserEmail fetches user email from the shared 'users' table
func (r *Repository) GetUserEmail(ctx context.Context, userID string) (string, error) {
	var email string
	query := `SELECT email FROM users WHERE id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&email)
	if err != nil {
		return "", fmt.Errorf("failed to get user email: %w", err)
	}
	return email, nil
}

// GetUserPhone fetches user phone from the shared 'users' table
func (r *Repository) GetUserPhone(ctx context.Context, userID string) (string, error) {
	var phone *string
	query := `SELECT phone_number FROM users WHERE id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&phone)
	if err != nil {
		return "", fmt.Errorf("failed to get user phone: %w", err)
	}
	if phone == nil {
		return "", nil // Or error if phone is required
	}
	return *phone, nil
}
