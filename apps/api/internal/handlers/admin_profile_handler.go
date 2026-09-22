package handlers

import (
	"net/http"

	"github.com/google/uuid"
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
	PreviewMergeProfiles(w http.ResponseWriter, r *http.Request)
	MergeProfiles(w http.ResponseWriter, r *http.Request)
	GetProfileMerges(w http.ResponseWriter, r *http.Request)
}

type adminProfileHandler struct {
	profileService           services.ProfileService
	adminProfileService      services.AdminProfileService
	adminProfileMergeService services.AdminProfileMergeService
}

func NewAdminProfileHandler(
	profileService services.ProfileService,
	adminProfileService services.AdminProfileService,
	adminProfileMergeService services.AdminProfileMergeService,
) AdminProfileHandler {
	return &adminProfileHandler{
		profileService:           profileService,
		adminProfileService:      adminProfileService,
		adminProfileMergeService: adminProfileMergeService,
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
	profiles, count, err := h.profileService.GetPublicProfiles(ctx, filter, pagination, binders.BindSort(r), uuid.Nil)
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

// PreviewMergeProfiles shows what merging the profile in the path into the target would do, blockers included.
// It never changes data.
func (h *adminProfileHandler) PreviewMergeProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindPreviewMergeProfiles(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	summary, err := h.adminProfileMergeService.PreviewMergeProfiles(ctx, cmd.SourceProfileID, cmd.TargetProfileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentProfileMergeSummary(summary, nil))
}

// MergeProfiles moves every site record of the profile in the path into the target profile, then deletes it.
func (h *adminProfileHandler) MergeProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindMergeProfiles(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	summary, resync, err := h.adminProfileMergeService.MergeProfiles(ctx, cmd.SourceProfileID, cmd.TargetProfileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentProfileMergeSummary(summary, resync))
}

// GetProfileMerges lists every profile merged into the profile in the path, newest first.
func (h *adminProfileHandler) GetProfileMerges(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profileID, err := binders.BindPathVariableAsUUID(r, binders.ProfileIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	merges, err := h.adminProfileMergeService.ListMergesInto(ctx, profileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentProfileMerges(merges))
}
