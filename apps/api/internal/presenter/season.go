package presenter

import (
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentSeasons(seasons []*domain.Season) []*responses.Season {
	res := make([]*responses.Season, len(seasons))
	for i, season := range seasons {
		res[i] = &responses.Season{
			ID: season.ID, Name: season.Name, PreviewImage: season.PreviewImage,
			StartDate: season.StartDate.UnixMilli(), EndDate: timeToMillis(season.EndDate),
			PublicAddress: season.PublicAddress,
			IsActive:      season.IsActive, IsPrimary: season.IsPrimary,
			OnlineAvailable: season.IsActive && season.HasServer,
			Preregistration: season.Preregistration, FreeRegistration: season.FreeRegistration,
			GameVersion: season.GameVersion, WorldURL: season.WorldURL,
		}
	}
	return res
}

func PresentAdminSeason(season *domain.Season) *responses.AdminSeason {
	return &responses.AdminSeason{
		Season: responses.Season{
			ID: season.ID, Name: season.Name, PreviewImage: season.PreviewImage,
			StartDate: season.StartDate.UnixMilli(), EndDate: timeToMillis(season.EndDate),
			PublicAddress: season.PublicAddress,
			IsActive:      season.IsActive, IsPrimary: season.IsPrimary,
			OnlineAvailable: season.IsActive && season.HasServer,
			Preregistration: season.Preregistration, FreeRegistration: season.FreeRegistration,
			GameVersion: season.GameVersion, WorldURL: season.WorldURL,
		},
		ShellAddress: season.ShellAddress,
		PlanURL:      season.PlanURL,
	}
}

func PresentAdminSeasons(seasons []*domain.Season) []*responses.AdminSeason {
	result := make([]*responses.AdminSeason, len(seasons))
	for i, season := range seasons {
		result[i] = PresentAdminSeason(season)
	}
	return result
}

func PresentSeasonScreenshots(screenshots []*domain.SeasonScreenshot) []*responses.SeasonScreenshot {
	result := make([]*responses.SeasonScreenshot, len(screenshots))
	for i, screenshot := range screenshots {
		authors := make([]responses.ScreenshotAuthor, len(screenshot.Authors))
		for j, author := range screenshot.Authors {
			authors[j] = responses.ScreenshotAuthor{ID: author.ID, Username: author.MinecraftUsername}
		}
		result[i] = &responses.SeasonScreenshot{
			ID: screenshot.ID, Image: screenshot.Image, Title: screenshot.Title,
			Position: screenshot.Position, Authors: authors,
		}
	}
	return result
}
