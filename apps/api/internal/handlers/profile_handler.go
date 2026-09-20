package handlers

import (
	"context"
	"net/http"
	"strconv"

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
	GetProfileDetailsByUsername(w http.ResponseWriter, r *http.Request)
	GetPublicProfiles(w http.ResponseWriter, r *http.Request)
	GetTopPlaytimeProfiles(w http.ResponseWriter, r *http.Request)
	GetProfilesStats(w http.ResponseWriter, r *http.Request)
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

const (
	defaultTopPlaytimeLimit = 10
	maxTopPlaytimeLimit     = 20
)

func (h *profileHandler) GetPublicProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination, err := binders.BindPagination(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	sort := binders.BindSort(r)
	filter := binders.BindProfileFilter(r)

	profiles, count, err := h.profileService.GetPublicProfiles(ctx, filter, pagination, sort)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	mcUUIDs := minecraftUUIDsOf(profiles)
	seasonsPlaytimes, err := h.profileService.GetSeasonsPlaytimeByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE] Failed to sum seasons playtime by minecraft uuid: %v", err)
		seasonsPlaytimes = make(map[uuid.UUID]int64)
	}

	res := h.presentPublicProfiles(ctx, profiles, seasonsPlaytimes)

	paginated := presenter.PresentPaginatedResponse(pagination, count, res)
	utils.WriteHttpJsonResponse(ctx, w, paginated)
}

func (h *profileHandler) GetTopPlaytimeProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, err := strconv.Atoi(binders.BindOptionalQueryParamAsString(r, "limit", ""))
	if err != nil || limit < 1 {
		limit = defaultTopPlaytimeLimit
	} else if limit > maxTopPlaytimeLimit {
		limit = maxTopPlaytimeLimit
	}

	top, err := h.profileService.GetTopPlaytimeProfiles(ctx, config.GetPrimarySeasonID(), limit)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profiles := make([]*domain.Profile, len(top))
	playtimes := make(map[uuid.UUID]int64, len(top))
	for i, playtime := range top {
		profiles[i] = playtime.Profile
		playtimes[playtime.MinecraftUUID] = playtime.Playtime
	}

	utils.WriteHttpJsonResponse(ctx, w, h.presentPublicProfiles(ctx, profiles, playtimes))
}

func (h *profileHandler) GetProfilesStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.profileService.GetProfilesStats(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, &responses.ProfilesStats{
		Total:       stats.Total,
		Online:      stats.Online,
		NewLastWeek: stats.NewLastWeek,
	})
}

func splitProfilePrefixes(prefixes []*domain.ProfilePrefix) (glyth *domain.NamePrefix, special *domain.NamePrefix) {
	for _, prefix := range prefixes {
		switch prefix.Type {
		case domain.ProfilePrefixTypeGlyth:
			glyth = prefix.NamePrefix
		case domain.ProfilePrefixTypeSpecial:
			special = prefix.NamePrefix
		}
	}
	return glyth, special
}

func minecraftUUIDsOf(profiles []*domain.Profile) uuid.UUIDs {
	mcUUIDs := make(uuid.UUIDs, len(profiles))
	for i, profile := range profiles {
		mcUUIDs[i] = profile.MinecraftUUID
	}
	return mcUUIDs
}

// presentPublicProfiles adds online status, role and Mojang UUID to profiles. Playtimes are in milliseconds.
func (h *profileHandler) presentPublicProfiles(ctx context.Context, profiles []*domain.Profile, playtimes map[uuid.UUID]int64) []*responses.PublicProfile {
	mcUUIDs := minecraftUUIDsOf(profiles)

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

	h.profileService.ApplyProfileRoles(ctx, profiles)

	profileIDs := make(uuid.UUIDs, len(profiles))
	for i, profile := range profiles {
		profileIDs[i] = profile.ID
	}
	prefixes, err := h.cosmeticsService.GetProfilesPrefixes(ctx, profileIDs)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE COSMETICS] Failed to get profiles prefixes: %v", err)
		prefixes = make(map[uuid.UUID][]*domain.ProfilePrefix)
	}

	res := make([]*responses.PublicProfile, len(profiles))
	for i, profile := range profiles {
		var nullableMojangUUID *uuid.UUID
		if mojangUUID, ok := mojangUUIDs[profile.MinecraftUUID]; ok {
			nullableMojangUUID = &mojangUUID
		}

		glythPrefix, specialPrefix := splitProfilePrefixes(prefixes[profile.ID])
		cosmetics := presenter.PresentProfileCosmetics(profile.NameColor, glythPrefix, specialPrefix)
		res[i] = presenter.PresentPublicProfile(profile, nullableMojangUUID, cosmetics, onlineMap[profile.MinecraftUUID], playtimes[profile.MinecraftUUID])
	}
	return res
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
			if access.SeasonID == config.GetPrimarySeasonID() {
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

		glythPrefix, specialPrefix := splitProfilePrefixes(profilePrefixes)

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

	h.writeProfileDetails(w, r, profile)
}

func (h *profileHandler) GetProfileDetailsByUsername(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username, err := binders.BindPathVariable(r, "username")
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profile, err := h.profileService.GetProfileByUsername(ctx, username)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	h.writeProfileDetails(w, r, profile)
}

func (h *profileHandler) writeProfileDetails(w http.ResponseWriter, r *http.Request, profile *domain.Profile) {
	ctx := r.Context()

	mcUUIDs := uuid.UUIDs{profile.MinecraftUUID}

	accesses, err := h.accessService.GetAccessesByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[ACCESS] Failed to get accesses by minecraft uuid: %v", err)
		accesses = make(map[uuid.UUID][]*domain.ProfileAccess)
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
		if access.SeasonID == config.GetPrimarySeasonID() {
			accessStatus = domain.AccessStatusActive
			break
		}
		accessStatus = domain.AccessStatusExpired
	}

	isOnline, ok := onlineMap[profile.MinecraftUUID]
	if !ok {
		isOnline = false
	}

	profilePrefixes, err := h.cosmeticsService.GetProfilePrefixes(ctx, profile.ID)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE COSMETICS] Failed to get profile prefixes: %v", err)
		profilePrefixes = []*domain.ProfilePrefix{}
	}

	glythPrefix, specialPrefix := splitProfilePrefixes(profilePrefixes)

	h.profileService.ApplyProfileRoles(ctx, []*domain.Profile{profile})

	cosmetics := presenter.PresentProfileCosmetics(profile.NameColor, glythPrefix, specialPrefix)
	res := presenter.PresentProfileDetails(profile, nullableMojangUUID, accessStatus, seasonsPlaytimes[profile.MinecraftUUID], isOnline, isModelSlim, cosmetics)
	utils.WriteHttpJsonResponse(ctx, w, res)
}
