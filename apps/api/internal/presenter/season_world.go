package presenter

import (
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentSeasonWorld(world *domain.SeasonWorld) *responses.SeasonWorld {
	return &responses.SeasonWorld{
		ID: world.ID, SeasonID: world.SeasonID, Slug: world.Slug, Name: world.Name,
		PreviewImage: world.PreviewImage, MapURL: world.MapURL,
		ClaimLimit: world.ClaimLimit, ClaimDimensions: nonNilStrings(world.ClaimDimensions),
		HiddenDimensions: nonNilStrings(world.HiddenDimensions),
		PlanServer:       world.PlanServer, ClaimMinPlaytimeHours: world.ClaimMinPlaytimeHours, Position: world.Position,
	}
}

func PresentSeasonWorlds(worlds []*domain.SeasonWorld) []*responses.SeasonWorld {
	result := make([]*responses.SeasonWorld, len(worlds))
	for i, world := range worlds {
		result[i] = PresentSeasonWorld(world)
	}
	return result
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
