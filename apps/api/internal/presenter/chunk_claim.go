package presenter

import (
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentChunkClaims(claims []*domain.ChunkClaim) *responses.ChunkClaims {
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
				ID: claim.ProfileID, Username: claim.Profile.MinecraftUsername,
			})
		}
	}
	return res
}
