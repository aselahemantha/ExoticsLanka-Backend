package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/exoticsLanka/auth-service/internal/config"
	"github.com/exoticsLanka/auth-service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authUseCase struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	auditRepo   domain.AuditRepository
	cfg         *config.Config
}

// NewAuthUseCase creates a new auth use case
func NewAuthUseCase(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	auditRepo domain.AuditRepository,
	cfg *config.Config,
) domain.AuthUseCase {
	return &authUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		cfg:         cfg,
	}
}

func (u *authUseCase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.RegisterResponse, error) {
	// 1. Check if user exists
	existingUser, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// 2. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. Create user
	userID := uuid.New()
	now := time.Now()
	user := &domain.User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if user.Role == "" {
		user.Role = "buyer"
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 4. Log audit
	_ = u.auditRepo.Create(ctx, &domain.AuditLog{
		UserID:        &userID,
		EventType:     "account_created",
		EventCategory: "account_management",
		Description:   nil, // Optional: add description
		IPAddress:     &req.IPAddress,
		UserAgent:     &req.UserAgent,
		Success:       true,
		CreatedAt:     now,
	})

	return &domain.RegisterResponse{
		ID:     user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Status: user.Status,
	}, nil
}

func (u *authUseCase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	// 1. Find user
	user, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// 2. Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Log failed attempt
		now := time.Now()
		_ = u.auditRepo.Create(ctx, &domain.AuditLog{
			UserID:        &user.ID,
			EventType:     "login_failed",
			EventCategory: "authentication",
			IPAddress:     &req.IPAddress,
			UserAgent:     &req.UserAgent,
			Success:       false,
			ErrorMessage:  &errStrInvalidCredentials,
			CreatedAt:     now,
		})
		return nil, errors.New("invalid credentials")
	}

	// 3. Generate Tokens
	accessToken, err := u.generateToken(user, 15*time.Minute)
	if err != nil {
		return nil, err
	}
	refreshToken, err := u.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	// 4. Create Session
	session := &domain.Session{
		ID:             uuid.New(),
		UserID:         user.ID,
		Token:          accessToken,
		RefreshToken:   &refreshToken,
		IPAddress:      &req.IPAddress,
		UserAgent:      &req.UserAgent,
		IsActive:       true,
		ExpiresAt:      time.Now().Add(15 * time.Minute),
		LastActivityAt: time.Now(),
		CreatedAt:      time.Now(),
	}

	if err := u.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// 5. Update last login
	now := time.Now()
	user.LastLoginAt = &now
	_ = u.userRepo.Update(ctx, user)

	// 6. Log success
	_ = u.auditRepo.Create(ctx, &domain.AuditLog{
		UserID:        &user.ID,
		EventType:     "login_success",
		EventCategory: "authentication",
		IPAddress:     &req.IPAddress,
		UserAgent:     &req.UserAgent,
		Success:       true,
		CreatedAt:     now,
	})

	return &domain.LoginResponse{
		User: &domain.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	}, nil
}

func (u *authUseCase) generateToken(user *domain.User, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(duration).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.cfg.JWTSecret))
}

func (u *authUseCase) generateRefreshToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"typ": "refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.cfg.JWTRefreshSecret))
}

var errStrInvalidCredentials = "invalid credentials"

func (u *authUseCase) Logout(ctx context.Context, token string) error {
	return u.sessionRepo.Delete(ctx, token)
}

func (u *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domain.LoginResponse, error) {
	// 1. Validate Refresh Token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(u.cfg.JWTRefreshSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("invalid token user id")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// 2. Get User
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 3. Generate New Access Token
	accessToken, err := u.generateToken(user, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	// 4. Update Session (Optional: Rotate Refresh Token)
	// For simplicity, we keep the same refresh token until it expires
	// But we need to ensure the user has an active session
	// Depending on implementation, we might want to check the session repository here.

	return &domain.LoginResponse{
		User: &domain.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Return same refresh token
		ExpiresIn:    900,
	}, nil
}

func (u *authUseCase) GetMe(ctx context.Context, userID uuid.UUID) (*domain.UserResponse, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return &domain.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (u *authUseCase) VerifyEmail(ctx context.Context, token string) error {
	// TODO: Implement actual token verification logic (check DB/Redis for verification token)
	// For now, we mock success
	return nil
}

func (u *authUseCase) ForgotPassword(ctx context.Context, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		// Don't reveal user existence
		return nil
	}
	// TODO: Generate reset token and send email
	return nil
}

func (u *authUseCase) ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error {
	// TODO: Verify token and update password
	return nil
}

func (u *authUseCase) ChangePassword(ctx context.Context, req *domain.ChangePasswordRequest) error {
	user, err := u.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("invalid current password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()
	// user.FailedLoginAttempts = 0 // Reset on password change?

	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Revoke all existing sessions
	return u.sessionRepo.DeleteByUserID(ctx, user.ID)
}
