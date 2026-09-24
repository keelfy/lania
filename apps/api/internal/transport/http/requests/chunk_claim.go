package requests

import "github.com/google/uuid"

// ClaimChunks names chunks of one dimension of the world as [x, z] chunk indexes. ProfileID is ignored on the
// admin route.
type ClaimChunks struct {
	ProfileID uuid.UUID `json:"profileId"`
	Dimension string    `json:"dimension"`
	Chunks    [][2]int  `json:"chunks"`
}
