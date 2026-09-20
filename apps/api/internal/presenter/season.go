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
			Preregistration: season.Preregistration, FreeRegistration: season.FreeRegistration,
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
			Preregistration: season.Preregistration, FreeRegistration: season.FreeRegistration,
		},
		SystemAddress:   season.SystemAddress,
		RCONPort:        season.RCONPort,
		RCONPasswordSet: season.RCONPassword != nil && *season.RCONPassword != "",
	}
}

func PresentAdminSeasons(seasons []*domain.Season) []*responses.AdminSeason {
	result := make([]*responses.AdminSeason, len(seasons))
	for i, season := range seasons {
		result[i] = PresentAdminSeason(season)
	}
	return result
}
