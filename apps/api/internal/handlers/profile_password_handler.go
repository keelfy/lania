package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

type ProfilePasswordHandler interface {
	// SetPassword gives the owner of an unlicensed profile a new in-game password in a season. It fails with 409
	// and the message player_not_registered when the profile never registered in game there.
	SetPassword(w http.ResponseWriter, r *http.Request)
}

type profilePasswordHandler struct {
	profilePasswordService services.ProfilePasswordService
}

func NewProfilePasswordHandler(profilePasswordService services.ProfilePasswordService) ProfilePasswordHandler {
	return &profilePasswordHandler{profilePasswordService: profilePasswordService}
}

func (h *profilePasswordHandler) SetPassword(w http.ResponseWriter, r *http.Request) {
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
	req, err := binders.BindSetProfilePassword(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := h.profilePasswordService.SetOwnedPassword(ctx, authUserID, profileID, req.SeasonID, req.Password); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
