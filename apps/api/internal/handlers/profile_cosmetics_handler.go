package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
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
	luckpermsService        services.LuckpermsService
	storage                 storage.MainStorage
}

func NewProfileCosmeticsHandler(
	profileCosmeticsService services.ProfileCosmeticsService,
	profileService services.ProfileService,
	luckpermsService services.LuckpermsService,
	storage storage.MainStorage,
) ProfileCosmeticsHandler {
	return &profileCosmeticsHandler{
		profileCosmeticsService: profileCosmeticsService,
		profileService:          profileService,
		luckpermsService:        luckpermsService,
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

	if profile.OwnerUserID != authUserID {
		utils.HttpError(ctx, w, utils.NewForbiddenError("only owner can access profile cosmetics options", nil))
		return
	}

	seasonID := config.GetActiveSeasonID()

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

	res := presenter.PresentProfileCosmeticOptions(nameColorOptions, glythPrefixOptions, specialPrefixOptions)
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

	if profile.OwnerUserID != authUserID {
		utils.HttpError(ctx, w, utils.NewForbiddenError("only owner can access profile cosmetics options", nil))
		return
	}

	seasonID := config.GetActiveSeasonID()
	nameColorOption, err := h.profileCosmeticsService.GetProfileNameColorOptionByIDAndProfileID(ctx, req.OptionID, profileID, &seasonID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profilePrefixes, err := h.profileCosmeticsService.GetProfilePrefixes(ctx, profileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	var glythPrefix *domain.NamePrefix
	var specialPrefix *domain.NamePrefix

	for _, prefix := range profilePrefixes {
		switch prefix.Type {
		case domain.ProfilePrefixTypeGlyth:
			glythPrefix = prefix.NamePrefix
		case domain.ProfilePrefixTypeSpecial:
			specialPrefix = prefix.NamePrefix
		}
	}

	formattedPrefix := h.profileCosmeticsService.GetProfileFullPrefix(ctx, nameColorOption.NameColor, glythPrefix, specialPrefix)

	err = h.storage.BeginTx(ctx, func(queries sql.Queries) error {
		err = h.profileCosmeticsService.SelectProfileNameColor(ctx, queries, profileID, nameColorOption.NameColorID)
		if err != nil {
			return err
		}

		err = h.luckpermsService.SetUserPrefixByMinecraftUUID(ctx, queries, profile.MinecraftUUID, formattedPrefix)
		if err != nil {
			return err
		}

		return nil
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

	if profile.OwnerUserID != authUserID {
		utils.HttpError(ctx, w, utils.NewForbiddenError("only owner can access profile cosmetics options", nil))
		return
	}

	seasonID := config.GetActiveSeasonID()
	var namePrefixOption *domain.ProfileNamePrefixOption

	if req.OptionID != uuid.Nil {
		namePrefixOption, err = h.profileCosmeticsService.GetProfileNamePrefixOptionByIDAndProfileIDAndType(ctx, req.OptionID, profileID, prefixType, &seasonID)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}
	}

	profilePrefixes, err := h.profileCosmeticsService.GetProfilePrefixes(ctx, profileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	var glythPrefix *domain.NamePrefix
	var specialPrefix *domain.NamePrefix

	for _, prefix := range profilePrefixes {
		switch prefix.Type {
		case domain.ProfilePrefixTypeGlyth:
			glythPrefix = prefix.NamePrefix
		case domain.ProfilePrefixTypeSpecial:
			specialPrefix = prefix.NamePrefix
		}
	}

	if namePrefixOption != nil {
		switch prefixType {
		case domain.ProfilePrefixTypeGlyth:
			glythPrefix = namePrefixOption.NamePrefix
		case domain.ProfilePrefixTypeSpecial:
			specialPrefix = namePrefixOption.NamePrefix
		}
	} else {
		switch prefixType {
		case domain.ProfilePrefixTypeGlyth:
			glythPrefix = nil
		case domain.ProfilePrefixTypeSpecial:
			specialPrefix = nil
		}
	}

	formattedPrefix := h.profileCosmeticsService.GetProfileFullPrefix(ctx, profile.NameColor, glythPrefix, specialPrefix)

	err = h.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if namePrefixOption != nil {
			err = h.profileCosmeticsService.SelectProfileNamePrefix(ctx, queries, profileID, namePrefixOption.NamePrefixID, prefixType)
			if err != nil {
				return err
			}
		} else {
			err = h.profileCosmeticsService.ClearProfilePrefixByType(ctx, queries, profileID, prefixType)
			if err != nil {
				return err
			}
		}

		err = h.luckpermsService.SetUserPrefixByMinecraftUUID(ctx, queries, profile.MinecraftUUID, formattedPrefix)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
