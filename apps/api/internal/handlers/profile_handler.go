package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

type ProfileHandler interface {
	GetUserProfiles(w http.ResponseWriter, r *http.Request)
	GetUserProfileDetails(w http.ResponseWriter, r *http.Request)
	GetPublicProfiles(w http.ResponseWriter, r *http.Request)
}

type profileHandler struct {
	profileService   services.ProfileService
	accessService    services.AccessService
	minecraftService services.MinecraftService
	mojangService    services.MojangService
	cosmeticsService services.ProfileCosmeticsService
}

func NewProfileHandler(
	profileService services.ProfileService,
	accessService services.AccessService,
	minecraftService services.MinecraftService,
	mojangService services.MojangService,
	cosmeticsService services.ProfileCosmeticsService,
) ProfileHandler {
	return &profileHandler{
		profileService:   profileService,
		accessService:    accessService,
		minecraftService: minecraftService,
		mojangService:    mojangService,
		cosmeticsService: cosmeticsService,
	}
}

var activeSeasonID = config.GetActiveSeasonID()

func (h *profileHandler) GetPublicProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination, err := binders.BindPagination(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	sort := binders.BindSort(r)

	profiles, err := h.profileService.GetPublicProfiles(ctx, pagination, sort)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	count, err := h.profileService.CountPublicProfiles(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	mcUUIDs := make(uuid.UUIDs, len(profiles))
	for i, profile := range profiles {
		mcUUIDs[i] = profile.MinecraftUUID
	}

	onlineMap, err := h.minecraftService.GetOnlineStatusByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[SHELL] Failed to get online status by minecraft uuid: %v", err)
		onlineMap = make(map[uuid.UUID]bool)
	}

	mojangUUIDs, err := h.mojangService.GetMojangUUIDsByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[MOJANG] Failed to get mojang uuids by minecraft uuids: %v", err)
		mojangUUIDs = make(map[uuid.UUID]uuid.UUID)
	}

	res := make([]*responses.PublicProfile, len(profiles))
	for i, profile := range profiles {
		role, err := h.profileService.GetProfileRole(ctx, profile.MinecraftUUID)
		if err != nil {
			logger.Warnf(ctx, "failed to get profile role: %v", err)
		}
		profile.Role = role

		isOnline, ok := onlineMap[profile.MinecraftUUID]
		if !ok {
			isOnline = false
		}

		var nullableMojangUUID *uuid.UUID
		if mojangUUID, ok := mojangUUIDs[profile.MinecraftUUID]; ok {
			nullableMojangUUID = &mojangUUID
		}

		cosmetics := presenter.PresentProfileCosmetics(profile.NameColor, nil, nil)
		res[i] = presenter.PresentPublicProfile(profile, nullableMojangUUID, cosmetics, isOnline)
	}

	paginated := presenter.PresentPaginatedResponse(pagination, count, res)
	utils.WriteHttpJsonResponse(ctx, w, paginated)
}

func (h *profileHandler) GetUserProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profiles, err := h.profileService.GetProfilesByOwnerUserID(ctx, userID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	mcUUIDs := make(uuid.UUIDs, len(profiles))
	for i, profile := range profiles {
		mcUUIDs[i] = profile.MinecraftUUID
	}

	accesses, err := h.accessService.GetAccessesByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	mojangUUIDs, err := h.mojangService.GetMojangUUIDsByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[MOJANG] Failed to get mojang uuids by minecraft uuids: %v", err)
		mojangUUIDs = make(map[uuid.UUID]uuid.UUID)
	}

	res := make([]*responses.Profile, len(profiles))
	for i, profile := range profiles {
		accessStatus := domain.AccessStatusInactive
		for _, access := range accesses[profile.MinecraftUUID] {
			if access.SeasonID == activeSeasonID {
				accessStatus = domain.AccessStatusActive
				break
			}
			accessStatus = domain.AccessStatusExpired
		}

		var nullableMojangUUID *uuid.UUID
		if mojangUUID, ok := mojangUUIDs[profile.MinecraftUUID]; ok {
			nullableMojangUUID = &mojangUUID
		}

		profilePrefixes, err := h.cosmeticsService.GetProfilePrefixes(ctx, profile.ID)
		if err != nil {
			logger.Errorf(ctx, "[PROFILE COSMETICS] Failed to get profile prefixes: %v", err)
			profilePrefixes = []*domain.ProfilePrefix{}
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

		cosmetics := presenter.PresentProfileCosmetics(profile.NameColor, glythPrefix, specialPrefix)
		res[i] = presenter.PresentProfile(profile, nullableMojangUUID, accessStatus, cosmetics)
	}

	utils.WriteHttpJsonResponse(ctx, w, res)
}

