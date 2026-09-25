package services

import (
	"context"
	stdsql "database/sql"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type ProfileService interface {
	GetProfilesByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) ([]*domain.Profile, error)
	// GetPublicProfiles returns a page of profiles and the total number of profiles matching the filter.
	// Online players and the last seen sort are those of the season. Sorting by last seen over every season
	// needs uuid.Nil, and then there is no online filter.
	GetPublicProfiles(ctx context.Context, filter domain.ProfileFilter, pagination *domain.Pagination, sort *domain.Sort, seasonID uuid.UUID) ([]*domain.Profile, int64, error)
	// GetProfilesStats returns community totals. Online is those on the server of the season, and is left empty
	// when the season has no running server or the server cannot be reached.
	GetProfilesStats(ctx context.Context, seasonID uuid.UUID) (*domain.ProfilesStats, error)
	// GetProfilesLastSeenInSeason returns when every profile was last seen in the season.
	// Profiles with no known date are missing from the result.
	GetProfilesLastSeenInSeason(ctx context.Context, mcUUIDs uuid.UUIDs, seasonID uuid.UUID) (map[uuid.UUID]time.Time, error)
	// GetTopPlaytimeProfiles returns players with the most playtime in the season, each with its Profile.
	GetTopPlaytimeProfiles(ctx context.Context, seasonID uuid.UUID, limit int) ([]*domain.ProfilePlaytime, error)
	GetOrCreateProfileByUsername(ctx context.Context, queries sql.Queries, ownerUserID uuid.UUID, username string) (*domain.Profile, error)
	GetProfileByUsername(ctx context.Context, username string) (*domain.Profile, error)
	GetProfileByMinecraftUUID(ctx context.Context, minecraftUUID uuid.UUID) (*domain.Profile, error)
	CreateProfileByUsername(ctx context.Context, queries sql.Queries, ownerUserID uuid.UUID, username string) error
	GetProfileByID(ctx context.Context, profileID uuid.UUID) (*domain.Profile, error)
	// SetProfileOwner gives the profile to the user, or releases it when ownerUserID is nil.
	// The previous owner does not matter, unlike in a claim.
	SetProfileOwner(ctx context.Context, profileID uuid.UUID, ownerUserID *uuid.UUID) (*domain.Profile, error)
	// SetProfileRole stores the role in the profile, which is the source of truth for roles.
	// Role sync pushes the change to the season servers shortly after. The same role is written again too, to push it again.
	SetProfileRole(ctx context.Context, profileID uuid.UUID, role domain.Role) (*domain.Profile, error)
	// GetSeasonsPlaytimeByMinecraftUUIDs returns playtime in milliseconds summed over all seasons.
	GetSeasonsPlaytimeByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]int64, error)
	// GetProfileSeasonStats returns the stats of the profile in every season it played in, the newest season first.
	GetProfileSeasonStats(ctx context.Context, mcUUID uuid.UUID) ([]*domain.ProfileSeasonStats, error)
}

type profileService struct {
	storage                 storage.MainStorage
	profileCosmeticsService ProfileCosmeticsService
	minecraftService        MinecraftService
	mojangService           MojangService
}

func NewProfileService(
	storage storage.MainStorage,
	profileCosmeticsService ProfileCosmeticsService,
	minecraftService MinecraftService,
	mojangService MojangService,
) ProfileService {
	return &profileService{
		storage:                 storage,
		profileCosmeticsService: profileCosmeticsService,
		minecraftService:        minecraftService,
		mojangService:           mojangService,
	}
}

func (s *profileService) GetProfilesByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) ([]*domain.Profile, error) {
	profiles, err := s.storage.Queries().GetProfilesByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get profiles by owner user id", err)
	}
	return profiles, nil
}

func (s *profileService) GetPublicProfiles(ctx context.Context, filter domain.ProfileFilter, pagination *domain.Pagination, sort *domain.Sort, seasonID uuid.UUID) ([]*domain.Profile, int64, error) {
	only, err := s.filterMinecraftUUIDs(ctx, filter, seasonID)
	if err != nil {
		return nil, 0, err
	}

	profiles, err := s.storage.Queries().FindPublicProfiles(ctx, filter.Search, only, sort.Column, sort.Direction, seasonID, pagination.Size, pagination.From)
	if err == stdsql.ErrNoRows {
		profiles = []*domain.Profile{}
	} else if err != nil {
		return nil, 0, utils.NewInternalServerError("failed to get public profiles", err)
	}

	count, err := s.storage.Queries().CountPublicProfiles(ctx, filter.Search, only)
	if err == stdsql.ErrNoRows {
		count = 0
	} else if err != nil {
		return nil, 0, utils.NewInternalServerError("failed to count public profiles", err)
	}
	return profiles, count, nil
}

