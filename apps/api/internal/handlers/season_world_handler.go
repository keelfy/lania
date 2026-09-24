package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

type SeasonWorldHandler interface {
	GetSeasonWorlds(w http.ResponseWriter, r *http.Request)
	GetSeasonWorld(w http.ResponseWriter, r *http.Request)
	CreateSeasonWorld(w http.ResponseWriter, r *http.Request)
	UpdateSeasonWorld(w http.ResponseWriter, r *http.Request)
	DeleteSeasonWorld(w http.ResponseWriter, r *http.Request)
}

type seasonWorldHandler struct {
	worldService services.SeasonWorldService
}

func NewSeasonWorldHandler(worldService services.SeasonWorldService) SeasonWorldHandler {
	return &seasonWorldHandler{worldService: worldService}
}

func (h *seasonWorldHandler) GetSeasonWorlds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seasonID, err := binders.BindPathVariableAsUUID(r, binders.SeasonIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	worlds, err := h.worldService.GetSeasonWorlds(ctx, seasonID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentSeasonWorlds(worlds))
}

func (h *seasonWorldHandler) GetSeasonWorld(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	worldID, err := binders.BindPathVariableAsUUID(r, binders.WorldIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	world, err := h.worldService.GetSeasonWorld(ctx, worldID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentSeasonWorld(world))
}

func (h *seasonWorldHandler) CreateSeasonWorld(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cmd, err := binders.BindCreateSeasonWorld(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}
	world, err := h.worldService.CreateSeasonWorld(ctx, cmd)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentSeasonWorld(world))
}

func (h *seasonWorldHandler) UpdateSeasonWorld(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cmd, err := binders.BindUpdateSeasonWorld(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}
	world, err := h.worldService.UpdateSeasonWorld(ctx, cmd)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentSeasonWorld(world))
}

func (h *seasonWorldHandler) DeleteSeasonWorld(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	worldID, err := binders.BindPathVariableAsUUID(r, binders.WorldIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := h.worldService.DeleteSeasonWorld(ctx, worldID); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
