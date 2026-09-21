package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/domain"
)

type fakeShell struct {
	clients.ShellAPI
	online    uuid.UUIDs
	err       error
	prefixes  []string
	whitelist int
	roles     map[uuid.UUID]string
	roleGroup []string
}

func (s *fakeShell) ListOnlinePlayers(context.Context) (uuid.UUIDs, error) {
	return s.online, s.err
}

func (s *fakeShell) GetOnlineStatus(_ context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	status := make(map[uuid.UUID]bool, len(mcUUIDs))
	for _, mcUUID := range mcUUIDs {
		for _, online := range s.online {
			if online == mcUUID {
				status[mcUUID] = true
			}
		}
		if _, ok := status[mcUUID]; !ok {
			status[mcUUID] = false
		}
	}
	return status, s.err
}

func (s *fakeShell) SetPlayerPrefix(_ context.Context, _ uuid.UUID, prefix string) error {
	s.prefixes = append(s.prefixes, prefix)
	return s.err
}

func (s *fakeShell) SetPlayerRoles(_ context.Context, roleGroups []string, roles map[uuid.UUID]string) error {
	s.roleGroup = roleGroups
	s.roles = roles
	return s.err
}

func (s *fakeShell) AddToWhitelist(context.Context, uuid.UUID, string) error {
	s.whitelist++
	return s.err
}

type fakeShellPool map[string]*fakeShell

func (p fakeShellPool) Get(address string) (clients.ShellAPI, error) { return p[address], nil }

type fakeSeasons struct {
	SeasonService
	seasons []*domain.Season
}

func (s *fakeSeasons) GetSeasons(context.Context) ([]*domain.Season, error) { return s.seasons, nil }

func (s *fakeSeasons) GetSeasonByID(_ context.Context, id uuid.UUID) (*domain.Season, error) {
	for _, season := range s.seasons {
		if season.ID == id {
			return season, nil
		}
	}
	return nil, errors.New("season not found")
}

func season(address string, active bool) *domain.Season {
	season := &domain.Season{ID: uuid.New(), IsActive: active}
	if address != "" {
		season.ShellAddress = &address
	}
	return season
}

func TestMinecraftService_OnlineReadsOnlyTheServerOfTheSeason(t *testing.T) {
	first, second, other := uuid.New(), uuid.New(), uuid.New()
	pool := fakeShellPool{
		"a:1": {online: uuid.UUIDs{first}},
		"b:1": {online: uuid.UUIDs{second}},
	}
	seasonA, seasonB := season("a:1", true), season("b:1", true)
	service := NewMinecraftService(&fakeSeasons{seasons: []*domain.Season{seasonA, seasonB}}, pool)
	ctx := context.Background()

	online, err := service.ListOnlineInSeason(ctx, seasonA.ID)
	if err != nil || len(online) != 1 || online[0] != first {
		t.Fatalf("online in season A = %v %v, want only the player of its server", online, err)
	}

	status, err := service.GetOnlineStatusInSeason(ctx, seasonB.ID, uuid.UUIDs{first, second, other})
	if err != nil || status[first] || !status[second] || status[other] {
		t.Fatalf("status in season B = %v %v, want only the player of its server online", status, err)
	}
}

func TestMinecraftService_OnlineIsUnavailableWithoutRunningServer(t *testing.T) {
	pool := fakeShellPool{"a:1": {online: uuid.UUIDs{uuid.New()}}, "c:1": {}}
	ended, noShell := season("c:1", false), season("", true)
	service := NewMinecraftService(&fakeSeasons{seasons: []*domain.Season{season("a:1", true), ended, noShell}}, pool)
	ctx := context.Background()

	for name, seasonID := range map[string]uuid.UUID{"ended": ended.ID, "without shell": noShell.ID} {
		if _, err := service.ListOnlineInSeason(ctx, seasonID); !errors.Is(err, ErrOnlineUnavailable) {
			t.Errorf("list in a season %s: error = %v, want ErrOnlineUnavailable", name, err)
		}
		if _, err := service.GetOnlineStatusInSeason(ctx, seasonID, uuid.UUIDs{uuid.New()}); !errors.Is(err, ErrOnlineUnavailable) {
			t.Errorf("status in a season %s: error = %v, want ErrOnlineUnavailable", name, err)
		}
	}
}

func TestMinecraftService_OnlineFailsWhenServerIsDown(t *testing.T) {
	pool := fakeShellPool{"a:1": {err: errors.New("down")}}
	active := season("a:1", true)
	service := NewMinecraftService(&fakeSeasons{seasons: []*domain.Season{active}}, pool)

	_, err := service.ListOnlineInSeason(context.Background(), active.ID)
	if err == nil || errors.Is(err, ErrOnlineUnavailable) {
		t.Fatalf("error = %v, want a failure that is not ErrOnlineUnavailable", err)
	}
}

