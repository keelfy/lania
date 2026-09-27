package services

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
)

type resyncMinecraft struct {
	MinecraftService
	roles  map[uuid.UUID]domain.Role
	prefix string
	// prefixes holds the last prefix written per season.
	prefixes  map[uuid.UUID]string
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
	m.prefixes[seasonID] = prefix
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
	// prefixes overrides the prefix of a season.
	prefixes map[uuid.UUID]string
}

func (c *resyncCosmetics) GetProfileChatPrefix(_ context.Context, _, seasonID uuid.UUID) (string, error) {
	if prefix, ok := c.prefixes[seasonID]; ok {
		return prefix, c.err
	}
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
		prefixes:     make(map[uuid.UUID]string),
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

	t.Run("every season gets the prefix of its own selection", func(t *testing.T) {
		minecraft := newResyncMinecraft()
		cosmetics := &resyncCosmetics{prefixes: map[uuid.UUID]string{withAccess.ID: "<reset>[A]", withoutAccess.ID: "<reset>[B]"}}
		if _, err := newService(minecraft, cosmetics).ResyncProfile(ctx, profile.ID); err != nil {
			t.Fatal(err)
		}

		if minecraft.prefixes[withAccess.ID] != "<reset>[A]" || minecraft.prefixes[withoutAccess.ID] != "<reset>[B]" {
			t.Errorf("prefixes = %v, want one prefix per season", minecraft.prefixes)
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

// moveMinecraft records which UUIDs a move touched on the season servers.
type moveMinecraft struct {
	*resyncMinecraft
	removed    uuid.UUIDs
	reset      uuid.UUIDs
	cleared    uuid.UUIDs
	registered uuid.UUIDs
}

func (m *moveMinecraft) RemoveFromWhitelist(ctx context.Context, seasonID uuid.UUID, profile *domain.Profile) error {
	m.removed = append(m.removed, profile.MinecraftUUID)
	return m.resyncMinecraft.RemoveFromWhitelist(ctx, seasonID, profile)
}

func (m *moveMinecraft) SetPlayerRolesInSeason(ctx context.Context, seasonID uuid.UUID, roles map[uuid.UUID]domain.Role) error {
	for mcUUID, role := range roles {
		if role == domain.RolePlayer {
			m.reset = append(m.reset, mcUUID)
		}
	}
	return m.resyncMinecraft.SetPlayerRolesInSeason(ctx, seasonID, roles)
}

func (m *moveMinecraft) SetPrefixInSeason(ctx context.Context, seasonID, mcUUID uuid.UUID, prefix string) error {
	if prefix == "" {
		m.cleared = append(m.cleared, mcUUID)
	}
	return m.resyncMinecraft.SetPrefixInSeason(ctx, seasonID, mcUUID, prefix)
}

func (m *moveMinecraft) RegisterPlayerInSeason(_ context.Context, _ uuid.UUID, profile *domain.Profile) error {
	m.registered = append(m.registered, profile.MinecraftUUID)
	return nil
}

func TestMoveProfileOnServers(t *testing.T) {
	ctx := context.Background()
	profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "newnick", Role: domain.RoleModerator}
	profiles := &stubProfileService{profiles: map[uuid.UUID]*domain.Profile{profile.ID: profile}}
	live, inactive := season("a:1", true), season("b:1", false)
	access := &resyncAccess{accesses: map[uuid.UUID][]*domain.ProfileAccess{
		profile.MinecraftUUID: {{MinecraftUUID: profile.MinecraftUUID, SeasonID: live.ID}},
	}}
	newService := func(minecraft *moveMinecraft) ProfileResyncService {
		return NewProfileResyncService(profiles, &resyncCosmetics{}, access, &fakeSeasons{seasons: []*domain.Season{live, inactive}}, minecraft)
	}

	t.Run("clears the old uuid, registers the nickname and resyncs the profile", func(t *testing.T) {
		old := &domain.Profile{ID: profile.ID, MinecraftUUID: uuid.New(), MinecraftUsername: "oldnick"}
		minecraft := &moveMinecraft{resyncMinecraft: newResyncMinecraft()}

		report := newService(minecraft).MoveProfileOnServers(ctx, profile.ID, old)

		if report == nil || !report.OK() || len(report.Seasons) != 1 {
			t.Fatalf("report = %+v, want a resync of the live season", report)
		}
		want := uuid.UUIDs{old.MinecraftUUID}
		if !slices.Equal(minecraft.removed, want) || !slices.Equal(minecraft.reset, want) || !slices.Equal(minecraft.cleared, want) {
			t.Errorf("removed %v, reset %v, cleared %v; want only the old uuid each time", minecraft.removed, minecraft.reset, minecraft.cleared)
		}
		if !slices.Equal(minecraft.registered, uuid.UUIDs{profile.MinecraftUUID}) {
			t.Errorf("registered = %v, want the current uuid", minecraft.registered)
		}
		if minecraft.whitelist[live.ID] != "add" || minecraft.roles[profile.MinecraftUUID] != domain.RoleModerator {
			t.Errorf("whitelist %v, roles %v; want the current uuid whitelisted with its role", minecraft.whitelist, minecraft.roles)
		}
	})

	t.Run("a licensed rename keeps the uuid and only registers the new name", func(t *testing.T) {
		old := &domain.Profile{ID: profile.ID, MinecraftUUID: profile.MinecraftUUID, MinecraftUsername: "oldnick"}
		minecraft := &moveMinecraft{resyncMinecraft: newResyncMinecraft()}

		newService(minecraft).MoveProfileOnServers(ctx, profile.ID, old)

		if len(minecraft.removed) != 0 || len(minecraft.reset) != 0 || len(minecraft.cleared) != 0 {
			t.Errorf("removed %v, reset %v, cleared %v; want nothing cleared", minecraft.removed, minecraft.reset, minecraft.cleared)
		}
		if !slices.Equal(minecraft.registered, uuid.UUIDs{profile.MinecraftUUID}) {
			t.Errorf("registered = %v, want the uuid under the new name", minecraft.registered)
		}
	})
}
