package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents the user model in the database
type User struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	Email               string     `json:"email" db:"email"`
	PasswordHash        string     `json:"-" db:"password_hash"`
	Status              string     `json:"status" db:"status"` // pending, active, suspended, deleted
	EmailVerified       bool       `json:"email_verified" db:"email_verified"`
	EmailVerifiedAt     *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	Role                string     `json:"role" db:"role"` // buyer, seller, dealer, admin, super_admin
	TwoFactorEnabled    bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret     *string    `json:"-" db:"two_factor_secret"`
	FailedLoginAttempts int        `json:"-" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty" db:"locked_until"`
	OAuthProvider       *string    `json:"oauth_provider,omitempty" db:"oauth_provider"`
	OAuthID             *string    `json:"oauth_id,omitempty" db:"oauth_id"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Session represents a user session
type Session struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	Token          string    `json:"token" db:"token"`
	RefreshToken   *string   `json:"refresh_token,omitempty" db:"refresh_token"`
	DeviceID       *string   `json:"device_id,omitempty" db:"device_id"`
	DeviceName     *string   `json:"device_name,omitempty" db:"device_name"`
	IPAddress      *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      *string   `json:"user_agent,omitempty" db:"user_agent"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	LastActivityAt time.Time `json:"last_activity_at" db:"last_activity_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// VerificationToken for email and password reset
type VerificationToken struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Token     string     `json:"token" db:"token"`
	TokenHash string     `json:"-" db:"token_hash"`
	Type      string     `json:"type" db:"type"` // email_verification, password_reset
	Used      bool       `json:"used" db:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	IPAddress *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string    `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// AuditLog tracks system events
type AuditLog struct {
	ID            int64                  `json:"id" db:"id"`
	UserID        *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	EventType     string                 `json:"event_type" db:"event_type"` // login, logout, etc.
	EventCategory string                 `json:"event_category" db:"event_category"`
	Description   *string                `json:"description,omitempty" db:"description"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	IPAddress     *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     *string                `json:"user_agent,omitempty" db:"user_agent"`
	Success       bool                   `json:"success" db:"success"`
	ErrorMessage  *string                `json:"error_message,omitempty" db:"error_message"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// UserRepository defines methods for user persistence
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
}

// SessionRepository defines methods for session persistence (Redis/DB)
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByToken(ctx context.Context, token string) (*Session, error)
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

// AuditRepository defines methods for audit logs
type AuditRepository interface {
	Create(ctx context.Context, log *AuditLog) error
}

// AuthUseCase defines the business logic for authentication
type AuthUseCase interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	GetMe(ctx context.Context, userID uuid.UUID) (*UserResponse, error)
	VerifyEmail(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
	ChangePassword(ctx context.Context, req *ChangePasswordRequest) error
}

// DTOs for UseCases
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	Role      string `json:"role"`
	IPAddress string `json:"-"`
	UserAgent string `json:"-"`
}

type RegisterResponse struct {
	ID     uuid.UUID `json:"id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Status string    `json:"status"`
}

type LoginRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	IPAddress string `json:"-"`
	UserAgent string `json:"-"`
}

type LoginResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int           `json:"expires_in"`
}

type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ChangePasswordRequest struct {
	UserID          uuid.UUID `json:"-"`
	CurrentPassword string    `json:"current_password" binding:"required"`
	NewPassword     string    `json:"new_password" binding:"required,min=8"`
}
