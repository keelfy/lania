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

func TestPresentAdminSeason_ShowsShellAddress(t *testing.T) {
	t.Parallel()

	address := "shell:9090"
	planURL := "https://plan.example.com"
	season := PresentAdminSeason(&domain.Season{
		ID: uuid.New(), Name: "Lania V", StartDate: time.Now(), ShellAddress: &address, PlanURL: &planURL,
	})

	if season.ShellAddress == nil || *season.ShellAddress != address {
		t.Fatalf("ShellAddress = %v, want %q", season.ShellAddress, address)
	}
	if season.PlanURL == nil || *season.PlanURL != planURL {
		t.Fatalf("PlanURL = %v, want %q", season.PlanURL, planURL)
	}
}

func TestPresentSeasons_GameVersionAndWorldURL(t *testing.T) {
	t.Parallel()

	version := "1.21.1"
	worldURL := "https://cdn.example.com/lania-v.zip"
	res := PresentSeasons([]*domain.Season{
		{ID: uuid.New(), Name: "Lania V", StartDate: time.Now(), GameVersion: &version, WorldURL: &worldURL},
	})

	if res[0].GameVersion == nil || *res[0].GameVersion != version {
		t.Errorf("GameVersion = %v, want %q", res[0].GameVersion, version)
	}
	if res[0].WorldURL == nil || *res[0].WorldURL != worldURL {
		t.Errorf("WorldURL = %v, want %q", res[0].WorldURL, worldURL)
	}
}

func TestPresentSeasonScreenshots(t *testing.T) {
	t.Parallel()

	title := "Spawn build"
	authorID := uuid.New()
	res := PresentSeasonScreenshots([]*domain.SeasonScreenshot{
		{
			ID: uuid.New(), Image: "s3://bucket/screenshot.jpg", Title: &title,
			Authors: []*domain.Profile{{ID: authorID, MinecraftUsername: "keelfy"}},
		},
		{ID: uuid.New(), Image: "s3://bucket/no-credit.jpg", Authors: nil},
	})

	if len(res) != 2 {
		t.Fatalf("got %d screenshots, want 2", len(res))
	}
	if res[0].Title == nil || *res[0].Title != title {
		t.Errorf("Title = %v, want %q", res[0].Title, title)
	}
	if len(res[0].Authors) != 1 || res[0].Authors[0].ID != authorID || res[0].Authors[0].Username != "keelfy" {
		t.Errorf("Authors = %+v", res[0].Authors)
	}
	if len(res[1].Authors) != 0 {
		t.Errorf("uncredited screenshot Authors = %+v, want empty", res[1].Authors)
	}
}
