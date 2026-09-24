package presenter

import (
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentSeasonWorld(world *domain.SeasonWorld) *responses.SeasonWorld {
	dimensions := world.ClaimDimensions
	if dimensions == nil {
		dimensions = []string{}
	}
	return &responses.SeasonWorld{
		ID: world.ID, SeasonID: world.SeasonID, Slug: world.Slug, Name: world.Name,
		PreviewImage: world.PreviewImage, MapURL: world.MapURL,
		ClaimLimit: world.ClaimLimit, ClaimDimensions: dimensions, Position: world.Position,
	}
}

func PresentSeasonWorlds(worlds []*domain.SeasonWorld) []*responses.SeasonWorld {
	result := make([]*responses.SeasonWorld, len(worlds))
	for i, world := range worlds {
		result[i] = PresentSeasonWorld(world)
	}
	return result
}
