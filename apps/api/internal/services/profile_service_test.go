package services

import (
	"context"
	stdsql "database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type stubMinecraftService struct {
	MinecraftService
	online uuid.UUIDs
	roles  map[uuid.UUID]domain.Role
	err    error
	calls  int
	// onlineSeason is the season the last online list was asked for.
	onlineSeason uuid.UUID
}

func (s *stubMinecraftService) ListOnlineInSeason(_ context.Context, seasonID uuid.UUID) (uuid.UUIDs, error) {
	s.onlineSeason = seasonID
	return s.online, s.err
}

func (s *stubMinecraftService) SetPlayerRoles(_ context.Context, roles map[uuid.UUID]domain.Role) error {
	s.calls++
	s.roles = roles
	return s.err
}

// stubRoleQueries keeps profile roles in memory.
type stubRoleQueries struct {
	sql.Queries
	profiles  map[uuid.UUID]*domain.Profile
	staff     uuid.UUIDs
	staffErr  error
	lastRoles []domain.Role
	changed   []*domain.ProfileRole
	changeErr error
	lastSince time.Duration
	writes    int
}

func (q *stubRoleQueries) FindProfileByID(_ context.Context, id uuid.UUID) (*domain.Profile, error) {
	if profile, ok := q.profiles[id]; ok {
		copied := *profile
		return &copied, nil
	}
	return nil, stdsql.ErrNoRows
}

func (q *stubRoleQueries) SetProfileRole(_ context.Context, id uuid.UUID, role domain.Role, _ *uuid.UUID) error {
	q.writes++
	q.profiles[id].Role = role
	return nil
}

func (q *stubRoleQueries) FindMinecraftUUIDsByRoles(_ context.Context, roles []domain.Role) (uuid.UUIDs, error) {
	q.lastRoles = roles
	return q.staff, q.staffErr
}

func (q *stubRoleQueries) FindProfileRolesChangedSince(_ context.Context, since time.Duration) ([]*domain.ProfileRole, error) {
	q.lastSince = since
	return q.changed, q.changeErr
}

func TestFilterMinecraftUUIDs(t *testing.T) {
	first, second, third := uuid.New(), uuid.New(), uuid.New()
	seasonID := uuid.New()
	minecraft := &stubMinecraftService{online: uuid.UUIDs{first, second}}
	queries := &stubRoleQueries{staff: uuid.UUIDs{second, third}}
	service := &profileService{storage: &stubMainStorage{queries: queries}, minecraftService: minecraft}
	ctx := context.Background()

	only, err := service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{Search: "a"}, seasonID)
	if err != nil || only != nil {
		t.Errorf("filter without server state = %v %v, want no restriction", only, err)
	}

	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true}, seasonID)
	if err != nil || only == nil || len(*only) != 2 {
		t.Errorf("online = %v %v, want 2 players", only, err)
	}

	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{StaffOnly: true}, seasonID)
	if err != nil || only == nil || len(*only) != 2 {
		t.Errorf("staff = %v %v, want 2 players", only, err)
	}
	if want := []domain.Role{domain.RoleOwner, domain.RoleAdmin, domain.RoleModerator}; len(queries.lastRoles) != 3 || queries.lastRoles[0] != want[0] || queries.lastRoles[1] != want[1] || queries.lastRoles[2] != want[2] {
		t.Errorf("roles = %v, want %v", queries.lastRoles, want)
	}

	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true, StaffOnly: true}, seasonID)
	if err != nil || only == nil || len(*only) != 1 || (*only)[0] != second {
		t.Errorf("online staff = %v %v, want only the second player", only, err)
	}

	queries.staff = uuid.UUIDs{third}
	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true, StaffOnly: true}, seasonID)
	if err != nil || only == nil || len(*only) != 0 {
		t.Errorf("no online staff = %v %v, want an empty restriction", only, err)
	}

	if minecraft.onlineSeason != seasonID {
		t.Errorf("online players were asked for season %s, want %s", minecraft.onlineSeason, seasonID)
	}

	minecraft.err = errors.New("shell down")
	if _, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true}, seasonID); err == nil {
		t.Error("online filter must fail when the server is unreachable")
	}

	minecraft.err = ErrOnlineUnavailable
	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true}, seasonID)
	if err != nil || only == nil || len(*only) != 0 {
		t.Errorf("online in a season without server = %v %v, want an empty restriction", only, err)
	}
	minecraft.err = errors.New("shell down")
	if only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{StaffOnly: true}, seasonID); err != nil || only == nil {
		t.Errorf("staff = %v %v, want the staff filter to work without the server", only, err)
	}
}

