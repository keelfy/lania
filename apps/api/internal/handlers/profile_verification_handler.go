package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

type ProfileVerificationHandler interface {
	// StartVerification opens a verification request of the profile for its owner.
	StartVerification(w http.ResponseWriter, r *http.Request)
	// ConfirmVerification checks the code the owner saw on the kick screen.
	ConfirmVerification(w http.ResponseWriter, r *http.Request)
	// VerificationLogin is called by the proxy plugin on every login, behind the API key.
	VerificationLogin(w http.ResponseWriter, r *http.Request)
}

type profileVerificationHandler struct {
	profileVerificationService services.ProfileVerificationService
}

func NewProfileVerificationHandler(profileVerificationService services.ProfileVerificationService) ProfileVerificationHandler {
	return &profileVerificationHandler{profileVerificationService: profileVerificationService}
}

func (h *profileVerificationHandler) StartVerification(w http.ResponseWriter, r *http.Request) {
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

	expiresAt, err := h.profileVerificationService.StartVerification(ctx, authUserID, profileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, &responses.ProfileVerification{ExpiresAt: expiresAt.UnixMilli()})
}

func (h *profileVerificationHandler) ConfirmVerification(w http.ResponseWriter, r *http.Request) {
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
	req, err := binders.BindConfirmProfileVerification(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := h.profileVerificationService.ConfirmVerification(ctx, authUserID, profileID, req.Code); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *profileVerificationHandler) VerificationLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := binders.BindVerificationLogin(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	code, err := h.profileVerificationService.IssueLoginCode(ctx, req.UUID, req.Username, req.OnlineMode)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, &responses.VerificationLogin{Code: code})
}
