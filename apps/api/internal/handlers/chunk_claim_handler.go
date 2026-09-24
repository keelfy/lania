package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

type ChunkClaimHandler interface {
	GetChunkClaims(w http.ResponseWriter, r *http.Request)
	ClaimChunks(w http.ResponseWriter, r *http.Request)
	ReleaseChunks(w http.ResponseWriter, r *http.Request)
	AdminReleaseChunks(w http.ResponseWriter, r *http.Request)
}

type chunkClaimHandler struct {
	chunkClaimService services.ChunkClaimService
	cosmeticsService  services.ProfileCosmeticsService
}

func NewChunkClaimHandler(chunkClaimService services.ChunkClaimService, cosmeticsService services.ProfileCosmeticsService) ChunkClaimHandler {
	return &chunkClaimHandler{chunkClaimService: chunkClaimService, cosmeticsService: cosmeticsService}
}

func (h *chunkClaimHandler) GetChunkClaims(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seasonID, err := binders.BindPathVariableAsUUID(r, binders.SeasonIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	world := r.URL.Query().Get(binders.WorldQueryParam)
	if world == "" {
		utils.HttpError(ctx, w, utils.NewBadRequestError("world is required", nil))
		return
	}
	claims, err := h.chunkClaimService.GetActiveClaims(ctx, seasonID, world)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	// Owners are shown the way they look in the season; claims still show when cosmetics cannot be read.
	profileIDs := make(uuid.UUIDs, 0, len(claims))
	for _, claim := range claims {
		profileIDs = append(profileIDs, claim.ProfileID)
	}
	cosmetics, err := h.cosmeticsService.GetProfilesCosmetics(ctx, profileIDs, seasonID)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE COSMETICS] Failed to get chunk claim owners cosmetics: %v", err)
		cosmetics = make(map[uuid.UUID]*domain.ProfileCosmetics)
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentChunkClaims(claims, cosmetics))
}

func (h *chunkClaimHandler) ClaimChunks(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, h.chunkClaimService.ClaimChunks)
}

func (h *chunkClaimHandler) ReleaseChunks(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, h.chunkClaimService.ReleaseChunks)
}

func (h *chunkClaimHandler) AdminReleaseChunks(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, h.chunkClaimService.AdminReleaseChunks)
}

// handle binds and validates the chunks, runs the change and answers 204.
func (h *chunkClaimHandler) handle(w http.ResponseWriter, r *http.Request, change func(ctx context.Context, cmd *commands.ClaimChunksCommand) error) {
	ctx := r.Context()
	cmd, err := binders.BindClaimChunks(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}
	if err := change(ctx, cmd); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
