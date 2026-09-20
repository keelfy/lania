package services

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
)

type resyncMinecraft struct {
	MinecraftService
	roles     map[uuid.UUID]domain.Role
	prefix    string
	whitelist map[uuid.UUID]string
	roleErr   error
	seasonErr map[uuid.UUID]error
}

func (m *resyncMinecraft) SetPlayerRoles(_ context.Context, roles map[uuid.UUID]domain.Role) error {
	m.roles = roles
	return m.roleErr
}

func (m *resyncMinecraft) SetPrefixByMinecraftUUID(_ context.Context, _ uuid.UUID, prefix string) error {
	m.prefix = prefix
	return nil
}

func (m *resyncMinecraft) AddToWhitelist(_ context.Context, seasonID uuid.UUID, _ *domain.Profile) error {
	m.whitelist[seasonID] = "add"
	return m.seasonErr[seasonID]
}

func (m *resyncMinecraft) RemoveFromWhitelist(_ context.Context, seasonID uuid.UUID, _ *domain.Profile) error {
	m.whitelist[seasonID] = "remove"
	return m.seasonErr[seasonID]
}

type resyncCosmetics struct {
	ProfileCosmeticsService
	err error
}

func (c *resyncCosmetics) GetProfileChatPrefix(context.Context, *domain.Profile) (string, error) {
	return "<reset>[X]", c.err
}

type resyncAccess struct {
	AccessService
	accesses map[uuid.UUID][]*domain.ProfileAccess
}

func (a *resyncAccess) GetAccessesByMinecraftUUIDs(context.Context, uuid.UUIDs) (map[uuid.UUID][]*domain.ProfileAccess, error) {
	return a.accesses, nil
}

func TestProfileResyncService(t *testing.T) {
	ctx := context.Background()
	owner := uuid.New()
	profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), OwnerUserID: &owner, Role: domain.RoleModerator}
	profiles := &stubProfileService{profiles: map[uuid.UUID]*domain.Profile{profile.ID: profile}}

	withAccess, withoutAccess, inactive, noShell := season("a:1", true), season("b:1", true), season("c:1", false), season("", true)
	seasons := &fakeSeasons{seasons: []*domain.Season{withAccess, withoutAccess, inactive, noShell}}
	access := &resyncAccess{accesses: map[uuid.UUID][]*domain.ProfileAccess{
		profile.MinecraftUUID: {{MinecraftUUID: profile.MinecraftUUID, SeasonID: withAccess.ID}},
	}}

	newService := func(minecraft *resyncMinecraft, cosmetics *resyncCosmetics) *profileResyncService {
		minecraft.whitelist = make(map[uuid.UUID]string)
		return NewProfileResyncService(profiles, cosmetics, access, seasons, minecraft).(*profileResyncService)
	}

	t.Run("writes role, prefix and whitelists", func(t *testing.T) {
		minecraft := &resyncMinecraft{}
		if err := newService(minecraft, &resyncCosmetics{}).ResyncProfile(ctx, profile.ID); err != nil {
			t.Fatal(err)
		}

		if minecraft.roles[profile.MinecraftUUID] != domain.RoleModerator || len(minecraft.roles) != 1 {
			t.Errorf("roles = %v, want the stored mod role", minecraft.roles)
		}
		if minecraft.prefix != "<reset>[X]" {
			t.Errorf("prefix = %q, want the chat prefix", minecraft.prefix)
		}
		if minecraft.whitelist[withAccess.ID] != "add" || minecraft.whitelist[withoutAccess.ID] != "remove" {
			t.Errorf("whitelist = %v, want add for the season with access and remove for the other", minecraft.whitelist)
		}
		if _, touched := minecraft.whitelist[inactive.ID]; touched {
			t.Error("an inactive season must not be touched")
		}
		if _, touched := minecraft.whitelist[noShell.ID]; touched {
			t.Error("a season without shell must not be touched")
		}
	})

	t.Run("a failing part does not stop the others and is named", func(t *testing.T) {
		minecraft := &resyncMinecraft{roleErr: errors.New("shell down"), seasonErr: map[uuid.UUID]error{withoutAccess.ID: errors.New("shell down")}}
		err := newService(minecraft, &resyncCosmetics{}).ResyncProfile(ctx, profile.ID)

		if utils.MapCustomErrorToHttpStatus(err) != http.StatusInternalServerError || err == nil || err.Error() != "failed to resync role, access" {
			t.Fatalf("error = %v, want failed to resync role, access", err)
		}
		if minecraft.prefix == "" || minecraft.whitelist[withAccess.ID] != "add" {
			t.Errorf("prefix = %q, whitelist = %v, want the other parts written", minecraft.prefix, minecraft.whitelist)
		}
	})

	t.Run("an unknown profile is not found", func(t *testing.T) {
		err := newService(&resyncMinecraft{}, &resyncCosmetics{}).ResyncProfile(ctx, uuid.New())
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusNotFound {
			t.Fatalf("error = %v, want not found", err)
		}
	})
}

func TestProfileResyncService_OwnerCooldown(t *testing.T) {
	ctx := context.Background()
	owner := uuid.New()
	profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), OwnerUserID: &owner}
	profiles := &stubProfileService{profiles: map[uuid.UUID]*domain.Profile{profile.ID: profile}}
	minecraft := &resyncMinecraft{whitelist: make(map[uuid.UUID]string)}
	service := NewProfileResyncService(profiles, &resyncCosmetics{}, &resyncAccess{}, &fakeSeasons{}, minecraft).(*profileResyncService)
	now := time.Now()
	service.now = func() time.Time { return now }

	if err := service.ResyncOwnedProfile(ctx, profile.ID, uuid.New()); utils.MapCustomErrorToHttpStatus(err) != http.StatusForbidden {
		t.Fatalf("error = %v, want forbidden for somebody else", err)
	}

	if err := service.ResyncOwnedProfile(ctx, profile.ID, owner); err != nil {
		t.Fatalf("first resync error = %v", err)
	}
	if err := service.ResyncOwnedProfile(ctx, profile.ID, owner); utils.MapCustomErrorToHttpStatus(err) != http.StatusTooManyRequests {
		t.Fatalf("second resync error = %v, want too many requests", err)
	}

	now = now.Add(ownerResyncCooldown)
	if err := service.ResyncOwnedProfile(ctx, profile.ID, owner); err != nil {
		t.Fatalf("resync after the cooldown error = %v", err)
	}

	minecraft.roles = nil
	if err := service.ResyncProfile(ctx, profile.ID); err != nil || minecraft.roles == nil {
		t.Fatalf("admin resync error = %v, want it to skip the cooldown", err)
	}
}