func TestSetProfileRole(t *testing.T) {
	ctx := context.Background()
	profile := &domain.Profile{ID: uuid.New(), Role: domain.RolePlayer}
	queries := &stubRoleQueries{profiles: map[uuid.UUID]*domain.Profile{profile.ID: profile}}
	service := &profileService{storage: &stubMainStorage{queries: queries}}

	t.Run("stores a valid role", func(t *testing.T) {
		got, err := service.SetProfileRole(ctx, profile.ID, domain.RoleModerator)
		if err != nil || got.Role != domain.RoleModerator || queries.profiles[profile.ID].Role != domain.RoleModerator {
			t.Fatalf("got %+v and error %v, want the mod role stored", got, err)
		}
	})

	t.Run("the same role is written again to push it again", func(t *testing.T) {
		queries.writes = 0
		if _, err := service.SetProfileRole(ctx, profile.ID, domain.RoleModerator); err != nil || queries.writes != 1 {
			t.Fatalf("error %v after %d writes, want 1 write", err, queries.writes)
		}
	})

	t.Run("an unknown role is refused", func(t *testing.T) {
		queries.writes = 0
		_, err := service.SetProfileRole(ctx, profile.ID, "vip")
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusBadRequest || queries.writes != 0 {
			t.Fatalf("error %v after %d writes, want bad request and no writes", err, queries.writes)
		}
	})

	t.Run("an unknown profile is not found", func(t *testing.T) {
		_, err := service.SetProfileRole(ctx, uuid.New(), domain.RoleAdmin)
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusNotFound {
			t.Fatalf("error %v, want not found", err)
		}
	})
}

type stubStatsQueries struct {
	sql.Queries
	total, recent, online int64
	lastOnly              *uuid.UUIDs
}

func (q *stubStatsQueries) CountPublicProfiles(_ context.Context, _ string, only *uuid.UUIDs) (int64, error) {
	q.lastOnly = only
	if only != nil {
		return q.online, nil
	}
	return q.total, nil
}

func (q *stubStatsQueries) CountRecentProfiles(context.Context, int) (int64, error) {
	return q.recent, nil
}

type stubMainStorage struct {
	storage.MainStorage
	queries sql.Queries
}

func (s *stubMainStorage) Queries() sql.Queries { return s.queries }

func TestGetProfilesStats(t *testing.T) {
	ctx := context.Background()
	first, second := uuid.New(), uuid.New()

	t.Run("counts online players that have a profile", func(t *testing.T) {
		queries := &stubStatsQueries{total: 120, recent: 7, online: 1}
		minecraft := &stubMinecraftService{online: uuid.UUIDs{first, second}}
		service := &profileService{storage: &stubMainStorage{queries: queries}, minecraftService: minecraft}

		stats, err := service.GetProfilesStats(ctx, uuid.New())
		if err != nil {
			t.Fatal(err)
		}
		if stats.Total != 120 || stats.NewLastWeek != 7 || stats.Online == nil || *stats.Online != 1 {
			t.Errorf("stats = %+v, want 120 total, 7 new and 1 online", stats)
		}
		if queries.lastOnly == nil || len(*queries.lastOnly) != 2 {
			t.Errorf("online count must be limited to the online players, got %v", queries.lastOnly)
		}
	})

	t.Run("online is left out when the server is down", func(t *testing.T) {
		queries := &stubStatsQueries{total: 120, recent: 7}
		minecraft := &stubMinecraftService{err: errors.New("shell down")}
		service := &profileService{storage: &stubMainStorage{queries: queries}, minecraftService: minecraft}

		stats, err := service.GetProfilesStats(ctx, uuid.New())
		if err != nil {
			t.Fatalf("stats must not fail without the server: %v", err)
		}
		if stats.Total != 120 || stats.NewLastWeek != 7 || stats.Online != nil {
			t.Errorf("stats = %+v, want totals without online", stats)
		}
	})

	t.Run("online is left out in a season without a running server", func(t *testing.T) {
		queries := &stubStatsQueries{total: 120, recent: 7, online: 1}
		minecraft := &stubMinecraftService{err: ErrOnlineUnavailable}
		service := &profileService{storage: &stubMainStorage{queries: queries}, minecraftService: minecraft}

		stats, err := service.GetProfilesStats(ctx, uuid.New())
		if err != nil || stats.Online != nil {
			t.Errorf("stats = %+v %v, want totals without online", stats, err)
		}
	})
}

type stubSeasonStatsQueries struct {
	sql.Queries
	stats    []*domain.ProfileSeasonStats
	err      error
	lastUUID uuid.UUID
}

func (q *stubSeasonStatsQueries) FindProfileSeasonStats(_ context.Context, mcUUID uuid.UUID) ([]*domain.ProfileSeasonStats, error) {
	q.lastUUID = mcUUID
	return q.stats, q.err
}

func TestGetProfileSeasonStats(t *testing.T) {
	ctx := context.Background()
	mcUUID := uuid.New()

	t.Run("returns the stats of the profile", func(t *testing.T) {
		want := []*domain.ProfileSeasonStats{{SeasonName: "Season 2", Playtime: 5}, {SeasonName: "Season 1", Playtime: 9}}
		queries := &stubSeasonStatsQueries{stats: want}
		service := &profileService{storage: &stubMainStorage{queries: queries}}

		got, err := service.GetProfileSeasonStats(ctx, mcUUID)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 || got[0].SeasonName != "Season 2" || got[1].SeasonName != "Season 1" {
			t.Errorf("stats = %+v, want the stored order", got)
		}
		if queries.lastUUID != mcUUID {
			t.Errorf("queried %s, want %s", queries.lastUUID, mcUUID)
		}
	})

	t.Run("a storage failure is an internal error", func(t *testing.T) {
		queries := &stubSeasonStatsQueries{err: errors.New("db down")}
		service := &profileService{storage: &stubMainStorage{queries: queries}}

		_, err := service.GetProfileSeasonStats(ctx, mcUUID)
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusInternalServerError {
			t.Fatalf("error %v, want internal server error", err)
		}
	})
}
