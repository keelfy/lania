package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

type AccessHandler interface {
	CheckUsernames(w http.ResponseWriter, r *http.Request)
	ObtainFreeAccessForProfiles(w http.ResponseWriter, r *http.Request)
	RegisterProfilesForSeason(w http.ResponseWriter, r *http.Request)
	ObtainAccessForProfiles(w http.ResponseWriter, r *http.Request)
}

type accessHandler struct {
	profileService  services.ProfileService
	accessService   services.AccessService
	seasonService   services.SeasonService
	identityService services.IdentityService
	basketService   services.BasketService
	productService  services.ProductService
	mojangService   services.MojangService
	storage         storage.MainStorage
}

func NewAccessHandler(
	profileService services.ProfileService,
	accessService services.AccessService,
	seasonService services.SeasonService,
	identityService services.IdentityService,
	basketService services.BasketService,
	productService services.ProductService,
	mojangService services.MojangService,
	storage storage.MainStorage,
) AccessHandler {
	return &accessHandler{
		profileService:  profileService,
		accessService:   accessService,
		seasonService:   seasonService,
		identityService: identityService,
		basketService:   basketService,
		productService:  productService,
		mojangService:   mojangService,
		storage:         storage,
	}
}

var maxProfilesPerUser = config.GetMaxProfilesPerUser()

func (h *accessHandler) CheckUsernames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cmd, err := binders.BindCheckUsername(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	resList := make([]*responses.CheckUsername, len(cmd.Usernames))
	for i, username := range cmd.Usernames {
		profile, err := h.profileService.GetProfileByUsername(ctx, username)
		if err != nil && err.(*utils.CustomError).HttpStatus != http.StatusNotFound {
			utils.HttpError(ctx, w, err)
			return
		}

		status := responses.UsernameStatusAvailable
		hasAccess := false

		if profile != nil {
			authUserID := utils.GetUserIDFromContextOrNil(ctx)

			if profile.OwnerUserID == nil {
				status = responses.UsernameStatusAvailable
			} else if authUserID != nil && *profile.OwnerUserID == *authUserID {
				status = responses.UsernameStatusOwnedByYou
			} else {
				status = responses.UsernameStatusTaken
			}
		}

		if status == responses.UsernameStatusOwnedByYou {
			seasonID := cmd.SeasonID
			if seasonID == nil {
				id, err := h.seasonService.GetPrimarySeasonID(ctx)
				if err != nil {
					utils.HttpError(ctx, w, err)
					return
				}
				seasonID = &id
			}

			hasAccess, err = h.accessService.CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx, profile.MinecraftUUID, *seasonID)
			if err != nil {
				logger.Errorf(ctx, "failed to check if profile %v has access: %v", profile.MinecraftUsername, err)
			}
		}

		// Only a warning for the player: when Mojang cannot answer, the check still passes.
		premium, err := h.mojangService.IsPremiumUsername(ctx, username)
		if err != nil {
			logger.Warnf(ctx, "failed to check if username %v is premium: %v", username, err)
		}

		resList[i] = &responses.CheckUsername{
			Status:    status,
			HasAccess: hasAccess,
			Premium:   premium,
		}
	}

	utils.WriteHttpJsonResponse(ctx, w, resList)
}

func (h *accessHandler) ObtainFreeAccessForProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cmd, season, ok := h.bindFreeAccess(w, r)
	if !ok {
		return
	}
	if !season.Preregistration {
		utils.HttpError(ctx, w, utils.NewBadRequestError("pre registration is disabled", nil))
		return
	}

	h.grantFreeAccess(w, r, cmd, domain.AccessSourceFree)
}

// RegisterProfilesForSeason grants access to the primary season right away: the
// player only has to pick a username, no season pass involved.
func (h *accessHandler) RegisterProfilesForSeason(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cmd, season, ok := h.bindFreeAccess(w, r)
	if !ok {
		return
	}
	if !season.FreeRegistration {
		utils.HttpError(ctx, w, utils.NewBadRequestError("free registration is disabled", nil))
		return
	}

	h.grantFreeAccess(w, r, cmd, domain.AccessSourceRegistration)
}

func (h *accessHandler) bindFreeAccess(
	w http.ResponseWriter,
	r *http.Request,
) (*commands.ObtainAccessByUsernamesCommand, *domain.Season, bool) {
	ctx := r.Context()
	cmd, err := binders.BindObtainProfileAccesses(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return nil, nil, false
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return nil, nil, false
	}
	season, err := h.seasonService.GetSeasonByID(ctx, cmd.SeasonID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return nil, nil, false
	}
	return cmd, season, true
}

