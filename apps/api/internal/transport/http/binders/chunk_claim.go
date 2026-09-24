package binders

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/utils"
)

// WorldQueryParam names the squaremap world, e.g. minecraft_overworld.
var WorldQueryParam = "world"

// BindClaimChunks reads the chunks of a claim or release request made by the signed-in user.
func BindClaimChunks(r *http.Request) (*commands.ClaimChunksCommand, error) {
	req := &requests.ClaimChunks{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}
	seasonID, err := BindPathVariableAsUUID(r, SeasonIDVariable)
	if err != nil {
		return nil, err
	}
	userID, err := utils.GetUserIDFromCtx(r.Context())
	if err != nil {
		return nil, err
	}

	chunks := make([]domain.ChunkPos, len(req.Chunks))
	for i, chunk := range req.Chunks {
		chunks[i] = domain.ChunkPos{X: chunk[0], Z: chunk[1]}
	}
	return &commands.ClaimChunksCommand{
		SeasonID:  seasonID,
		UserID:    userID,
		ProfileID: req.ProfileID,
		World:     strings.TrimSpace(req.World),
		Chunks:    chunks,
	}, nil
}
