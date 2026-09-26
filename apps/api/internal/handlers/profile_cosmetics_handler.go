package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/utils"
)

type ProfileCosmeticsHandler interface {
	GetProfileCosmeticOptions(w http.ResponseWriter, r *http.Request)
	SelectProfileNameColor(w http.ResponseWriter, r *http.Request)
	SelectProfileNamePrefix(w http.ResponseWriter, r *http.Request)
}

type profileCosmeticsHandler struct {
	profileCosmeticsService services.ProfileCosmeticsService
	profileService          services.ProfileService
	minecraftService        services.MinecraftService
	seasonService           services.SeasonService
	storage                 storage.MainStorage
}

func NewProfileCosmeticsHandler(
	profileCosmeticsService services.ProfileCosmeticsService,
	profileService services.ProfileService,
	minecraftService services.MinecraftService,
	seasonService services.SeasonService,
	storage storage.MainStorage,
) ProfileCosmeticsHandler {
	return &profileCosmeticsHandler{
		profileCosmeticsService: profileCosmeticsService,
		profileService:          profileService,
		minecraftService:        minecraftService,
		seasonService:           seasonService,
		storage:                 storage,
	}
}

func (h *profileCosmeticsHandler) GetProfileCosmeticOptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profileID, err := binders.BindPathVariableAsUUID(r, "profileId")
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profile, err := h.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if profile.OwnerUserID == nil || *profile.OwnerUserID != authUserID {
		utils.HttpError(ctx, w, utils.NewForbiddenError("only owner can access profile cosmetics options", nil))
		return
	}

	seasonID, err := seasonIDFromRequest(r, h.seasonService)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	var nameColorOptions []*domain.ProfileNameColorOption
	var glythPrefixOptions []*domain.ProfileNamePrefixOption
	specialPrefixOptions := []*domain.ProfileNamePrefixOption{}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		nameColorOptions, err = h.profileCosmeticsService.GetProfileNameColorOptions(ctx, profileID, &seasonID)
		if err != nil {
			utils.LogCustomError(ctx, err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		glythPrefixOptions, err = h.profileCosmeticsService.GetProfileNamePrefixOptionsByProfileIDAndType(ctx, profileID, domain.ProfilePrefixTypeGlyth, &seasonID)
		if err != nil {
			utils.LogCustomError(ctx, err)
		}
	}()

	wg.Wait()

	if mayUseAnyCosmetic(ctx, profile) {
		nameColorOptions, err = h.profileCosmeticsService.AppendCatalogNameColorOptions(ctx, nameColorOptions, profileID)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}
		glythPrefixOptions, err = h.profileCosmeticsService.AppendCatalogNamePrefixOptions(ctx, glythPrefixOptions, profileID, domain.ProfilePrefixTypeGlyth)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}
	}

	res := presenter.PresentProfileCosmeticOptions(nameColorOptions, glythPrefixOptions, specialPrefixOptions, utils.GetLocaleFromCtx(ctx))
	utils.WriteHttpJsonResponse(ctx, w, res)
}

func (h *profileCosmeticsHandler) SelectProfileNameColor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profileID, err := binders.BindPathVariableAsUUID(r, "profileId")
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	req := &requests.SelectCosmeticOption{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	profile, err := h.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if profile.OwnerUserID == nil || *profile.OwnerUserID != authUserID {
		utils.HttpError(ctx, w, utils.NewForbiddenError("only owner can access profile cosmetics options", nil))
		return
	}

	seasonID, err := activeSeasonIDFromRequest(r, h.seasonService)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	nameColorOption, err := h.profileCosmeticsService.GetProfileNameColorOptionByIDAndProfileID(ctx, req.OptionID, profileID, &seasonID)
	if isNotFound(err) && mayUseAnyCosmetic(ctx, profile) {
		nameColorOption, err = h.profileCosmeticsService.GetCatalogNameColorOption(ctx, req.OptionID, profileID)
	}
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	err = h.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if err := h.profileCosmeticsService.SelectProfileNameColor(ctx, queries, profileID, seasonID, nameColorOption.NameColorID); err != nil {
			return err
		}
		return h.pushPrefix(ctx, queries, profile, seasonID)
	})
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *profileCosmeticsHandler) SelectProfileNamePrefix(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profileID, err := binders.BindPathVariableAsUUID(r, "profileId")
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	rawPrefixType, err := binders.BindPathVariable(r, "type")
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	prefixType := domain.ProfilePrefixType(rawPrefixType)

	req := &requests.SelectCosmeticOption{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	profile, err := h.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if profile.OwnerUserID == nil || *profile.OwnerUserID != authUserID {
		utils.HttpError(ctx, w, utils.NewForbiddenError("only owner can access profile cosmetics options", nil))
		return
	}

	seasonID, err := activeSeasonIDFromRequest(r, h.seasonService)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	// A nil UUID clears the prefix of the type.
	var namePrefixOption *domain.ProfileNamePrefixOption
	if req.OptionID != uuid.Nil {
		namePrefixOption, err = h.profileCosmeticsService.GetProfileNamePrefixOptionByIDAndProfileIDAndType(ctx, req.OptionID, profileID, prefixType, &seasonID)
		if isNotFound(err) && mayUseAnyCosmetic(ctx, profile) {
			namePrefixOption, err = h.profileCosmeticsService.GetCatalogNamePrefixOption(ctx, req.OptionID, profileID, prefixType)
		}
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}
	}

	err = h.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if namePrefixOption != nil {
			err = h.profileCosmeticsService.SelectProfileNamePrefix(ctx, queries, profileID, seasonID, namePrefixOption.NamePrefixID, prefixType)
		} else {
			err = h.profileCosmeticsService.ClearProfilePrefixByType(ctx, queries, profileID, seasonID, prefixType)
		}
		if err != nil {
			return err
		}
		return h.pushPrefix(ctx, queries, profile, seasonID)
	})
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// mayUseAnyCosmetic tells whether the profile may show any cosmetic on the site without a grant.
// Owners and admins of the site may do it on their profiles, and so may profiles with the owner or admin role.
func mayUseAnyCosmetic(ctx context.Context, profile *domain.Profile) bool {
	if profile.Role.IsAdmin() {
		return true
	}
	session, err := utils.GetSessionFromCtx(ctx)
	if err != nil {
		return false
	}
	return domain.RoleFromMetadata(session.Identity.MetadataPublic).IsAdmin()
}

func isNotFound(err error) bool {
	return utils.MapCustomErrorToHttpStatus(err) == http.StatusNotFound
}

// pushPrefix builds the chat prefix from the selection in the season, as queries sees it, and sends it to the season server.
func (h *profileCosmeticsHandler) pushPrefix(ctx context.Context, queries sql.Queries, profile *domain.Profile, seasonID uuid.UUID) error {
	prefix, err := h.profileCosmeticsService.GetProfileChatPrefixWithQueries(ctx, queries, profile.ID, seasonID)
	if err != nil {
		return err
	}
	return h.minecraftService.SetPrefixInSeason(ctx, seasonID, profile.MinecraftUUID, prefix)
}
