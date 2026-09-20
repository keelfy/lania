package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

func TestRoleSyncService_SyncRoles(t *testing.T) {
	ctx := context.Background()
	admin, player := uuid.New(), uuid.New()

	t.Run("pushes changed roles to the servers", func(t *testing.T) {
		queries := &stubRoleQueries{changed: []*domain.ProfileRole{
			{MinecraftUUID: admin, Role: domain.RoleAdmin},
			{MinecraftUUID: player, Role: domain.RolePlayer},
		}}
		minecraft := &stubMinecraftService{}
		service := &roleSyncService{storage: &stubMainStorage{queries: queries}, minecraftService: minecraft}

		if err := service.syncRoles(ctx); err != nil {
			t.Fatal(err)
		}
		if minecraft.calls != 1 || minecraft.roles[admin] != domain.RoleAdmin || minecraft.roles[player] != domain.RolePlayer {
			t.Errorf("calls = %d, roles = %v, want both roles in one call", minecraft.calls, minecraft.roles)
		}
	})

	t.Run("no change leaves the servers alone", func(t *testing.T) {
		minecraft := &stubMinecraftService{}
		service := &roleSyncService{storage: &stubMainStorage{queries: &stubRoleQueries{}}, minecraftService: minecraft}

		if err := service.syncRoles(ctx); err != nil || minecraft.calls != 0 {
			t.Errorf("error = %v, calls = %d, want no call", err, minecraft.calls)
		}
	})

	t.Run("a failing server is reported, so the next run tries again", func(t *testing.T) {
		queries := &stubRoleQueries{changed: []*domain.ProfileRole{{MinecraftUUID: admin, Role: domain.RoleAdmin}}}
		minecraft := &stubMinecraftService{err: errors.New("shell down")}
		service := &roleSyncService{storage: &stubMainStorage{queries: queries}, minecraftService: minecraft}

		if err := service.syncRoles(ctx); err == nil {
			t.Error("want the shell error")
		}
	})

	t.Run("a database error is reported and nothing is pushed", func(t *testing.T) {
		queries := &stubRoleQueries{changeErr: errors.New("db down")}
		minecraft := &stubMinecraftService{}
		service := &roleSyncService{storage: &stubMainStorage{queries: queries}, minecraftService: minecraft}

		if err := service.syncRoles(ctx); err == nil || minecraft.calls != 0 {
			t.Errorf("error = %v, calls = %d, want an error and no call", err, minecraft.calls)
		}
	})

	t.Run("the window comes from the config and never drops under two runs", func(t *testing.T) {
		queries := &stubRoleQueries{}
		service := &roleSyncService{storage: &stubMainStorage{queries: queries}, minecraftService: &stubMinecraftService{}}

		t.Setenv("ROLE_SYNC_WINDOW_MINUTES", "30")
		_ = service.syncRoles(ctx)
		if queries.lastSince.Minutes() != 30 {
			t.Errorf("window = %s, want 30m", queries.lastSince)
		}

		t.Setenv("ROLE_SYNC_WINDOW_MINUTES", "1")
		_ = service.syncRoles(ctx)
		if queries.lastSince != 2*roleSyncInterval {
			t.Errorf("window = %s, want 2 runs", queries.lastSince)
		}

		t.Setenv("ROLE_SYNC_WINDOW_MINUTES", "")
		_ = service.syncRoles(ctx)
		if queries.lastSince.Minutes() != 10 {
			t.Errorf("window = %s, want the 10m default", queries.lastSince)
		}
	})
}
