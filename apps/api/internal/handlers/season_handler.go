package handlers

import (
	"context"
	"net/http"

	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

type SeasonHandler interface {
	// GetSeasons lists every season, the newest start first.
	GetSeasons(w http.ResponseWriter, r *http.Request)
	GetAdminSeasons(w http.ResponseWriter, r *http.Request)
	CreateSeason(w http.ResponseWriter, r *http.Request)
	UpdateSeason(w http.ResponseWriter, r *http.Request)
	DeleteSeason(w http.ResponseWriter, r *http.Request)
	InitializePrimarySeason(ctx context.Context) error

	GetSeasonScreenshots(w http.ResponseWriter, r *http.Request)
	CreateSeasonScreenshot(w http.ResponseWriter, r *http.Request)
	UpdateSeasonScreenshot(w http.ResponseWriter, r *http.Request)
	DeleteSeasonScreenshot(w http.ResponseWriter, r *http.Request)
}

func (h *seasonHandler) InitializePrimarySeason(ctx context.Context) error {
	return h.seasonService.InitializePrimarySeason(ctx)
}

type seasonHandler struct {
	seasonService services.SeasonService
}

func NewSeasonHandler(seasonService services.SeasonService) SeasonHandler {
	return &seasonHandler{seasonService: seasonService}
}

func (h *seasonHandler) GetSeasons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seasons, err := h.seasonService.GetPublicSeasons(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentSeasons(seasons))
}

func (h *seasonHandler) GetAdminSeasons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seasons, err := h.seasonService.GetSeasons(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentAdminSeasons(seasons))
}

func (h *seasonHandler) CreateSeason(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cmd, err := binders.BindSaveSeason(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}
	season, err := h.seasonService.CreateSeason(ctx, cmd)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentAdminSeason(season))
}

func (h *seasonHandler) UpdateSeason(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cmd, err := binders.BindUpdateSeason(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}
	season, err := h.seasonService.UpdateSeason(ctx, cmd)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentAdminSeason(season))
}

func (h *seasonHandler) DeleteSeason(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seasonID, err := binders.BindPathVariableAsUUID(r, binders.SeasonIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := h.seasonService.DeleteSeason(ctx, seasonID); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *seasonHandler) GetSeasonScreenshots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seasonID, err := binders.BindPathVariableAsUUID(r, binders.SeasonIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	screenshots, err := h.seasonService.GetSeasonScreenshots(ctx, seasonID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentSeasonScreenshots(screenshots))
}

func (h *seasonHandler) CreateSeasonScreenshot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cmd, err := binders.BindSaveSeasonScreenshot(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}
	screenshot, err := h.seasonService.CreateSeasonScreenshot(ctx, cmd)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentSeasonScreenshots([]*domain.SeasonScreenshot{screenshot})[0])
}

func (h *seasonHandler) UpdateSeasonScreenshot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cmd, err := binders.BindUpdateSeasonScreenshot(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}
	screenshot, err := h.seasonService.UpdateSeasonScreenshot(ctx, cmd)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentSeasonScreenshots([]*domain.SeasonScreenshot{screenshot})[0])
}

func (h *seasonHandler) DeleteSeasonScreenshot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seasonID, err := binders.BindPathVariableAsUUID(r, binders.SeasonIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	screenshotID, err := binders.BindPathVariableAsUUID(r, binders.ScreenshotIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := h.seasonService.DeleteSeasonScreenshot(ctx, seasonID, screenshotID); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
