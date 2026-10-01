package services

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
)

type skinMinecraftService struct {
	MinecraftService
	err        error
	skin       string
	variant    domain.SkinVariant
	mojangUUID *uuid.UUID
	clears     int
}

func (s *skinMinecraftService) SetPlayerSkinInSeason(ctx context.Context, _ uuid.UUID, _ *domain.Profile, skin string, variant domain.SkinVariant, mojangUUID *uuid.UUID) (*domain.PlayerSkin, error) {
	if ctx.Done() != nil {
		return nil, errors.New("the wait for the server must outlive the request")
	}
	s.skin, s.variant, s.mojangUUID = skin, variant, mojangUUID
	if s.err != nil {
		return nil, s.err
	}
	return &domain.PlayerSkin{TextureURL: "https://textures.minecraft.net/texture/aa", Slim: variant == domain.SkinVariantSlim}, nil
}

func (s *skinMinecraftService) ClearPlayerSkinInSeason(context.Context, uuid.UUID, *domain.Profile) error {
	s.clears++
	return s.err
}

func TestProfileSkinServiceSetOwnedSkinFile(t *testing.T) {
	t.Setenv("API_PUBLIC_URL", "https://api.lania.example/")
	owner := uuid.New()
	profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), OwnerUserID: &owner}
	skin64, skin32 := encodePNG(t, 64, 64), encodePNG(t, 64, 32)

	tests := []struct {
		name       string
		userID     uuid.UUID
		content    []byte
		variant    domain.SkinVariant
		inactive   bool
		noAccess   bool
		shellErr   error
		wantStatus int
		wantErr    error
	}{
		{name: "64x64 file", userID: owner, content: skin64, variant: domain.SkinVariantSlim},
		{name: "legacy 64x32 file", userID: owner, content: skin32, variant: domain.SkinVariantClassic},
		{name: "HD skin", userID: owner, content: encodePNG(t, 128, 128), variant: domain.SkinVariantClassic, wantStatus: http.StatusBadRequest},
		{name: "not a png", userID: owner, content: []byte("GIF89a"), variant: domain.SkinVariantClassic, wantStatus: http.StatusBadRequest},
		{name: "unknown variant", userID: owner, content: skin64, variant: "wide", wantStatus: http.StatusBadRequest},
		{name: "not the owner", userID: uuid.New(), content: skin64, variant: domain.SkinVariantClassic, wantStatus: http.StatusForbidden},
		{name: "ended season", userID: owner, content: skin64, variant: domain.SkinVariantClassic, inactive: true, wantStatus: http.StatusBadRequest},
		{name: "no access to the season", userID: owner, content: skin64, variant: domain.SkinVariantClassic, noAccess: true, wantStatus: http.StatusForbidden},
		{name: "server is slow", userID: owner, content: skin64, variant: domain.SkinVariantClassic, shellErr: ErrSkinPending, wantErr: ErrSkinPending},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := &fakeObjectStorage{bucket: "lania"}
			minecraft := &skinMinecraftService{err: tt.shellErr}
			service := NewProfileSkinService(
				&stubProfileService{profiles: map[uuid.UUID]*domain.Profile{profile.ID: profile}},
				&passwordAccessService{hasAccess: !tt.noAccess},
				&passwordSeasonService{season: &domain.Season{ID: uuid.New(), IsActive: !tt.inactive}},
				minecraft, nil, objects,
			)

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			skin, err := service.SetOwnedSkinFile(ctx, tt.userID, profile.ID, uuid.New(), tt.content, tt.variant)
			if tt.wantStatus != 0 {
				if status := utils.MapCustomErrorToHttpStatus(err); status != tt.wantStatus {
					t.Fatalf("status = %d (%v), want %d", status, err, tt.wantStatus)
				}
				if objects.putCalls != 0 || minecraft.skin != "" {
					t.Error("a refused skin must not be stored or sent to the server")
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && skin == nil {
				t.Fatal("skin = nil")
			}

			fileID := strings.TrimSuffix(strings.TrimPrefix(objects.putKey, "skins/"), ".png")
			if !skinFileIDPattern.MatchString(fileID) {
				t.Fatalf("stored under %q, want skins/<16 hex>.png", objects.putKey)
			}
			if want := "https://api.lania.example/v1/skins/files/" + fileID + ".png"; minecraft.skin != want {
				t.Errorf("server got %q, want %q", minecraft.skin, want)
			}
			if minecraft.variant != tt.variant {
				t.Errorf("variant = %q, want %q", minecraft.variant, tt.variant)
			}
		})
	}
}

