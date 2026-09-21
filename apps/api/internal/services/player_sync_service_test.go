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