const newProfilesDays = 7

func (s *profileService) GetProfilesStats(ctx context.Context, seasonID uuid.UUID) (*domain.ProfilesStats, error) {
	total, err := s.storage.Queries().CountPublicProfiles(ctx, "", nil)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to count profiles", err)
	}
	recent, err := s.storage.Queries().CountRecentProfiles(ctx, newProfilesDays)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to count recent profiles", err)
	}

	stats := &domain.ProfilesStats{Total: total, NewLastWeek: recent}
	// Only players with a profile count, the same as in the online filter of the list.
	onlineUUIDs, err := s.minecraftService.ListOnlineInSeason(ctx, seasonID)
	if err != nil {
		if !errors.Is(err, ErrOnlineUnavailable) {
			logger.Warnf(ctx, "failed to list online players for stats: %v", err)
		}
		return stats, nil
	}
	online, err := s.storage.Queries().CountPublicProfiles(ctx, "", &onlineUUIDs)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to count online profiles", err)
	}
	stats.Online = &online
	return stats, nil
}

func (s *profileService) GetTopPlaytimeProfiles(ctx context.Context, seasonID uuid.UUID, limit int) ([]*domain.ProfilePlaytime, error) {
	playtimes, err := s.storage.Queries().FindTopProfilePlaytimes(ctx, seasonID, limit)
	if err == stdsql.ErrNoRows {
		return []*domain.ProfilePlaytime{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get top playtime profiles", err)
	}
	return playtimes, nil
}

// filterMinecraftUUIDs resolves filters that depend on the Minecraft server into the players they keep.
// It returns nil when no such filter is set. Online players need the server of the season, so that filter fails
// when the server cannot be reached. A season with no running server has nobody online.
func (s *profileService) filterMinecraftUUIDs(ctx context.Context, filter domain.ProfileFilter, seasonID uuid.UUID) (*uuid.UUIDs, error) {
	var only *uuid.UUIDs
	keep := func(mcUUIDs uuid.UUIDs) {
		if only == nil {
			only = &mcUUIDs
			return
		}
		kept := uuid.UUIDs{}
		for _, mcUUID := range *only {
			if slices.Contains(mcUUIDs, mcUUID) {
				kept = append(kept, mcUUID)
			}
		}
		only = &kept
	}

	if filter.OnlineOnly {
		online, err := s.minecraftService.ListOnlineInSeason(ctx, seasonID)
		if errors.Is(err, ErrOnlineUnavailable) {
			online = uuid.UUIDs{}
		} else if err != nil {
			return nil, err
		}
		keep(online)
	}
	if filter.StaffOnly {
		staff, err := s.storage.Queries().FindMinecraftUUIDsByRoles(ctx, domain.StaffRoles)
		if err != nil {
			return nil, utils.NewInternalServerError("failed to find staff profiles", err)
		}
		keep(staff)
	}
	return only, nil
}

func (s *profileService) GetProfilesLastSeenInSeason(ctx context.Context, mcUUIDs uuid.UUIDs, seasonID uuid.UUID) (map[uuid.UUID]time.Time, error) {
	lastSeen, err := s.storage.Queries().FindProfilesLastSeenInSeason(ctx, mcUUIDs, seasonID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to find profiles last seen in season", err)
	}
	return lastSeen, nil
}

func (s *profileService) GetProfileByID(ctx context.Context, profileID uuid.UUID) (*domain.Profile, error) {
	profile, err := s.storage.Queries().FindProfileByID(ctx, profileID)
	if err == stdsql.ErrNoRows {
		return nil, utils.NewNotFoundError("profile not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to find profile by id", err)
	}
	return profile, nil
}

func (s *profileService) SetProfileOwner(ctx context.Context, profileID uuid.UUID, ownerUserID *uuid.UUID) (*domain.Profile, error) {
	profile, err := s.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	err = s.storage.Queries().SetProfileOwner(ctx, profileID, ownerUserID, utils.GetUserIDFromContextOrNil(ctx))
	if err != nil {
		return nil, utils.NewInternalServerError("failed to set profile owner", err)
	}

	profile.OwnerUserID = ownerUserID
	return profile, nil
}

func (s *profileService) SetProfileRole(ctx context.Context, profileID uuid.UUID, role domain.Role) (*domain.Profile, error) {
	if !role.Valid() {
		return nil, utils.NewBadRequestError("role is invalid", nil)
	}

	profile, err := s.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	// The same role is written again on purpose: it makes role sync push the role once more.
	err = s.storage.Queries().SetProfileRole(ctx, profileID, role, utils.GetUserIDFromContextOrNil(ctx))
	if err != nil {
		return nil, utils.NewInternalServerError("failed to set profile role", err)
	}

	profile.Role = role
	return profile, nil
}

// A profile is found by its username rather than by a UUID computed from it: a premium profile has its Mojang UUID.
func (s *profileService) GetOrCreateProfileByUsername(ctx context.Context, queries sql.Queries, ownerUserID uuid.UUID, username string) (*domain.Profile, error) {
	profile, err := s.storage.Queries().FindProfileByUsername(ctx, username)
	if err != nil && err != stdsql.ErrNoRows {
		logger.Errorf(ctx, "failed to find profile by username %s: %v", username, err)
	} else if profile != nil && err == nil {
		if profile.OwnerUserID == nil {
			return s.claimProfile(ctx, queries, profile, ownerUserID)
		}
		return profile, nil
	}

	minecraftUUID, err := s.createProfile(ctx, queries, ownerUserID, username)
	if err != nil {
		return nil, err
	}

	return s.GetProfileByMinecraftUUID(ctx, minecraftUUID)
}

func (s *profileService) claimProfile(ctx context.Context, queries sql.Queries, profile *domain.Profile, ownerUserID uuid.UUID) (*domain.Profile, error) {
	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	claimed, err := queries.ClaimProfile(ctx, profile.ID, ownerUserID, authUserID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to claim profile", err)
	}
	if !claimed {
		// lost the race, somebody else claimed it first
		return s.GetProfileByMinecraftUUID(ctx, profile.MinecraftUUID)
	}

	profile.OwnerUserID = &ownerUserID
	return profile, nil
}

func (s *profileService) GetProfileByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	profile, err := s.storage.Queries().FindProfileByUsername(ctx, username)
	if err == stdsql.ErrNoRows {
		return nil, utils.NewNotFoundError("profile not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to find profile by username", err)
	}
	return profile, nil
}