func TestProfileSkinServiceSetOwnedSkinNickname(t *testing.T) {
	owner, account := uuid.New(), uuid.New()
	profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), OwnerUserID: &owner}

	tests := []struct {
		name       string
		nickname   string
		wantStatus int
	}{
		{name: "licensed nickname", nickname: "Notch"},
		{name: "free nickname", nickname: "NobodyHasIt", wantStatus: http.StatusBadRequest},
		{name: "nickname that alters the command", nickname: `Notch" 00000000`, wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minecraft := &skinMinecraftService{}
			service := NewProfileSkinService(
				&stubProfileService{profiles: map[uuid.UUID]*domain.Profile{profile.ID: profile}},
				&passwordAccessService{hasAccess: true},
				&passwordSeasonService{season: &domain.Season{ID: uuid.New(), IsActive: true}},
				minecraft,
				&renameMojang{accounts: map[string]uuid.UUID{"notch": account}},
				&fakeObjectStorage{},
			)

			_, err := service.SetOwnedSkinNickname(context.Background(), owner, profile.ID, uuid.New(), tt.nickname)
			if tt.wantStatus != 0 {
				if status := utils.MapCustomErrorToHttpStatus(err); status != tt.wantStatus {
					t.Fatalf("status = %d (%v), want %d", status, err, tt.wantStatus)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if minecraft.skin != tt.nickname || minecraft.mojangUUID == nil || *minecraft.mojangUUID != account || minecraft.variant != "" {
				t.Errorf("server got %q %q %v, want the nickname with its account", minecraft.skin, minecraft.variant, minecraft.mojangUUID)
			}
		})
	}
}

func TestProfileSkinServiceCooldown(t *testing.T) {
	owner := uuid.New()
	profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), OwnerUserID: &owner}
	minecraft := &skinMinecraftService{}
	service := NewProfileSkinService(
		&stubProfileService{profiles: map[uuid.UUID]*domain.Profile{profile.ID: profile}},
		&passwordAccessService{hasAccess: true},
		&passwordSeasonService{season: &domain.Season{ID: uuid.New(), IsActive: true}},
		minecraft, nil, &fakeObjectStorage{},
	).(*profileSkinService)
	now := time.Now()
	service.now = func() time.Time { return now }

	ctx := context.Background()
	if err := service.ClearOwnedSkin(ctx, owner, profile.ID, uuid.New()); err != nil {
		t.Fatal(err)
	}
	err := service.ClearOwnedSkin(ctx, owner, profile.ID, uuid.New())
	if status := utils.MapCustomErrorToHttpStatus(err); status != http.StatusTooManyRequests {
		t.Fatalf("second change: status = %d (%v), want 429", status, err)
	}
	now = now.Add(ownerSkinCooldown)
	if err := service.ClearOwnedSkin(ctx, owner, profile.ID, uuid.New()); err != nil {
		t.Fatalf("after the cooldown: %v", err)
	}
	if minecraft.clears != 2 {
		t.Errorf("clears = %d, want 2", minecraft.clears)
	}
}

func TestProfileSkinServiceGetSkinFile(t *testing.T) {
	service := NewProfileSkinService(nil, nil, nil, nil, nil, &fakeObjectStorage{
		objects: map[string][]byte{"skins/0123456789abcdef.png": []byte("png")},
	})

	if content, err := service.GetSkinFile(context.Background(), "0123456789abcdef"); err != nil || string(content) != "png" {
		t.Errorf("stored file = %q, %v", content, err)
	}
	for _, fileID := range []string{"../glyth_preview/x", "0123456789ABCDEF", "0123"} {
		_, err := service.GetSkinFile(context.Background(), fileID)
		if status := utils.MapCustomErrorToHttpStatus(err); status != http.StatusNotFound {
			t.Errorf("%q: status = %d, want 404", fileID, status)
		}
	}
}
