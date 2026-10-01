package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

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
	GetProfileStats(w http.ResponseWriter, r *http.Request)
}

type profileHandler struct {
	profileService   services.ProfileService
	accessService    services.AccessService
	seasonService    services.SeasonService
	minecraftService services.MinecraftService
	mojangService    services.MojangService
	cosmeticsService services.ProfileCosmeticsService
	renameService    services.ProfileRenameService
}

func NewProfileHandler(
	profileService services.ProfileService,
	accessService services.AccessService,
	seasonService services.SeasonService,
	minecraftService services.MinecraftService,
	mojangService services.MojangService,
	cosmeticsService services.ProfileCosmeticsService,
	renameService services.ProfileRenameService,
) ProfileHandler {
	return &profileHandler{
		profileService:   profileService,
		accessService:    accessService,
		seasonService:    seasonService,
		minecraftService: minecraftService,
		mojangService:    mojangService,
		cosmeticsService: cosmeticsService,
		renameService:    renameService,
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

	seasonID, err := seasonIDFromRequest(r, h.seasonService)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	profiles, count, err := h.profileService.GetPublicProfiles(ctx, filter, pagination, sort, seasonID)
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

	res := h.presentPublicProfiles(ctx, profiles, seasonsPlaytimes, seasonID)

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

	seasonID, err := seasonIDFromRequest(r, h.seasonService)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	top, err := h.profileService.GetTopPlaytimeProfiles(ctx, seasonID, limit)
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

	utils.WriteHttpJsonResponse(ctx, w, h.presentPublicProfiles(ctx, profiles, playtimes, seasonID))
}

func (h *profileHandler) GetProfilesStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	seasonID, err := seasonIDFromRequest(r, h.seasonService)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	stats, err := h.profileService.GetProfilesStats(ctx, seasonID)
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

func (h *profileHandler) GetProfileStats(w http.ResponseWriter, r *http.Request) {
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

	stats, err := h.profileService.GetProfileSeasonStats(ctx, profile.MinecraftUUID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentProfileStats(stats))
}

// onlineStatusInSeason tells who is online on the server of the season. Everybody is offline when the season
// has no running server or the server cannot be reached.
func (h *profileHandler) onlineStatusInSeason(ctx context.Context, seasonID uuid.UUID, mcUUIDs uuid.UUIDs) map[uuid.UUID]bool {
	onlineMap, err := h.minecraftService.GetOnlineStatusInSeason(ctx, seasonID, mcUUIDs)
	if err != nil {
		if !errors.Is(err, services.ErrOnlineUnavailable) {
			logger.Errorf(ctx, "[SHELL] Failed to get online status by minecraft uuid: %v", err)
		}
		return make(map[uuid.UUID]bool)
	}
	return onlineMap
}

// skinsInSeason returns the skins players chose in game in the season. Nobody has one when the server cannot be
// reached, so the site falls back to the licensed skins.
func (h *profileHandler) skinsInSeason(ctx context.Context, seasonID uuid.UUID, mcUUIDs uuid.UUIDs) map[uuid.UUID]*domain.PlayerSkin {
	skins, err := h.minecraftService.GetPlayerSkinsInSeason(ctx, seasonID, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[SHELL] Failed to get player skins by minecraft uuid: %v", err)
		return make(map[uuid.UUID]*domain.PlayerSkin)
	}
	return skins
}

// seenAt returns the date of the player in lastSeen, nil when the date is unknown.
func seenAt(lastSeen map[uuid.UUID]time.Time, mcUUID uuid.UUID) *time.Time {
	if seen, ok := lastSeen[mcUUID]; ok {
		return &seen
	}
	return nil
}

func minecraftUUIDsOf(profiles []*domain.Profile) uuid.UUIDs {
	mcUUIDs := make(uuid.UUIDs, len(profiles))
	for i, profile := range profiles {
		mcUUIDs[i] = profile.MinecraftUUID
	}
	return mcUUIDs
}

// presentPublicProfiles adds online status, Mojang UUID and what the profiles wear in the season.
// Playtimes are in milliseconds.
func (h *profileHandler) presentPublicProfiles(ctx context.Context, profiles []*domain.Profile, playtimes map[uuid.UUID]int64, seasonID uuid.UUID) []*responses.PublicProfile {
	mcUUIDs := minecraftUUIDsOf(profiles)

	onlineMap := h.onlineStatusInSeason(ctx, seasonID, mcUUIDs)
	skins := h.skinsInSeason(ctx, seasonID, mcUUIDs)

	lastSeenMap, err := h.profileService.GetProfilesLastSeenInSeason(ctx, mcUUIDs, seasonID)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE] Failed to get last seen dates in season: %v", err)
		lastSeenMap = make(map[uuid.UUID]time.Time)
	}

	mojangUUIDs, err := h.mojangService.GetMojangUUIDsByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		logger.Errorf(ctx, "[MOJANG] Failed to get mojang uuids by minecraft uuids: %v", err)
		mojangUUIDs = make(map[uuid.UUID]uuid.UUID)
	}

	profileIDs := make(uuid.UUIDs, len(profiles))
	for i, profile := range profiles {
		profileIDs[i] = profile.ID
	}
	cosmetics, err := h.cosmeticsService.GetProfilesCosmetics(ctx, profileIDs, seasonID)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE COSMETICS] Failed to get profiles cosmetics: %v", err)
		cosmetics = make(map[uuid.UUID]*domain.ProfileCosmetics)
	}

	res := make([]*responses.PublicProfile, len(profiles))
	for i, profile := range profiles {
		var nullableMojangUUID *uuid.UUID
		if mojangUUID, ok := mojangUUIDs[profile.MinecraftUUID]; ok {
			nullableMojangUUID = &mojangUUID
		}

		res[i] = presenter.PresentPublicProfile(profile, nullableMojangUUID, presenter.PresentProfileCosmetics(cosmetics[profile.ID], utils.GetLocaleFromCtx(ctx)), onlineMap[profile.MinecraftUUID], playtimes[profile.MinecraftUUID], seenAt(lastSeenMap, profile.MinecraftUUID))
		res[i].Skin = presenter.PresentPlayerSkin(skins[profile.MinecraftUUID])
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

	seasons, err := h.seasonService.GetSeasons(ctx)
	if err != nil {
		logger.Errorf(ctx, "[SEASON] Failed to get seasons: %v", err)
		seasons = nil
	}

	primarySeasonID, err := h.seasonService.GetPrimarySeasonID(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	seasonID, err := seasonIDFromRequest(r, h.seasonService)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	profileIDs := make(uuid.UUIDs, len(profiles))
	for i, profile := range profiles {
		profileIDs[i] = profile.ID
	}
	cosmetics, err := h.cosmeticsService.GetProfilesCosmetics(ctx, profileIDs, seasonID)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE COSMETICS] Failed to get profiles cosmetics: %v", err)
		cosmetics = make(map[uuid.UUID]*domain.ProfileCosmetics)
	}
	skins := h.skinsInSeason(ctx, seasonID, mcUUIDs)
	usernameChangeAvailableAt, err := h.renameService.GetUsernameChangeAvailableAt(ctx, profileIDs)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE RENAME] Failed to get when nicknames can be changed: %v", err)
		usernameChangeAvailableAt = make(map[uuid.UUID]time.Time)
	}

	res := make([]*responses.Profile, len(profiles))
	for i, profile := range profiles {
		accessStatus := domain.AccessStatusInactive
		for _, access := range accesses[profile.MinecraftUUID] {
			if access.SeasonID == primarySeasonID {
				accessStatus = domain.AccessStatusActive
				break
			}
			accessStatus = domain.AccessStatusExpired
		}

		var nullableMojangUUID *uuid.UUID
		if mojangUUID, ok := mojangUUIDs[profile.MinecraftUUID]; ok {
			nullableMojangUUID = &mojangUUID
		}

		res[i] = presenter.PresentProfile(profile, nullableMojangUUID, accessStatus, domain.SeasonAccessStatuses(seasons, accesses[profile.MinecraftUUID]), presenter.PresentProfileCosmetics(cosmetics[profile.ID], utils.GetLocaleFromCtx(ctx)))
		res[i].Skin = presenter.PresentPlayerSkin(skins[profile.MinecraftUUID])
		// The cooldown is only for unlicensed profiles: a licensed one follows its name on Mojang.
		if nullableMojangUUID == nil || *nullableMojangUUID != profile.MinecraftUUID {
			res[i].UsernameChangeAvailableAt = profileTimeMillis(usernameChangeAvailableAt, profile.ID)
			res[i].UsernameChangeCooldownDays = int(config.GetNicknameChangeCooldown() / (24 * time.Hour))
		}
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

	seasonID, err := seasonIDFromRequest(r, h.seasonService)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	onlineMap := h.onlineStatusInSeason(ctx, seasonID, mcUUIDs)
	skins := h.skinsInSeason(ctx, seasonID, mcUUIDs)

	lastSeenMap, err := h.profileService.GetProfilesLastSeenInSeason(ctx, mcUUIDs, seasonID)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE] Failed to get last seen dates in season: %v", err)
		lastSeenMap = make(map[uuid.UUID]time.Time)
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

	primarySeasonID, err := h.seasonService.GetPrimarySeasonID(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	accessStatus := domain.AccessStatusInactive
	for _, access := range accesses[profile.MinecraftUUID] {
		if access.SeasonID == primarySeasonID {
			accessStatus = domain.AccessStatusActive
			break
		}
		accessStatus = domain.AccessStatusExpired
	}

	seasons, err := h.seasonService.GetSeasons(ctx)
	if err != nil {
		logger.Errorf(ctx, "[SEASON] Failed to get seasons: %v", err)
		seasons = nil
	}

	isOnline, ok := onlineMap[profile.MinecraftUUID]
	if !ok {
		isOnline = false
	}

	cosmetics, err := h.cosmeticsService.GetProfilesCosmetics(ctx, uuid.UUIDs{profile.ID}, seasonID)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE COSMETICS] Failed to get profile cosmetics: %v", err)
	}

	res := presenter.PresentProfileDetails(profile, nullableMojangUUID, accessStatus, domain.SeasonAccessStatuses(seasons, accesses[profile.MinecraftUUID]), seasonsPlaytimes[profile.MinecraftUUID], isOnline, isModelSlim, presenter.PresentProfileCosmetics(cosmetics[profile.ID], utils.GetLocaleFromCtx(ctx)), seenAt(lastSeenMap, profile.MinecraftUUID))
	res.Skin = presenter.PresentPlayerSkin(skins[profile.MinecraftUUID])
	utils.WriteHttpJsonResponse(ctx, w, res)
}

// profileTimeMillis returns the time of the profile in milliseconds, nil when the profile has none.
func profileTimeMillis(times map[uuid.UUID]time.Time, profileID uuid.UUID) *int64 {
	t, ok := times[profileID]
	if !ok {
		return nil
	}
	millis := t.UnixMilli()
	return &millis
}
