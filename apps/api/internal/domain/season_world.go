package domain

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

// OverworldDimension is the squaremap world name of the overworld, where claims are on by default.
const OverworldDimension = "minecraft_overworld"

// SeasonWorld is one server of a season network: survival, farms, creative. Each has its own squaremap and its
// own chunk claims.
type SeasonWorld struct {
	ID       uuid.UUID
	SeasonID uuid.UUID
	// Slug is the key of the world page, /worlds/<slug>; unique within the season.
	Slug string
	// Name is the same in every language.
	Name string
	// PreviewImage is an s3://bucket/key location, the same form as Season.PreviewImage.
	PreviewImage *string
	// MapURL is the squaremap of the world server. The world has no map page while it is empty.
	MapURL *string
	// ClaimLimit is how many chunks one profile may hold in the world, over its claim dimensions together.
	ClaimLimit int
	// ClaimDimensions are the squaremap world names where chunks can be claimed; none makes the map view-only.
	ClaimDimensions []string
	// PlanServer is the Plan name of the world server, where playtime for claims is counted. Nil counts the
	// whole season network.
	PlanServer *string
	// ClaimMinPlaytimeHours is how long a profile must have played on the world server before it can claim
	// chunks; zero lets anyone with season access claim.
	ClaimMinPlaytimeHours int
	Position              int
}

// DefaultClaimMinPlaytimeHours is the playtime a world asks of claimers when it sets none of its own.
const DefaultClaimMinPlaytimeHours = 5

// ClaimMinPlaytime is the playtime on the world server a profile needs to claim chunks.
func (w *SeasonWorld) ClaimMinPlaytime() time.Duration {
	return time.Duration(w.ClaimMinPlaytimeHours) * time.Hour
}

// ClaimsIn tells whether chunks of the dimension can be claimed in the world.
func (w *SeasonWorld) ClaimsIn(dimension string) bool {
	return w.MapURL != nil && slices.Contains(w.ClaimDimensions, dimension)
}
