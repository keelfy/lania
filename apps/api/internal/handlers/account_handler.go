package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

type AccountHandler interface {
	DeleteAccount(w http.ResponseWriter, r *http.Request)
	// ReleaseProfile releases a game profile of the signed in user, so anybody can claim it.
	ReleaseProfile(w http.ResponseWriter, r *http.Request)
}

type accountHandler struct {
	accountService services.AccountService
}

func NewAccountHandler(accountService services.AccountService) AccountHandler {
	return &accountHandler{accountService: accountService}
}

// DeleteAccount deletes the account of the signed in user and releases the game profiles.
// It fails with 403 and the message session_refresh_required when the user signed in too long ago.
func (h *accountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, err := utils.GetSessionFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	userID, err := uuid.Parse(session.Identity.Id)
	if err != nil {
		utils.HttpError(ctx, w, utils.ErrJWTMissing)
		return
	}
	// A session without the time of the sign in is treated as too old.
	authenticatedAt := session.GetAuthenticatedAt()

	siteRole := domain.RoleFromMetadata(session.Identity.MetadataPublic)
	if err := h.accountService.DeleteAccount(ctx, userID, siteRole, authenticatedAt); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *accountHandler) ReleaseProfile(w http.ResponseWriter, r *http.Request) {
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

	if err := h.accountService.ReleaseOwnedProfile(ctx, authUserID, profileID); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
