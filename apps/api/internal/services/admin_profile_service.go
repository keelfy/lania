package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

type AdminProfileService interface {
	// GetProfileDetails returns the profile and its owner, nil while nobody owns the profile.
	GetProfileDetails(ctx context.Context, profileID uuid.UUID) (*domain.Profile, *domain.User, error)
	// TransferProfile gives the profile to the user with the email. Limits on profiles per user do not apply.
	TransferProfile(ctx context.Context, profileID uuid.UUID, email string) (*domain.Profile, *domain.User, error)
	// ReleaseProfile removes the owner of the profile, so anybody can claim it.
	ReleaseProfile(ctx context.Context, profileID uuid.UUID) (*domain.Profile, error)
}

type adminProfileService struct {
	profileService   ProfileService
	adminUserService AdminUserService
}

func NewAdminProfileService(
	profileService ProfileService,
	adminUserService AdminUserService,
) AdminProfileService {
	return &adminProfileService{
		profileService:   profileService,
		adminUserService: adminUserService,
	}
}

func (s *adminProfileService) GetProfileDetails(ctx context.Context, profileID uuid.UUID) (*domain.Profile, *domain.User, error) {
	profile, err := s.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, nil, err
	}
	if profile.OwnerUserID == nil {
		return profile, nil, nil
	}

	owner, err := s.adminUserService.GetUserByID(ctx, *profile.OwnerUserID)
	if err != nil {
		return nil, nil, err
	}
	return profile, owner, nil
}

func (s *adminProfileService) TransferProfile(ctx context.Context, profileID uuid.UUID, email string) (*domain.Profile, *domain.User, error) {
	// Look the profile up first, so an unknown profile is reported before an unknown email.
	if _, err := s.profileService.GetProfileByID(ctx, profileID); err != nil {
		return nil, nil, err
	}

	owner, err := s.adminUserService.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}

	profile, err := s.profileService.SetProfileOwner(ctx, profileID, &owner.ID)
	if err != nil {
		return nil, nil, err
	}
	return profile, owner, nil
}

func (s *adminProfileService) ReleaseProfile(ctx context.Context, profileID uuid.UUID) (*domain.Profile, error) {
	return s.profileService.SetProfileOwner(ctx, profileID, nil)
}
