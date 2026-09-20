package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/utils"
)

type SeasonHandler interface {
	// GetSeasons lists every season, the newest start first.
	GetSeasons(w http.ResponseWriter, r *http.Request)
}

type seasonHandler struct {
	seasonService services.SeasonService
}

func NewSeasonHandler(seasonService services.SeasonService) SeasonHandler {
	return &seasonHandler{seasonService: seasonService}
}

func (h *seasonHandler) GetSeasons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	seasons, err := h.seasonService.GetSeasons(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentSeasons(seasons, activeSeasonID))
}
