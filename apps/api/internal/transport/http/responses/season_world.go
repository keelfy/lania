package responses

import "github.com/google/uuid"

type SeasonWorld struct {
	ID       uuid.UUID `json:"id"`
	SeasonID uuid.UUID `json:"seasonId"`
	Slug     string    `json:"slug"`
	// Name is the same in every language.
	Name         string  `json:"name"`
	PreviewImage *string `json:"previewImage,omitempty"`
	// MapURL is the squaremap of the world server, missing when the world has no map.
	MapURL     *string `json:"mapUrl,omitempty"`
	ClaimLimit int     `json:"claimLimit"`
	// ClaimDimensions are the squaremap world names where chunks can be claimed; empty for a view-only map.
	ClaimDimensions []string `json:"claimDimensions"`
	// PlanServer is the Plan name of the world server; missing counts playtime on the whole season network.
	PlanServer *string `json:"planServer,omitempty"`
	// ClaimMinPlaytimeHours is the playtime on the world server a profile needs to claim chunks.
	ClaimMinPlaytimeHours int `json:"claimMinPlaytimeHours"`
	Position              int `json:"position"`
}
