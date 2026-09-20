package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

type AdminGrantHandler interface {
	GetGrants(w http.ResponseWriter, r *http.Request)
	GrantProduct(w http.ResponseWriter, r *http.Request)
	// GetCosmetics lists the name colors and name prefixes that can be granted.
	GetCosmetics(w http.ResponseWriter, r *http.Request)
	GrantCosmetic(w http.ResponseWriter, r *http.Request)
	RevokeGrant(w http.ResponseWriter, r *http.Request)
}

type adminGrantHandler struct {
	adminGrantService services.AdminGrantService
}

func NewAdminGrantHandler(adminGrantService services.AdminGrantService) AdminGrantHandler {
	return &adminGrantHandler{adminGrantService: adminGrantService}
}

func (h *adminGrantHandler) GetGrants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profileID, err := binders.BindPathVariableAsUUID(r, binders.ProfileIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	h.writeGrants(w, r, profileID)
}

func (h *adminGrantHandler) GrantProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindGrantProduct(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	if err := h.adminGrantService.GrantProduct(ctx, cmd.ProfileID, cmd.ProductID, cmd.SeasonID); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	h.writeGrants(w, r, cmd.ProfileID)
}

func (h *adminGrantHandler) GetCosmetics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	catalog, err := h.adminGrantService.GetCosmeticsCatalog(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentAdminCosmeticsCatalog(catalog))
}

func (h *adminGrantHandler) GrantCosmetic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindGrantCosmetic(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	if err := h.adminGrantService.GrantCosmetic(ctx, cmd.ProfileID, cmd.Type, cmd.ItemID, cmd.PrefixType, cmd.SeasonID); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	h.writeGrants(w, r, cmd.ProfileID)
}

func (h *adminGrantHandler) RevokeGrant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindRevokeGrant(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	if err := h.adminGrantService.RevokeGrant(ctx, cmd.ProfileID, cmd.GrantType, cmd.GrantID); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	h.writeGrants(w, r, cmd.ProfileID)
}

// writeGrants answers with the current grants of the profile, so the client needs no second request.
func (h *adminGrantHandler) writeGrants(w http.ResponseWriter, r *http.Request, profileID uuid.UUID) {
	ctx := r.Context()

	grants, err := h.adminGrantService.GetProfileGrants(ctx, profileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentAdminGrants(grants))
}
