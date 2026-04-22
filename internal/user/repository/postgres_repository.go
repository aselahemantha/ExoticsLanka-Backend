package repository

import (
	"context"
	"errors"

	"github.com/aselahemantha/exoticsLanka/internal/user/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, name, email, role, phone, avatar_url, bio, location, verified, verified_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`
	_, err := r.db.Exec(ctx, query,
		user.ID, user.Name, user.Email, user.Role, user.Phone, user.AvatarURL, user.Bio, user.Location, user.Verified, user.VerifiedAt, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, name, email, role, phone, avatar_url, bio, location, verified, verified_at, created_at, updated_at
		FROM users WHERE id = $1
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.Phone, &user.AvatarURL, &user.Bio, &user.Location, &user.Verified, &user.VerifiedAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil, nil when not found to easily check existence
		}
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, role, phone, avatar_url, bio, location, verified, verified_at, created_at, updated_at
		FROM users WHERE email = $1
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.Phone, &user.AvatarURL, &user.Bio, &user.Location, &user.Verified, &user.VerifiedAt, &user.CreatedAt, &user.UpdatedAt,
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
			name = $1, phone = $2, avatar_url = $3, bio = $4, location = $5, verified = $6, verified_at = $7, updated_at = NOW()
		WHERE id = $8
	`
	// Note: We don't typically update email/role via profile updates easily, those belong to Auth/Admin domains usually.
	_, err := r.db.Exec(ctx, query,
		user.Name, user.Phone, user.AvatarURL, user.Bio, user.Location, user.Verified, user.VerifiedAt, user.ID,
	)
	return err
}

func (r *postgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// -- Verification Repository --

type postgresVerificationRepository struct {
	db *pgxpool.Pool
}

func NewPostgresVerificationRepository(db *pgxpool.Pool) domain.VerificationRepository {
	return &postgresVerificationRepository{db: db}
}

func (r *postgresVerificationRepository) CreateRequest(ctx context.Context, req *domain.VerificationRequest) error {
	query := `
		INSERT INTO verification_requests (
			id, user_id, request_type, status, requested_at, reviewed_at, reviewer_id, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`
	_, err := r.db.Exec(ctx, query,
		req.ID, req.UserID, req.RequestType, req.Status, req.RequestedAt, req.ReviewedAt, req.ReviewerID, req.Notes,
	)
	return err
}

func (r *postgresVerificationRepository) GetRequestByID(ctx context.Context, id uuid.UUID) (*domain.VerificationRequest, error) {
	query := `
		SELECT id, user_id, request_type, status, requested_at, reviewed_at, reviewer_id, notes
		FROM verification_requests WHERE id = $1
	`
	var req domain.VerificationRequest
	err := r.db.QueryRow(ctx, query, id).Scan(
		&req.ID, &req.UserID, &req.RequestType, &req.Status, &req.RequestedAt, &req.ReviewedAt, &req.ReviewerID, &req.Notes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}
