package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
)

type stubMinecraftService struct {
	MinecraftService
	online     uuid.UUIDs
	staff      uuid.UUIDs
	groups     map[uuid.UUID][]string
	err        error
	lastGroup  []string
	groupCalls int
}

func (s *stubMinecraftService) GetGroupsByMinecraftUUIDs(context.Context, uuid.UUIDs) (map[uuid.UUID][]string, error) {
	s.groupCalls++
	return s.groups, s.err
}

func (s *stubMinecraftService) ListOnlineMinecraftUUIDs(context.Context) (uuid.UUIDs, error) {
	return s.online, s.err
}

func (s *stubMinecraftService) ListMinecraftUUIDsByGroups(_ context.Context, groups []string) (uuid.UUIDs, error) {
	s.lastGroup = groups
	return s.staff, s.err
}

func TestFilterMinecraftUUIDs(t *testing.T) {
	first, second, third := uuid.New(), uuid.New(), uuid.New()
	minecraft := &stubMinecraftService{online: uuid.UUIDs{first, second}, staff: uuid.UUIDs{second, third}}
	service := &profileService{minecraftService: minecraft}
	ctx := context.Background()

	only, err := service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{Search: "a"})
	if err != nil || only != nil {
		t.Errorf("filter without server state = %v %v, want no restriction", only, err)
	}

	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true})
	if err != nil || only == nil || len(*only) != 2 {
		t.Errorf("online = %v %v, want 2 players", only, err)
	}

	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{StaffOnly: true})
	if err != nil || only == nil || len(*only) != 2 {
		t.Errorf("staff = %v %v, want 2 players", only, err)
	}
	if want := []string{"owner", "admin", "mod"}; len(minecraft.lastGroup) != 3 || minecraft.lastGroup[0] != want[0] || minecraft.lastGroup[1] != want[1] || minecraft.lastGroup[2] != want[2] {
		t.Errorf("groups = %v, want %v", minecraft.lastGroup, want)
	}

	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true, StaffOnly: true})
	if err != nil || only == nil || len(*only) != 1 || (*only)[0] != second {
		t.Errorf("online staff = %v %v, want only the second player", only, err)
	}

	minecraft.staff = uuid.UUIDs{third}
	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true, StaffOnly: true})
	if err != nil || only == nil || len(*only) != 0 {
		t.Errorf("no online staff = %v %v, want an empty restriction", only, err)
	}

	minecraft.err = errors.New("shell down")
	if _, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true}); err == nil {
		t.Error("filter must fail when the server is unreachable")
	}
}

type stubCache struct {
	storage.CacheStorage
	values map[string]string
}

func (c *stubCache) GetKey(_ context.Context, key string) (string, error) {
	value, ok := c.values[key]
	if !ok {
		return "", errors.New("cache miss")
	}
	return value, nil
}

func (c *stubCache) SetKey(_ context.Context, key string, value interface{}, _ time.Duration) error {
	c.values[key] = value.(string)
	return nil
}

func TestApplyProfileRoles(t *testing.T) {
	admin, cached, unknown := uuid.New(), uuid.New(), uuid.New()
	newProfiles := func() []*domain.Profile {
		return []*domain.Profile{
			{MinecraftUUID: admin, Role: domain.RolePlayer},
			{MinecraftUUID: cached, Role: domain.RolePlayer},
			{MinecraftUUID: unknown, Role: domain.RoleModerator},
		}
	}
	ctx := context.Background()

	t.Run("live roles win and are read in one call", func(t *testing.T) {
		minecraft := &stubMinecraftService{groups: map[uuid.UUID][]string{admin: {"default", "admin"}, unknown: {"default"}}}
		cache := &stubCache{values: map[string]string{profileRoleCacheKey(cached): "owner"}}
		service := &profileService{minecraftService: minecraft, cache: cache}

		profiles := newProfiles()
		service.ApplyProfileRoles(ctx, profiles)

		if profiles[0].Role != domain.RoleAdmin || profiles[1].Role != domain.RoleOwner || profiles[2].Role != domain.RolePlayer {
			t.Errorf("roles = %v %v %v, want admin owner player", profiles[0].Role, profiles[1].Role, profiles[2].Role)
		}
		if minecraft.groupCalls != 1 {
			t.Errorf("server calls = %d, want 1 for all uncached profiles", minecraft.groupCalls)
		}
		if cache.values[profileRoleCacheKey(admin)] != "admin" {
			t.Errorf("live role must be cached, got %v", cache.values)
		}
	})

	t.Run("cached roles skip the server", func(t *testing.T) {
		minecraft := &stubMinecraftService{}
		cache := &stubCache{values: map[string]string{profileRoleCacheKey(admin): "admin"}}
		service := &profileService{minecraftService: minecraft, cache: cache}

		profiles := []*domain.Profile{{MinecraftUUID: admin, Role: domain.RolePlayer}}
		service.ApplyProfileRoles(ctx, profiles)

		if profiles[0].Role != domain.RoleAdmin || minecraft.groupCalls != 0 {
			t.Errorf("role = %v, server calls = %d, want admin without calls", profiles[0].Role, minecraft.groupCalls)
		}
	})

	t.Run("stored roles are kept when the server is down", func(t *testing.T) {
		minecraft := &stubMinecraftService{err: errors.New("shell down")}
		cache := &stubCache{values: map[string]string{}}
		service := &profileService{minecraftService: minecraft, cache: cache}

		profiles := []*domain.Profile{
			{MinecraftUUID: admin, Role: domain.RoleAdmin},
			{MinecraftUUID: unknown, Role: domain.RoleModerator},
		}
		service.ApplyProfileRoles(ctx, profiles)

		if profiles[0].Role != domain.RoleAdmin || profiles[1].Role != domain.RoleModerator {
			t.Errorf("roles = %v %v, want the stored admin and mod", profiles[0].Role, profiles[1].Role)
		}
		if len(cache.values) != 0 {
			t.Errorf("stored roles must not be cached as live ones, got %v", cache.values)
		}
	})
}
