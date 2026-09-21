package presenter

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
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

func TestPresentProfileResync(t *testing.T) {
	fine, broken := uuid.New(), uuid.New()
	resync := &domain.ProfileResync{Seasons: []*domain.SeasonResync{
		{SeasonID: fine, SeasonName: "Season 1", Parts: []domain.PartResync{
			{Part: domain.ResyncPartRole}, {Part: domain.ResyncPartAccess},
		}},
		{SeasonID: broken, SeasonName: "Season 2", Parts: []domain.PartResync{
			{Part: domain.ResyncPartRole},
			{Part: domain.ResyncPartAccess, Err: utils.NewInternalServerError("failed to add profile to whitelist", errors.New("dial tcp 10.0.0.5:9000"))},
		}},
	}}

	got := PresentProfileResync(resync)

	if got.OK || len(got.Seasons) != 2 {
		t.Fatalf("resync = %+v, want a failure over two seasons", got)
	}
	if !got.Seasons[0].OK || got.Seasons[0].SeasonName != "Season 1" || got.Seasons[0].Parts[0].Error != nil {
		t.Errorf("first season = %+v, want ok without errors", got.Seasons[0])
	}
	failed := got.Seasons[1].Parts[1]
	if got.Seasons[1].OK || failed.OK || failed.Part != "access" {
		t.Fatalf("second season = %+v, want the access part failed", got.Seasons[1])
	}
	if failed.Error == nil || *failed.Error != "failed to add profile to whitelist" {
		t.Errorf("error = %v, want the message without the original error", failed.Error)
	}
}

func TestPresentProfileResyncEmpty(t *testing.T) {
	got := PresentProfileResync(&domain.ProfileResync{})

	if !got.OK || got.Seasons == nil || len(got.Seasons) != 0 {
		t.Fatalf("resync = %+v, want ok with an empty list that is not null", got)
	}
}
