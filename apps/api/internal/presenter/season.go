package presenter

import (
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentSeasons(seasons []*domain.Season, activeSeasonID uuid.UUID) []*responses.Season {
	res := make([]*responses.Season, len(seasons))
	for i, season := range seasons {
		res[i] = &responses.Season{
			ID:           season.ID,
			Name:         season.Name,
			PreviewImage: season.PreviewImage,
			StartDate:    season.StartDate.UnixMilli(),
			EndDate:      timeToMillis(season.EndDate),
			IsActive:     season.ID == activeSeasonID,
		}
	}
	return res
}
