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
	// roleWrites and prefixWrites count the writes per season.
	roleWrites   map[uuid.UUID]int
	prefixWrites map[uuid.UUID]int
	roleErr      map[uuid.UUID]error
	prefixErr    map[uuid.UUID]error
	seasonErr    map[uuid.UUID]error
}

func (m *resyncMinecraft) SetPlayerRolesInSeason(_ context.Context, seasonID uuid.UUID, roles map[uuid.UUID]domain.Role) error {
	m.roles = roles
	m.roleWrites[seasonID]++
	return m.roleErr[seasonID]
}

func (m *resyncMinecraft) SetPrefixInSeason(_ context.Context, seasonID, _ uuid.UUID, prefix string) error {
	m.prefix = prefix
	m.prefixWrites[seasonID]++
	return m.prefixErr[seasonID]
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

func newResyncMinecraft() *resyncMinecraft {
	return &resyncMinecraft{
		whitelist:    make(map[uuid.UUID]string),
		roleWrites:   make(map[uuid.UUID]int),
		prefixWrites: make(map[uuid.UUID]int),
	}
}

// partErrors returns the error of every part of the season, by part.
func partErrors(t *testing.T, report *domain.ProfileResync, seasonID uuid.UUID) map[domain.ResyncPart]error {
	t.Helper()
	for _, season := range report.Seasons {
		if season.SeasonID != seasonID {
			continue
		}
		errs := make(map[domain.ResyncPart]error, len(season.Parts))
		for _, part := range season.Parts {
			errs[part.Part] = part.Err
		}
		return errs
	}
	t.Fatalf("report has no season %s", seasonID)
	return nil
}

func TestProfileResyncService(t *testing.T) {
	ctx := context.Background()
	owner := uuid.New()
	profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), OwnerUserID: &owner, Role: domain.RoleModerator}
	profiles := &stubProfileService{profiles: map[uuid.UUID]*domain.Profile{profile.ID: profile}}

	withAccess, withoutAccess, inactive, noShell := season("a:1", true), season("b:1", true), season("c:1", false), season("", true)
	withAccess.Name, withoutAccess.Name = "Season A", "Season B"
	seasons := &fakeSeasons{seasons: []*domain.Season{withAccess, withoutAccess, inactive, noShell}}
	access := &resyncAccess{accesses: map[uuid.UUID][]*domain.ProfileAccess{
		profile.MinecraftUUID: {{MinecraftUUID: profile.MinecraftUUID, SeasonID: withAccess.ID}},
	}}

	newService := func(minecraft *resyncMinecraft, cosmetics *resyncCosmetics) *profileResyncService {
		return NewProfileResyncService(profiles, cosmetics, access, seasons, minecraft).(*profileResyncService)
	}

	t.Run("writes role, prefix and whitelist to every season and reports all of them", func(t *testing.T) {
		minecraft := newResyncMinecraft()
		report, err := newService(minecraft, &resyncCosmetics{}).ResyncProfile(ctx, profile.ID)
		if err != nil {
			t.Fatal(err)
		}

		if !report.OK() || len(report.Seasons) != 2 {
			t.Fatalf("report = %+v, want ok for the two seasons with a server", report)
		}
		if report.Seasons[0].SeasonName != "Season A" || report.Seasons[1].SeasonName != "Season B" {
			t.Errorf("seasons = %q and %q, want them named in order", report.Seasons[0].SeasonName, report.Seasons[1].SeasonName)
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
		for _, untouched := range []*domain.Season{inactive, noShell} {
			if _, touched := minecraft.whitelist[untouched.ID]; touched || minecraft.roleWrites[untouched.ID] != 0 {
				t.Errorf("season %s must not be touched", untouched.ID)
			}
		}
	})

	t.Run("a failing part is reported for its season only and does not stop the others", func(t *testing.T) {
		minecraft := newResyncMinecraft()
		minecraft.roleErr = map[uuid.UUID]error{withAccess.ID: errors.New("shell down")}
		minecraft.seasonErr = map[uuid.UUID]error{withAccess.ID: utils.NewInternalServerError("failed to add profile to whitelist", errors.New("rcon down"))}
		report, err := newService(minecraft, &resyncCosmetics{}).ResyncProfile(ctx, profile.ID)
		if err != nil {
			t.Fatal(err)
		}

		if report.OK() {
			t.Fatal("report is ok, want a failure")
		}
		failed, fine := partErrors(t, report, withAccess.ID), partErrors(t, report, withoutAccess.ID)
		if failed[domain.ResyncPartRole] == nil || failed[domain.ResyncPartAccess] == nil || failed[domain.ResyncPartCosmetics] != nil {
			t.Errorf("season with failures = %v, want role and access failed, cosmetics fine", failed)
		}
		if fine[domain.ResyncPartRole] != nil || fine[domain.ResyncPartCosmetics] != nil || fine[domain.ResyncPartAccess] != nil {
			t.Errorf("other season = %v, want every part fine", fine)
		}
		if report.Seasons[0].OK() || !report.Seasons[1].OK() {
			t.Error("want the first season failed and the second one fine")
		}
	})

	t.Run("a prefix that cannot be built fails the cosmetics of every season and keeps the other parts", func(t *testing.T) {
		minecraft := newResyncMinecraft()
		report, err := newService(minecraft, &resyncCosmetics{err: errors.New("no colors")}).ResyncProfile(ctx, profile.ID)
		if err != nil {
			t.Fatal(err)
		}

		for _, season := range []*domain.Season{withAccess, withoutAccess} {
			errs := partErrors(t, report, season.ID)
			if errs[domain.ResyncPartCosmetics] == nil || errs[domain.ResyncPartRole] != nil || errs[domain.ResyncPartAccess] != nil {
				t.Errorf("season %s = %v, want only cosmetics failed", season.Name, errs)
			}
			if minecraft.prefixWrites[season.ID] != 0 {
				t.Errorf("season %s got a prefix that could not be built", season.Name)
			}
		}
	})

	t.Run("seasons that share a shell get one role and prefix write", func(t *testing.T) {
		first, second := season("shared:1", true), season("shared:1", true)
		shared := &fakeSeasons{seasons: []*domain.Season{first, second}}
		minecraft := newResyncMinecraft()
		minecraft.roleErr = map[uuid.UUID]error{first.ID: errors.New("shell down")}
		report, err := NewProfileResyncService(profiles, &resyncCosmetics{}, access, shared, minecraft).ResyncProfile(ctx, profile.ID)
		if err != nil {
			t.Fatal(err)
		}

		if minecraft.roleWrites[first.ID]+minecraft.roleWrites[second.ID] != 1 || minecraft.prefixWrites[first.ID]+minecraft.prefixWrites[second.ID] != 1 {
			t.Errorf("role writes = %v, prefix writes = %v, want one of each", minecraft.roleWrites, minecraft.prefixWrites)
		}
		if partErrors(t, report, first.ID)[domain.ResyncPartRole] == nil || partErrors(t, report, second.ID)[domain.ResyncPartRole] == nil {
			t.Error("both seasons must show the role failure of their shared shell")
		}
	})

	t.Run("an unknown profile is not found", func(t *testing.T) {
		_, err := newService(newResyncMinecraft(), &resyncCosmetics{}).ResyncProfile(ctx, uuid.New())
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
	minecraft := newResyncMinecraft()
	active := season("a:1", true)
	service := NewProfileResyncService(profiles, &resyncCosmetics{}, &resyncAccess{}, &fakeSeasons{seasons: []*domain.Season{active}}, minecraft).(*profileResyncService)
	now := time.Now()
	service.now = func() time.Time { return now }

	if _, err := service.ResyncOwnedProfile(ctx, profile.ID, uuid.New()); utils.MapCustomErrorToHttpStatus(err) != http.StatusForbidden {
		t.Fatalf("error = %v, want forbidden for somebody else", err)
	}

	if _, err := service.ResyncOwnedProfile(ctx, profile.ID, owner); err != nil {
		t.Fatalf("first resync error = %v", err)
	}
	if _, err := service.ResyncOwnedProfile(ctx, profile.ID, owner); utils.MapCustomErrorToHttpStatus(err) != http.StatusTooManyRequests {
		t.Fatalf("second resync error = %v, want too many requests", err)
	}

	now = now.Add(ownerResyncCooldown)
	if _, err := service.ResyncOwnedProfile(ctx, profile.ID, owner); err != nil {
		t.Fatalf("resync after the cooldown error = %v", err)
	}

	minecraft.roles = nil
	if _, err := service.ResyncProfile(ctx, profile.ID); err != nil || minecraft.roles == nil {
		t.Fatalf("admin resync error = %v, want it to skip the cooldown", err)
	}
}
