package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

type AdminProfileHandler interface {
	GetProfiles(w http.ResponseWriter, r *http.Request)
	GetProfileDetails(w http.ResponseWriter, r *http.Request)
	TransferProfile(w http.ResponseWriter, r *http.Request)
	ReleaseProfile(w http.ResponseWriter, r *http.Request)
	SetProfileRole(w http.ResponseWriter, r *http.Request)
}

type adminProfileHandler struct {
	profileService      services.ProfileService
	adminProfileService services.AdminProfileService
}

func NewAdminProfileHandler(
	profileService services.ProfileService,
	adminProfileService services.AdminProfileService,
) AdminProfileHandler {
	return &adminProfileHandler{
		profileService:      profileService,
		adminProfileService: adminProfileService,
	}
}

// GetProfiles lists every profile, owned or not. The search matches usernames by prefix.
func (h *adminProfileHandler) GetProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination, err := binders.BindPagination(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	filter := domain.ProfileFilter{Search: binders.BindSearch(r)}
	profiles, count, err := h.profileService.GetPublicProfiles(ctx, filter, pagination, binders.BindSort(r))
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	content := make([]*responses.AdminProfile, len(profiles))
	for i, profile := range profiles {
		content[i] = presenter.PresentAdminProfile(profile)
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentPaginatedResponse(pagination, count, content))
}

func (h *adminProfileHandler) GetProfileDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profileID, err := binders.BindPathVariableAsUUID(r, binders.ProfileIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profile, owner, err := h.adminProfileService.GetProfileDetails(ctx, profileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentAdminProfileDetails(profile, owner))
}

func (h *adminProfileHandler) TransferProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindTransferProfileOwner(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	profile, owner, err := h.adminProfileService.TransferProfile(ctx, cmd.ProfileID, cmd.Email)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentAdminProfileDetails(profile, owner))
}

func (h *adminProfileHandler) ReleaseProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profileID, err := binders.BindPathVariableAsUUID(r, binders.ProfileIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profile, err := h.adminProfileService.ReleaseProfile(ctx, profileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentAdminProfileDetails(profile, nil))
}

func (h *adminProfileHandler) SetProfileRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindSetProfileRole(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	profile, owner, err := h.adminProfileService.SetProfileRole(ctx, cmd.ProfileID, cmd.Role)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentAdminProfileDetails(profile, owner))
}