func TestMinecraftService_WhitelistUsesShellOfSeason(t *testing.T) {
	pool := fakeShellPool{"a:1": {}, "b:1": {}}
	past, upcoming := season("a:1", false), season("", true)
	other := season("b:1", true)
	service := NewMinecraftService(&fakeSeasons{seasons: []*domain.Season{past, upcoming, other}}, pool)
	ctx := context.Background()
	profile := &domain.Profile{MinecraftUUID: uuid.New(), MinecraftUsername: "steve"}

	for _, id := range []uuid.UUID{past.ID, upcoming.ID} {
		if err := service.AddToWhitelist(ctx, id, profile); err != nil {
			t.Fatal(err)
		}
	}
	if pool["a:1"].whitelist != 1 || pool["b:1"].whitelist != 0 {
		t.Fatalf("whitelist calls a=%d b=%d, want 1 0; a season without shell is skipped", pool["a:1"].whitelist, pool["b:1"].whitelist)
	}
}

func TestMinecraftService_RolesGoToEveryActiveServer(t *testing.T) {
	admin, player := uuid.New(), uuid.New()
	pool := fakeShellPool{"a:1": {}, "b:1": {}, "c:1": {}}
	seasons := &fakeSeasons{seasons: []*domain.Season{season("a:1", true), season("b:1", true), season("c:1", false)}}
	service := NewMinecraftService(seasons, pool)
	ctx := context.Background()

	roles := map[uuid.UUID]domain.Role{admin: domain.RoleAdmin, player: domain.RolePlayer}
	if err := service.SetPlayerRoles(ctx, roles); err != nil {
		t.Fatal(err)
	}
	for _, address := range []string{"a:1", "b:1"} {
		shell := pool[address]
		if shell.roles[admin] != "admin" || shell.roles[player] != "" || len(shell.roles) != 2 {
			t.Errorf("%s roles = %v, want admin and a player without group", address, shell.roles)
		}
		if len(shell.roleGroup) != 3 || shell.roleGroup[0] != "owner" || shell.roleGroup[1] != "admin" || shell.roleGroup[2] != "mod" {
			t.Errorf("%s role groups = %v, want owner admin mod", address, shell.roleGroup)
		}
	}
	if pool["c:1"].roles != nil {
		t.Error("an inactive season must not get roles")
	}

	pool["a:1"].err = errors.New("down")
	pool["b:1"].roles = nil
	if err := service.SetPlayerRoles(ctx, roles); err == nil {
		t.Fatal("want an error when a server rejects the roles")
	}
	if pool["b:1"].roles == nil {
		t.Fatal("the other server must still get the roles")
	}
}

func TestMinecraftService_RoleAndPrefixInSeasonUseShellOfSeason(t *testing.T) {
	admin := uuid.New()
	pool := fakeShellPool{"a:1": {}, "b:1": {}}
	first, second, upcoming := season("a:1", true), season("b:1", true), season("", true)
	service := NewMinecraftService(&fakeSeasons{seasons: []*domain.Season{first, second, upcoming}}, pool)
	ctx := context.Background()

	roles := map[uuid.UUID]domain.Role{admin: domain.RoleAdmin}
	if err := service.SetPlayerRolesInSeason(ctx, first.ID, roles); err != nil {
		t.Fatal(err)
	}
	if err := service.SetPrefixInSeason(ctx, first.ID, admin, "[X]"); err != nil {
		t.Fatal(err)
	}
	if pool["a:1"].roles[admin] != "admin" || len(pool["a:1"].prefixes) != 1 {
		t.Errorf("a:1 roles = %v, prefixes = %v, want the role and the prefix", pool["a:1"].roles, pool["a:1"].prefixes)
	}
	if pool["b:1"].roles != nil || len(pool["b:1"].prefixes) != 0 {
		t.Error("the shell of another season must not be written")
	}

	if err := service.SetPlayerRolesInSeason(ctx, upcoming.ID, roles); err != nil {
		t.Fatalf("a season without shell error = %v, want nothing to do", err)
	}
	if err := service.SetPrefixInSeason(ctx, upcoming.ID, admin, "[X]"); err != nil {
		t.Fatalf("a season without shell error = %v, want nothing to do", err)
	}

	pool["b:1"].err = errors.New("down")
	if err := service.SetPlayerRolesInSeason(ctx, second.ID, roles); err == nil {
		t.Fatal("want an error when the server rejects the roles")
	}
	if err := service.SetPrefixInSeason(ctx, second.ID, admin, "[X]"); err == nil {
		t.Fatal("want an error when the server rejects the prefix")
	}
}
