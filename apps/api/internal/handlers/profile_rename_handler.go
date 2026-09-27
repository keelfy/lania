package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

type ProfileRenameHandler interface {
	// ChangeUsername renames the profile for its owner.
	ChangeUsername(w http.ResponseWriter, r *http.Request)
}

type profileRenameHandler struct {
	profileRenameService services.ProfileRenameService
}

func NewProfileRenameHandler(profileRenameService services.ProfileRenameService) ProfileRenameHandler {
	return &profileRenameHandler{profileRenameService: profileRenameService}
}

func (h *profileRenameHandler) ChangeUsername(w http.ResponseWriter, r *http.Request) {
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
	req, err := binders.BindChangeProfileUsername(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if _, err := h.profileRenameService.ChangeOwnedUsername(ctx, authUserID, profileID, req.Username, req.Unlicensed); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
