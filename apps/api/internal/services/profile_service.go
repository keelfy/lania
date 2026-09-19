package services

import (
	"context"
	stdsql "database/sql"
	"fmt"
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
	GetPublicProfiles(ctx context.Context, filter domain.ProfileFilter, pagination *domain.Pagination, sort *domain.Sort) ([]*domain.Profile, int64, error)
	GetOrCreateProfileByUsername(ctx context.Context, queries sql.Queries, ownerUserID uuid.UUID, username string) (*domain.Profile, error)
	GetProfileByUsername(ctx context.Context, username string) (*domain.Profile, error)
	GetProfileByMinecraftUUID(ctx context.Context, minecraftUUID uuid.UUID) (*domain.Profile, error)
	CreateProfileByUsername(ctx context.Context, queries sql.Queries, ownerUserID uuid.UUID, username string) error
	GetProfileByID(ctx context.Context, profileID uuid.UUID) (*domain.Profile, error)
	GetProfileRole(ctx context.Context, mcUUID uuid.UUID) (domain.Role, error)
	// GetSeasonsPlaytimeByMinecraftUUIDs returns playtime in milliseconds summed over all seasons.
	GetSeasonsPlaytimeByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]int64, error)
}

type profileService struct {
	storage                 storage.MainStorage
	cache                   storage.CacheStorage
	profileCosmeticsService ProfileCosmeticsService
	minecraftService        MinecraftService
}

func NewProfileService(
	storage storage.MainStorage,
	cache storage.CacheStorage,
	profileCosmeticsService ProfileCosmeticsService,
	minecraftService MinecraftService,
) ProfileService {
	return &profileService{
		storage:                 storage,
		cache:                   cache,
		profileCosmeticsService: profileCosmeticsService,
		minecraftService:        minecraftService,
	}
}

func (s *profileService) GetProfilesByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) ([]*domain.Profile, error) {
	profiles, err := s.storage.Queries().GetProfilesByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get profiles by owner user id", err)
	}
	return profiles, nil
}

func (s *profileService) GetPublicProfiles(ctx context.Context, filter domain.ProfileFilter, pagination *domain.Pagination, sort *domain.Sort) ([]*domain.Profile, int64, error) {
	only, err := s.filterMinecraftUUIDs(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	profiles, err := s.storage.Queries().FindPublicProfiles(ctx, filter.Search, only, sort.Column, sort.Direction, pagination.Size, pagination.From)
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

// filterMinecraftUUIDs resolves filters that depend on the Minecraft server into the players they keep.
// It returns nil when no such filter is set. Failing is intended: without the server the filter cannot be applied.
func (s *profileService) filterMinecraftUUIDs(ctx context.Context, filter domain.ProfileFilter) (*uuid.UUIDs, error) {
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
		online, err := s.minecraftService.ListOnlineMinecraftUUIDs(ctx)
		if err != nil {
			return nil, err
		}
		keep(online)
	}
	if filter.StaffOnly {
		groups := make([]string, len(domain.StaffRoles))
		for i, role := range domain.StaffRoles {
			groups[i] = string(role)
		}
		staff, err := s.minecraftService.ListMinecraftUUIDsByGroups(ctx, groups)
		if err != nil {
			return nil, err
		}
		keep(staff)
	}
	return only, nil
}

func (s *profileService) GetProfileByID(ctx context.Context, profileID uuid.UUID) (*domain.Profile, error) {
	profile, err := s.storage.Queries().FindProfileByID(ctx, profileID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to find profile by id", err)
	} else if err == stdsql.ErrNoRows {
		return nil, utils.NewNotFoundError("profile not found", err)
	}
	return profile, nil
}

func (s *profileService) GetOrCreateProfileByUsername(ctx context.Context, queries sql.Queries, ownerUserID uuid.UUID, username string) (*domain.Profile, error) {
	minecraftUUID, err := utils.GetOfflinePlayerUUID(username)
	if err != nil {
		return nil, err
	}

	profile, err := s.storage.Queries().FindProfileByMinecraftUUID(ctx, minecraftUUID)
	if err != nil && err != stdsql.ErrNoRows {
		logger.Errorf(ctx, "failed to find profile by username %s: %v", username, err)
	} else if profile != nil && err == nil {
		if profile.OwnerUserID == nil {
			return s.claimProfile(ctx, queries, profile, ownerUserID)
		}
		return profile, nil
	}

	err = s.createProfile(ctx, queries, ownerUserID, minecraftUUID, username)
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
	minecraftUUID, err := utils.GetOfflinePlayerUUID(username)
	if err != nil {
		return nil, err
	}

	return s.GetProfileByMinecraftUUID(ctx, minecraftUUID)
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
	minecraftUUID, err := utils.GetOfflinePlayerUUID(username)
	if err != nil {
		return err
	}

	return s.createProfile(ctx, queries, ownerUserID, minecraftUUID, username)
}

func (s *profileService) createProfile(ctx context.Context, queries sql.Queries, ownerUserID, mcUUID uuid.UUID, username string) error {
	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		return err
	}

	profileID := uuid.New()
	err = queries.InsertProfile(ctx, sql.InsertProfileParams{
		ID:                profileID,
		MinecraftUUID:     mcUUID,
		MinecraftUsername: username,
		OwnerUserID:       &ownerUserID,
		Role:              string(domain.RolePlayer),
		IsSlim:            false,
		NameColorID:       config.GetDefaultNameColorID(),
		UpdatedBy:         authUserID,
	})
	if err != nil {
		return utils.NewInternalServerError("failed to insert profile", err)
	}

	err = s.profileCosmeticsService.AddProfileNameColorOption(ctx, queries, profileID, config.GetDefaultNameColorID(), nil, nil)
	if err != nil {
		return utils.NewInternalServerError("failed to add profile name default color option", err)
	}
	return nil
}

func (s *profileService) GetSeasonsPlaytimeByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]int64, error) {
	totals, err := s.storage.Queries().SumProfilePlaytimesByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to sum profile playtimes by minecraft uuid", err)
	}
	return totals, nil
}

func (s *profileService) GetProfileRole(ctx context.Context, mcUUID uuid.UUID) (domain.Role, error) {
	cacheValue, err := s.cache.GetKey(ctx, fmt.Sprintf("profile_role:%s", mcUUID.String()))
	if err == nil {
		return domain.Role(cacheValue), nil
	}

	groups, err := s.minecraftService.GetGroupsByMinecraftUUIDs(ctx, uuid.UUIDs{mcUUID})
	if err != nil {
		return domain.RolePlayer, err
	}

	role := domain.HighestRole(groups[mcUUID])

	_ = s.cache.SetKey(ctx, fmt.Sprintf("profile_role:%s", mcUUID.String()), string(role), 1*time.Hour)
	return role, nil
}
