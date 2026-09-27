package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
)

type stubProfileService struct {
	ProfileService
	profiles map[uuid.UUID]*domain.Profile
	setCalls int
}

func (s *stubProfileService) GetProfileByID(_ context.Context, id uuid.UUID) (*domain.Profile, error) {
	if profile, ok := s.profiles[id]; ok {
		copied := *profile
		return &copied, nil
	}
	return nil, utils.NewNotFoundError("profile not found", nil)
}

func (s *stubProfileService) SetProfileOwner(ctx context.Context, id uuid.UUID, owner *uuid.UUID) (*domain.Profile, error) {
	s.setCalls++
	profile, err := s.GetProfileByID(ctx, id)
	if err != nil {
		return nil, err
	}
	profile.OwnerUserID = owner
	s.profiles[id] = profile
	return profile, nil
}

type stubAdminUserService struct {
	AdminUserService
	users map[string]*domain.User
}

func (s *stubAdminUserService) FindUserByEmail(_ context.Context, email string) (*domain.User, error) {
	if user, ok := s.users[email]; ok {
		return user, nil
	}
	return nil, utils.NewNotFoundError("user not found", nil)
}

func (s *stubAdminUserService) GetUserByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	for _, user := range s.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, utils.NewNotFoundError("user not found", nil)
}

func TestAdminProfileService(t *testing.T) {
	previousOwner := &domain.User{ID: uuid.New(), Email: "old@x.io"}
	newOwner := &domain.User{ID: uuid.New(), Email: "new@x.io"}
	profile := &domain.Profile{ID: uuid.New(), OwnerUserID: &previousOwner.ID}

	profiles := &stubProfileService{profiles: map[uuid.UUID]*domain.Profile{profile.ID: profile}}
	users := &stubAdminUserService{users: map[string]*domain.User{previousOwner.Email: previousOwner, newOwner.Email: newOwner}}
	svc := NewAdminProfileService(profiles, users)
	ctx := context.Background()

	t.Run("details include the owner", func(t *testing.T) {
		_, owner, err := svc.GetProfileDetails(ctx, profile.ID)
		if err != nil || owner == nil || owner.ID != previousOwner.ID {
			t.Fatalf("got owner %+v and error %v", owner, err)
		}
	})

	t.Run("transfer replaces an existing owner", func(t *testing.T) {
		got, owner, err := svc.TransferProfile(ctx, profile.ID, newOwner.Email)
		if err != nil {
			t.Fatal(err)
		}
		if got.OwnerUserID == nil || *got.OwnerUserID != newOwner.ID || owner.ID != newOwner.ID {
			t.Errorf("got owner %v, want %v", got.OwnerUserID, newOwner.ID)
		}
	})

	t.Run("unknown email leaves the owner alone", func(t *testing.T) {
		profiles.setCalls = 0
		_, _, err := svc.TransferProfile(ctx, profile.ID, "nobody@x.io")
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusNotFound || profiles.setCalls != 0 {
			t.Fatalf("got error %v after %d writes, want not found and no writes", err, profiles.setCalls)
		}
	})

	t.Run("unknown profile is reported before the email is looked up", func(t *testing.T) {
		_, _, err := svc.TransferProfile(ctx, uuid.New(), "nobody@x.io")
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusNotFound || err.Error() != "profile not found" {
			t.Fatalf("got error %v, want profile not found", err)
		}
	})

	t.Run("release removes the owner", func(t *testing.T) {
		got, err := svc.ReleaseProfile(ctx, profile.ID)
		if err != nil || got.OwnerUserID != nil {
			t.Fatalf("got owner %v and error %v, want no owner", got.OwnerUserID, err)
		}

		_, owner, err := svc.GetProfileDetails(ctx, profile.ID)
		if err != nil || owner != nil {
			t.Fatalf("got owner %+v and error %v, want no owner", owner, err)
		}
	})
}