func (s *profileService) GetProfileByMinecraftUUID(ctx context.Context, minecraftUUID uuid.UUID) (*domain.Profile, error) {
	profile, err := s.storage.Queries().FindProfileByMinecraftUUID(ctx, minecraftUUID)
	if err == stdsql.ErrNoRows {
		return nil, utils.NewNotFoundError("profile not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to find profile by minecraft uuid", err)
	}
	return profile, nil
}

func (s *profileService) CreateProfileByUsername(ctx context.Context, queries sql.Queries, ownerUserID uuid.UUID, username string) error {
	_, err := s.createProfile(ctx, queries, ownerUserID, username)
	return err
}

// createProfile gives the profile the UUID the server gives its nickname, and returns it.
// The Mojang lookup is stored too, so the background sync does not ask again.
func (s *profileService) createProfile(ctx context.Context, queries sql.Queries, ownerUserID uuid.UUID, username string) (uuid.UUID, error) {
	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	mcUUID, mojangUUID, err := s.mojangService.ResolveGameUUID(ctx, username)
	if err != nil {
		return uuid.Nil, err
	}

	profileID := uuid.New()
	err = queries.InsertProfile(ctx, sql.InsertProfileParams{
		ID:                profileID,
		MinecraftUUID:     mcUUID,
		MinecraftUsername: username,
		OwnerUserID:       &ownerUserID,
		Role:              string(domain.RolePlayer),
		IsSlim:            false,
		UpdatedBy:         authUserID,
	})
	if err != nil {
		return uuid.Nil, utils.NewInternalServerError("failed to insert profile", err)
	}

	err = queries.UpsertProfileMojangUUID(ctx, mcUUID, mojangUUID)
	if err != nil {
		return uuid.Nil, utils.NewInternalServerError("failed to store profile mojang uuid", err)
	}

	err = s.profileCosmeticsService.AddProfileNameColorOption(ctx, queries, profileID, config.GetDefaultNameColorID(), nil, nil)
	if err != nil {
		return uuid.Nil, utils.NewInternalServerError("failed to add profile name default color option", err)
	}
	return mcUUID, nil
}

func (s *profileService) GetSeasonsPlaytimeByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]int64, error) {
	totals, err := s.storage.Queries().SumProfilePlaytimesByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to sum profile playtimes by minecraft uuid", err)
	}
	return totals, nil
}

func (s *profileService) GetProfileSeasonStats(ctx context.Context, mcUUID uuid.UUID) ([]*domain.ProfileSeasonStats, error) {
	stats, err := s.storage.Queries().FindProfileSeasonStats(ctx, mcUUID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile season stats", err)
	}
	return stats, nil
}