func (h *profileHandler) GetUserProfileDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	mcUUIDs := uuid.UUIDs{profile.MinecraftUUID}

	accesses, err := h.accessService.GetAccessesByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[ACCESS] Failed to get accesses by minecraft uuid: %v", err)
		accesses = make(map[uuid.UUID][]*domain.ProfileAccess)
	}

	playtimes, err := h.minecraftService.GetPlaytimeByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[SHELL] Failed to get playtime by minecraft uuid: %v", err)
		playtimes = make(map[uuid.UUID]*domain.Playtime)
	}

	seasonsPlaytimes, err := h.profileService.GetSeasonsPlaytimeByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE] Failed to sum seasons playtime by minecraft uuid: %v", err)
		seasonsPlaytimes = make(map[uuid.UUID]int64)
	}

	onlineMap, err := h.minecraftService.GetOnlineStatusByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[SHELL] Failed to get online status by minecraft uuid: %v", err)
		onlineMap = make(map[uuid.UUID]bool)
	}

	mojangUUIDs, err := h.mojangService.GetMojangUUIDsByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[MOJANG] Failed to get mojang uuids by minecraft uuids: %v", err)
		mojangUUIDs = make(map[uuid.UUID]uuid.UUID)
	}

	var nullableMojangUUID *uuid.UUID
	isModelSlim := false
	if mojangUUID, ok := mojangUUIDs[profile.MinecraftUUID]; ok {
		nullableMojangUUID = &mojangUUID

		isModelSlim, err = h.mojangService.IsPlayerModelSlim(ctx, mojangUUID)
		if err != nil {
			logger.Errorf(ctx, "[MOJANG] Failed to get player model slim: %v", err)
			isModelSlim = false
		}
	}

	accessStatus := domain.AccessStatusInactive
	for _, access := range accesses[profile.MinecraftUUID] {
		if access.SeasonID == activeSeasonID {
			accessStatus = domain.AccessStatusActive
			break
		}
		accessStatus = domain.AccessStatusExpired
	}

	isOnline, ok := onlineMap[profile.MinecraftUUID]
	if !ok {
		isOnline = false
	}

	playtime := playtimes[profile.MinecraftUUID]
	if playtime == nil {
		playtime = &domain.Playtime{
			TotalPlaytime:     0,
			FirstSessionStart: nil,
			LastSessionEnd:    nil,
		}
	}

	profilePrefixes, err := h.cosmeticsService.GetProfilePrefixes(ctx, profile.ID)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE COSMETICS] Failed to get profile prefixes: %v", err)
		profilePrefixes = []*domain.ProfilePrefix{}
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

	role, err := h.profileService.GetProfileRole(ctx, profile.MinecraftUUID)
	if err != nil {
		logger.Errorf(ctx, "Failed to get profile role: %v", err)
		role = domain.RolePlayer
	}
	profile.Role = role

	cosmetics := presenter.PresentProfileCosmetics(profile.NameColor, glythPrefix, specialPrefix)
	res := presenter.PresentProfileDetails(profile, nullableMojangUUID, accessStatus, playtime, seasonsPlaytimes[profile.MinecraftUUID], isOnline, isModelSlim, cosmetics)
	utils.WriteHttpJsonResponse(ctx, w, res)
}
