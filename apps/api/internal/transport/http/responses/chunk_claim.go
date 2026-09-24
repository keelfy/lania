package responses

import "github.com/google/uuid"

type ChunkClaim struct {
	X         int       `json:"x"`
	Z         int       `json:"z"`
	ProfileID uuid.UUID `json:"profileId"`
	ClaimedAt int64     `json:"claimedAt"`
}

type ChunkClaimProfile struct {
	ID        uuid.UUID         `json:"id"`
	Username  string            `json:"username"`
	Cosmetics *ProfileCosmetics `json:"cosmetics"`
}

// ChunkClaims lists each claiming profile once instead of repeating it on every chunk it holds.
type ChunkClaims struct {
	Claims   []*ChunkClaim        `json:"claims"`
	Profiles []*ChunkClaimProfile `json:"profiles"`
}
