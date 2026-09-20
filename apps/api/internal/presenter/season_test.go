package presenter

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

func TestPresentSeasons(t *testing.T) {
	t.Parallel()
	image := "s3://bucket/lania-5-preview.jpg"
	start := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	active, past := uuid.New(), uuid.New()

	res := PresentSeasons([]*domain.Season{
		{
			ID: active, Name: "Lania V", PreviewImage: &image, StartDate: start,
			IsActive: true, IsPrimary: true, Preregistration: true, FreeRegistration: true,
		},
		{ID: past, Name: "Glory Mine 7", StartDate: start, EndDate: &end},
	})

	if len(res) != 2 {
		t.Fatalf("got %d seasons, want 2", len(res))
	}
	if res[0].Name != "Lania V" || !res[0].IsActive || !res[0].IsPrimary ||
		!res[0].Preregistration || !res[0].FreeRegistration ||
		res[0].PreviewImage == nil || *res[0].PreviewImage != image || res[0].EndDate != nil {
		t.Errorf("primary season = %+v", res[0])
	}
	if res[1].IsActive || res[1].PreviewImage != nil || res[1].EndDate == nil || *res[1].EndDate != end.UnixMilli() {
		t.Errorf("past season = %+v", res[1])
	}
}

func TestPresentAdminSeason_HidesRCONPassword(t *testing.T) {
	t.Parallel()

	password := "do-not-return"
	season := PresentAdminSeason(&domain.Season{
		ID: uuid.New(), Name: "Lania V", StartDate: time.Now(), RCONPassword: &password,
	})

	if !season.RCONPasswordSet {
		t.Fatal("RCONPasswordSet = false, want true")
	}
}
