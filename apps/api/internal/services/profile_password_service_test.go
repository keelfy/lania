package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type passwordSeasonService struct {
	SeasonService
	season *domain.Season
}

func (s *passwordSeasonService) GetSeasonByID(context.Context, uuid.UUID) (*domain.Season, error) {
	return s.season, nil
}

type passwordAccessService struct {
	AccessService
	hasAccess bool
}

func (s *passwordAccessService) CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return s.hasAccess, nil
}

type passwordMinecraftService struct {
	MinecraftService
	err  error
	hash string
}

func (s *passwordMinecraftService) SetPasswordInSeason(_ context.Context, _ uuid.UUID, _ *domain.Profile, passwordBcrypt string) error {
	s.hash = passwordBcrypt
	return s.err
}

func TestProfilePasswordService(t *testing.T) {
	ctx := context.Background()
	owner := uuid.New()

	tests := []struct {
		name       string
		userID     uuid.UUID
		password   string
		licensed   bool
		inactive   bool
		noAccess   bool
		shellErr   error
		wantStatus int
	}{
		{name: "owner sets a password", userID: owner, password: "newpass1"},
		{name: "password without a digit", userID: owner, password: "newpass", wantStatus: http.StatusBadRequest},
		{name: "password with a space", userID: owner, password: "new pass1", wantStatus: http.StatusBadRequest},
		{name: "not the owner", userID: uuid.New(), password: "newpass1", wantStatus: http.StatusForbidden},
		{name: "licensed profile", userID: owner, password: "newpass1", licensed: true, wantStatus: http.StatusConflict},
		{name: "ended season", userID: owner, password: "newpass1", inactive: true, wantStatus: http.StatusBadRequest},
		{name: "no access to the season", userID: owner, password: "newpass1", noAccess: true, wantStatus: http.StatusForbidden},
		{name: "never registered in game", userID: owner, password: "newpass1", shellErr: ErrPlayerNotRegistered, wantStatus: http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcUUID := uuid.New()
			queries := &fakeVerificationQueries{
				profile: &domain.Profile{ID: uuid.New(), MinecraftUUID: mcUUID, MinecraftUsername: "Steve", OwnerUserID: &owner},
			}
			if tt.licensed {
				queries.mojangUUID = &mcUUID
			}
			minecraft := &passwordMinecraftService{err: tt.shellErr}
			service := NewProfilePasswordService(
				&stubMainStorage{queries: queries},
				&passwordAccessService{hasAccess: !tt.noAccess},
				&passwordSeasonService{season: &domain.Season{IsActive: !tt.inactive}},
				minecraft,
			)

			err := service.SetOwnedPassword(ctx, tt.userID, queries.profile.ID, uuid.New(), tt.password)
			if tt.wantStatus != 0 {
				if httpStatus(err) != tt.wantStatus {
					t.Fatalf("status %d (%v), want %d", httpStatus(err), err, tt.wantStatus)
				}
				return
			}
			if err != nil {
				t.Fatalf("SetOwnedPassword() error = %v", err)
			}
			if bcrypt.CompareHashAndPassword([]byte(minecraft.hash), []byte(tt.password)) != nil {
				t.Errorf("hash %q does not match the password", minecraft.hash)
			}
		})
	}
}
