package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

type ProfileResyncHandler interface {
	// ResyncProfile is for the owner of the profile, and it has a cooldown. It answers 200 with a report
	// per season, also when some servers failed.
	ResyncProfile(w http.ResponseWriter, r *http.Request)
	// AdminResyncProfile is for admins, for any profile and without a cooldown.
	AdminResyncProfile(w http.ResponseWriter, r *http.Request)
}

type profileResyncHandler struct {
	profileResyncService services.ProfileResyncService
}

func NewProfileResyncHandler(profileResyncService services.ProfileResyncService) ProfileResyncHandler {
	return &profileResyncHandler{profileResyncService: profileResyncService}
}

func (h *profileResyncHandler) ResyncProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profileID, err := binders.BindPathVariableAsUUID(r, binders.ProfileIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	resync, err := h.profileResyncService.ResyncOwnedProfile(ctx, profileID, authUserID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentProfileResync(resync))
}

func (h *profileResyncHandler) AdminResyncProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profileID, err := binders.BindPathVariableAsUUID(r, binders.ProfileIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	resync, err := h.profileResyncService.ResyncProfile(ctx, profileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentProfileResync(resync))
}
