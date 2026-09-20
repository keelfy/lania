package presenter

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

func TestPresentProfileStats(t *testing.T) {
	start := time.UnixMilli(1_700_000_000_000)
	end := time.UnixMilli(1_710_000_000_000)
	stats := []*domain.ProfileSeasonStats{
		{SeasonID: uuid.New(), SeasonName: "Season 2", StartDate: end, IsActive: true, IsPrimary: true, Playtime: 3_600_000},
		{SeasonID: uuid.New(), SeasonName: "Season 1", StartDate: start, EndDate: &end, Playtime: 1_800_000},
	}

	got := PresentProfileStats(stats)

	if got.TotalPlaytime != 5_400_000 {
		t.Errorf("total = %d, want 5400000", got.TotalPlaytime)
	}
	if len(got.Seasons) != 2 || got.Seasons[0].SeasonName != "Season 2" || got.Seasons[1].SeasonName != "Season 1" {
		t.Fatalf("seasons = %+v, want the given order", got.Seasons)
	}
	if got.Seasons[0].EndDate != nil {
		t.Errorf("a running season must have no end date, got %v", *got.Seasons[0].EndDate)
	}
	if got.Seasons[1].EndDate == nil || *got.Seasons[1].EndDate != end.UnixMilli() || got.Seasons[1].StartDate != start.UnixMilli() {
		t.Errorf("dates = %+v, want start and end in milliseconds", got.Seasons[1])
	}
}

func TestPresentProfileStatsEmpty(t *testing.T) {
	got := PresentProfileStats(nil)
	if got.TotalPlaytime != 0 || got.Seasons == nil || len(got.Seasons) != 0 {
		t.Errorf("stats = %+v, want zero total and an empty, non-nil list", got)
	}
}
