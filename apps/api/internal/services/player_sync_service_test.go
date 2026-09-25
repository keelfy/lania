package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	sql "github.com/lania-smp/backend/internal/storage/main"
)

type syncQueries struct {
	sql.Queries
	playtimeSeen   *time.Time
	playtimeSeason uuid.UUID
	profileSeen    *time.Time
}

func (q *syncQueries) UpsertProfilePlaytime(_ context.Context, _, seasonID uuid.UUID, _ int64, lastSeenAt *time.Time) error {
	q.playtimeSeason = seasonID
	q.playtimeSeen = lastSeenAt
	return nil
}

func (q *syncQueries) UpdateProfileSeenAt(_ context.Context, _ uuid.UUID, _, lastSeenAt *time.Time) error {
	q.profileSeen = lastSeenAt
	return nil
}

func TestSyncPlayerStoresLastSeenInTheSeason(t *testing.T) {
	seasonID := uuid.New()
	end := time.Date(2026, 9, 20, 18, 30, 0, 0, time.UTC).UnixMilli()
	queries := &syncQueries{}

	err := syncPlayer(context.Background(), queries, seasonID, uuid.New(), &domain.Playtime{TotalPlaytime: 1000, LastSessionEnd: &end})
	if err != nil {
		t.Fatal(err)
	}

	if queries.playtimeSeason != seasonID || queries.playtimeSeen == nil || queries.playtimeSeen.UnixMilli() != end {
		t.Errorf("season last seen = %v in %s, want the end of the last session in the season", queries.playtimeSeen, queries.playtimeSeason)
	}
	if queries.profileSeen == nil || queries.profileSeen.UnixMilli() != end {
		t.Errorf("profile last seen = %v, want the same date for the admin panel", queries.profileSeen)
	}
}

func TestSyncPlayerWithoutSessionKeepsStoredDate(t *testing.T) {
	queries := &syncQueries{}

	if err := syncPlayer(context.Background(), queries, uuid.New(), uuid.New(), &domain.Playtime{TotalPlaytime: 1000}); err != nil {
		t.Fatal(err)
	}
	if queries.playtimeSeen != nil || queries.profileSeen != nil {
		t.Errorf("last seen = %v and %v, want nil so the stored dates stay", queries.playtimeSeen, queries.profileSeen)
	}
}

type legacySyncQueries struct {
	sql.Queries
	legacy map[uuid.UUID]uuid.UUID
}

func (q *legacySyncQueries) FindLegacyMinecraftUUIDs(context.Context, uuid.UUIDs) (map[uuid.UUID]uuid.UUID, error) {
	return q.legacy, nil
}

type playtimeMinecraftService struct {
	MinecraftService
	playtimes map[uuid.UUID]*domain.Playtime
}

func (s *playtimeMinecraftService) GetPlaytimesInSeason(context.Context, uuid.UUID, uuid.UUIDs) (map[uuid.UUID]*domain.Playtime, error) {
	return s.playtimes, nil
}

func TestAddLegacyPlaytimesSumsBothUUIDs(t *testing.T) {
	offline, mojang := uuid.New(), uuid.New()
	first, oldEnd, newEnd := int64(100), int64(500), int64(900)
	minecraft := &playtimeMinecraftService{playtimes: map[uuid.UUID]*domain.Playtime{
		offline: {TotalPlaytime: 3000, FirstSessionStart: &first, LastSessionEnd: &oldEnd},
		mojang:  {TotalPlaytime: 2000, LastSessionEnd: &newEnd},
	}}
	service := &playerSyncService{
		storage:          &stubMainStorage{queries: &legacySyncQueries{legacy: map[uuid.UUID]uuid.UUID{offline: mojang}}},
		minecraftService: minecraft,
	}
	// Only the new UUID played since the cursor.
	playtimes := map[uuid.UUID]*domain.Playtime{mojang: {TotalPlaytime: 2000, LastSessionEnd: &newEnd}}

	if err := service.addLegacyPlaytimes(context.Background(), uuid.New(), playtimes); err != nil {
		t.Fatal(err)
	}

	got := playtimes[mojang]
	if len(playtimes) != 1 || got.TotalPlaytime != 5000 || *got.FirstSessionStart != first || *got.LastSessionEnd != newEnd {
		t.Errorf("playtimes = %v, want 5000 ms under the Mojang UUID from the first offline session to the last new one", playtimes)
	}
}
