package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents the user profile entity
type User struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	Name       *string    `json:"name" db:"name"`
	Email      string     `json:"email" db:"email"`
	Role       string     `json:"role" db:"role"`
	Phone      *string    `json:"phone" db:"phone"`
	AvatarURL  *string    `json:"avatarUrl" db:"avatar_url"`
	Bio        *string    `json:"bio" db:"bio"`
	Location   *string    `json:"location" db:"location"`
	Verified   bool       `json:"verified" db:"verified"`
	VerifiedAt *time.Time `json:"verifiedAt,omitempty" db:"verified_at"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time  `json:"updatedAt" db:"updated_at"`
}

// VerificationRequest represents a seller/dealer validation request
type VerificationRequest struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"userId" db:"user_id"`
	RequestType string     `json:"requestType" db:"request_type"`
	Status      string     `json:"status" db:"status"` // pending, approved, rejected
	RequestedAt time.Time  `json:"requestedAt" db:"requested_at"`
	ReviewedAt  *time.Time `json:"reviewedAt,omitempty" db:"reviewed_at"`
	ReviewerID  *uuid.UUID `json:"reviewerId,omitempty" db:"reviewer_id"`
	Notes       *string    `json:"notes,omitempty" db:"notes"`
}

// --- Repositories ---

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type VerificationRepository interface {
	CreateRequest(ctx context.Context, req *VerificationRequest) error
	GetRequestByID(ctx context.Context, id uuid.UUID) (*VerificationRequest, error)
}

// --- UseCases ---

type ProfileUseCase interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*ProfileResponse, error)
	RequestVerification(ctx context.Context, userID uuid.UUID, req *VerificationRequestPayload) error
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
}

// --- DTOs ---

// ProfileResponse mirrors the User representation tailored for responses
type ProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Phone     string    `json:"phone,omitempty"`
	AvatarURL string    `json:"avatarUrl,omitempty"`
	Bio       string    `json:"bio,omitempty"`
	Location  string    `json:"location,omitempty"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"createdAt"`
}

// UpdateProfileRequest tracks the possible mutable profile values
type UpdateProfileRequest struct {
	Name      *string `json:"name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
	Bio       *string `json:"bio,omitempty"`
	Location  *string `json:"location,omitempty"`
}

// VerificationRequestPayload holds submission parameters for starting verification
type VerificationRequestPayload struct {
	Type string `json:"type" binding:"required,oneof=seller dealer"`
}
