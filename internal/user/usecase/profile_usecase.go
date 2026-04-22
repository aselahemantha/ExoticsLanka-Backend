package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/aselahemantha/exoticsLanka/internal/user/domain"
	"github.com/google/uuid"
)

type profileUseCase struct {
	userRepo         domain.UserRepository
	verificationRepo domain.VerificationRepository
}

func NewProfileUseCase(userRepo domain.UserRepository, verificationRepo domain.VerificationRepository) domain.ProfileUseCase {
	return &profileUseCase{
		userRepo:         userRepo,
		verificationRepo: verificationRepo,
	}
}

func (u *profileUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.ProfileResponse, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user profile not found")
	}

	return buildProfileResponse(user), nil
}

func (u *profileUseCase) UpdateProfile(ctx context.Context, userID uuid.UUID, req *domain.UpdateProfileRequest) (*domain.ProfileResponse, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user profile not found")
	}

	if req.Name != nil {
		user.Name = req.Name
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}
	if req.Location != nil {
		user.Location = req.Location
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return buildProfileResponse(user), nil
}

func (u *profileUseCase) RequestVerification(ctx context.Context, userID uuid.UUID, req *domain.VerificationRequestPayload) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Double check role
	if user.Role != "seller" && user.Role != "dealer" {
		return errors.New("only sellers and dealers can request verification")
	}

	// Create verification request record
	vReq := &domain.VerificationRequest{
		ID:          uuid.New(),
		UserID:      user.ID,
		RequestType: req.Type,
		Status:      "pending",
		RequestedAt: time.Now(),
	}

	return u.verificationRepo.CreateRequest(ctx, vReq)
}

func (u *profileUseCase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return u.userRepo.Delete(ctx, userID)
}

func buildProfileResponse(u *domain.User) *domain.ProfileResponse {
	name := ""
	if u.Name != nil {
		name = *u.Name
	}

	phone := ""
	if u.Phone != nil {
		phone = *u.Phone
	}

	avatar := ""
	if u.AvatarURL != nil {
		avatar = *u.AvatarURL
	}

	bio := ""
	if u.Bio != nil {
		bio = *u.Bio
	}

	location := ""
	if u.Location != nil {
		location = *u.Location
	}

	return &domain.ProfileResponse{
		ID:        u.ID,
		Name:      name,
		Email:     u.Email,
		Role:      u.Role,
		Phone:     phone,
		AvatarURL: avatar,
		Bio:       bio,
		Location:  location,
		Verified:  u.Verified,
		CreatedAt: u.CreatedAt,
	}
}
