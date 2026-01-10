package repository

import (
	"context"
	"errors"

	"github.com/exoticsLanka/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresUserRepository struct {
	db *pgxpool.Pool
}

// NewPostgresUserRepository creates a new user repository
func NewPostgresUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, status, role, 
			email_verified, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 
			$6, $7, $8
		)
	`
	_, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Status, user.Role,
		user.EmailVerified, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, status, role, email_verified, created_at, updated_at
		FROM users WHERE id = $1
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Status, &user.Role,
		&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, status, role, email_verified, created_at, updated_at
		FROM users WHERE email = $1
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Status, &user.Role,
		&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET 
			status = $1, role = $2, email_verified = $3, 
			updated_at = $4, last_login_at = $5, failed_login_attempts = $6,
			locked_until = $7
		WHERE id = $8
	`
	_, err := r.db.Exec(ctx, query,
		user.Status, user.Role, user.EmailVerified,
		user.UpdatedAt, user.LastLoginAt, user.FailedLoginAttempts,
		user.LockedUntil, user.ID,
	)
	return err
}

type postgresAuditRepository struct {
	db *pgxpool.Pool
}

// NewPostgresAuditRepository creates a new audit repository
func NewPostgresAuditRepository(db *pgxpool.Pool) domain.AuditRepository {
	return &postgresAuditRepository{db: db}
}

func (r *postgresAuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			user_id, event_type, event_category, description, 
			ip_address, user_agent, success, created_at
		) VALUES (
			$1, $2, $3, $4, 
			$5, $6, $7, $8
		)
	`
	_, err := r.db.Exec(ctx, query,
		log.UserID, log.EventType, log.EventCategory, log.Description,
		log.IPAddress, log.UserAgent, log.Success, log.CreatedAt,
	)
	return err
}
