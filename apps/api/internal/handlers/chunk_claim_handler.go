package handlers

import (
	"context"
	"net/http"

	"github.com/lania-smp/backend/internal/commands"
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
}

func NewChunkClaimHandler(chunkClaimService services.ChunkClaimService) ChunkClaimHandler {
	return &chunkClaimHandler{chunkClaimService: chunkClaimService}
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
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentChunkClaims(claims))
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
