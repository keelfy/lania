package presenter

import (
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

// PresentChunkClaims lists the owners with their cosmetics; an owner missing from cosmetics wears none.
func PresentChunkClaims(claims []*domain.ChunkClaim, cosmetics map[uuid.UUID]*domain.ProfileCosmetics) *responses.ChunkClaims {
	res := &responses.ChunkClaims{
		Claims:   make([]*responses.ChunkClaim, len(claims)),
		Profiles: make([]*responses.ChunkClaimProfile, 0),
	}
	seen := make(map[uuid.UUID]bool)
	for i, claim := range claims {
		res.Claims[i] = &responses.ChunkClaim{
			X: claim.Chunk.X, Z: claim.Chunk.Z, ProfileID: claim.ProfileID, ClaimedAt: claim.ClaimedAt.UnixMilli(),
		}
		if !seen[claim.ProfileID] {
			seen[claim.ProfileID] = true
			res.Profiles = append(res.Profiles, &responses.ChunkClaimProfile{
				ID:        claim.ProfileID,
				Username:  claim.Profile.MinecraftUsername,
				Cosmetics: PresentProfileCosmetics(cosmetics[claim.ProfileID]),
			})
		}
	}
	return res
}