// grantFreeAccess creates the requested profiles if needed and hands them access
// to the requested season without any payment.
func (h *accessHandler) grantFreeAccess(
	w http.ResponseWriter,
	r *http.Request,
	cmd *commands.ObtainAccessByUsernamesCommand,
	source domain.AccessSource,
) {
	ctx := r.Context()

	if source == domain.AccessSourceRegistration {
		primarySeasonID, err := h.seasonService.GetPrimarySeasonID(ctx)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}
		if cmd.SeasonID != primarySeasonID {
			utils.HttpError(ctx, w, utils.NewBadRequestError("registration is only available for the primary season", nil))
			return
		}
	}

	for _, username := range cmd.Usernames {
		existingProfiles, err := h.profileService.GetProfilesByOwnerUserID(ctx, cmd.OwnerUserID)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}

		isAccessForExistingProfile := false
		for _, profile := range existingProfiles {
			if profile.MinecraftUsername == username {
				isAccessForExistingProfile = true
				break
			}
		}

		if !isAccessForExistingProfile && len(existingProfiles) >= maxProfilesPerUser {
			utils.HttpError(ctx, w, utils.NewBadRequestError(fmt.Sprintf("you can't have more than %d profiles", maxProfilesPerUser), nil))
			return
		}

		profile, err := h.profileService.GetOrCreateProfileByUsername(ctx, h.storage.Queries(), cmd.OwnerUserID, username)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}

		hasAccess, err := h.accessService.CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx, profile.MinecraftUUID, cmd.SeasonID)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}

		if hasAccess {
			utils.HttpError(ctx, w, utils.NewBadRequestError("profile already has access for the season", nil))
			return
		}

		err = h.accessService.ObtainAccessForProfile(ctx, cmd.SeasonID, profile, source)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *accessHandler) ObtainAccessForProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cmd, err := binders.BindObtainProfileAccesses(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	season, err := h.seasonService.GetSeasonByID(ctx, cmd.SeasonID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if !season.IsActive {
		utils.HttpError(ctx, w, utils.NewBadRequestError("season is not active", nil))
		return
	}

	for _, username := range cmd.Usernames {
		existingProfiles, err := h.profileService.GetProfilesByOwnerUserID(ctx, cmd.OwnerUserID)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}

		isAccessForExistingProfile := false
		for _, profile := range existingProfiles {
			if profile.MinecraftUsername == username {
				isAccessForExistingProfile = true
				break
			}
		}

		if !isAccessForExistingProfile && len(existingProfiles) >= maxProfilesPerUser {
			utils.HttpError(ctx, w, utils.NewBadRequestError(fmt.Sprintf("you can't have more than %d profiles", maxProfilesPerUser), nil))
			return
		}

		products, err := h.productService.GetProductsByCategory(ctx, domain.ProductCategoryUpgrade)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}

		var seasonAccessProduct *domain.Product
		for _, product := range products {
			var metadata domain.UpgradeProductMetadata
			err := json.Unmarshal(product.Metadata, &metadata)
			if err != nil {
				utils.HttpError(ctx, w, utils.NewInternalServerError("failed to unmarshal upgrade product metadata", err))
				return
			}

			if metadata.Action == domain.ProductUpgradeActionSeasonAccess {
				seasonAccessProduct = product
				break
			}
		}

		if seasonAccessProduct == nil {
			utils.HttpError(ctx, w, utils.NewBadRequestError("season access product not found", nil))
			return
		}

		profile, err := h.profileService.GetOrCreateProfileByUsername(ctx, h.storage.Queries(), cmd.OwnerUserID, username)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}

		hasAccess, err := h.accessService.CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx, profile.MinecraftUUID, cmd.SeasonID)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}

		if hasAccess {
			utils.HttpError(ctx, w, utils.NewBadRequestError("profile already has access for the season", nil))
			return
		}

		basketItems, err := h.basketService.GetBasketItemsByUserID(ctx, cmd.OwnerUserID)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}

		for _, basketItem := range basketItems {
			if basketItem.ProductID == seasonAccessProduct.ID && basketItem.ProfileID == profile.ID && basketItem.SeasonID == cmd.SeasonID {
				utils.HttpError(ctx, w, utils.NewBadRequestError("profile already has access for the season", nil))
				return
			}
		}

		err = h.basketService.AddBasketItem(ctx, h.storage.Queries(), &commands.AddBasketItemCommand{
			UserID:    cmd.OwnerUserID,
			ProductID: seasonAccessProduct.ID,
			ProfileID: profile.ID,
			SeasonID:  cmd.SeasonID,
			Quantity:  1,
		})
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
