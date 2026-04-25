package services

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"strings"
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
	GetPublicProfiles(ctx context.Context, pagination *domain.Pagination, sort *domain.Sort) ([]*domain.Profile, error)
	CountPublicProfiles(ctx context.Context) (int64, error)
	GetOrCreateProfileByUsername(ctx context.Context, queries sql.Queries, ownerUserID uuid.UUID, username string) (*domain.Profile, error)
	GetProfileByUsername(ctx context.Context, username string) (*domain.Profile, error)
	GetProfileByMinecraftUUID(ctx context.Context, minecraftUUID uuid.UUID) (*domain.Profile, error)
	CreateProfileByUsername(ctx context.Context, queries sql.Queries, ownerUserID uuid.UUID, username string) error
	GetProfileByID(ctx context.Context, profileID uuid.UUID) (*domain.Profile, error)
	GetProfileRole(ctx context.Context, mcUUID uuid.UUID) (domain.Role, error)
}

type profileService struct {
	storage                 storage.MainStorage
	cache                   storage.CacheStorage
	profileCosmeticsService ProfileCosmeticsService
}

func NewProfileService(
	storage storage.MainStorage,
	cache storage.CacheStorage,
	profileCosmeticsService ProfileCosmeticsService,
) ProfileService {
	return &profileService{
		storage:                 storage,
		cache:                   cache,
		profileCosmeticsService: profileCosmeticsService,
	}
}

func (s *profileService) GetProfilesByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) ([]*domain.Profile, error) {
	profiles, err := s.storage.Queries().GetProfilesByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get profiles by owner user id", err)
	}
	return profiles, nil
}

func (s *profileService) GetPublicProfiles(ctx context.Context, pagination *domain.Pagination, sort *domain.Sort) ([]*domain.Profile, error) {
	profiles, err := s.storage.Queries().FindPublicProfiles(ctx, sort.Column, sort.Direction, pagination.Size, pagination.From)
	if err == stdsql.ErrNoRows {
		return []*domain.Profile{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get public profiles", err)
	}
	return profiles, nil
}

func (s *profileService) CountPublicProfiles(ctx context.Context) (int64, error) {
	count, err := s.storage.Queries().CountPublicProfiles(ctx)
	if err == stdsql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, utils.NewInternalServerError("failed to count public profiles", err)
	}
	return count, nil
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
		return profile, nil
	}

	err = s.createProfile(ctx, queries, ownerUserID, minecraftUUID, username)
	if err != nil {
		return nil, err
	}

	return s.GetProfileByMinecraftUUID(ctx, minecraftUUID)
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
		OwnerUserID:       ownerUserID,
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

func (s *profileService) GetProfileRole(ctx context.Context, mcUUID uuid.UUID) (domain.Role, error) {
	cacheValue, err := s.cache.GetKey(ctx, fmt.Sprintf("profile_role:%s", mcUUID.String()))
	if err == nil {
		return domain.Role(cacheValue), nil
	}

	permissions, err := s.storage.Queries().FindLuckpermsPermissionLikeByMinecraftUUID(ctx, mcUUID, "group.")
	if err != nil {
		return domain.RolePlayer, err
	}

	role := domain.RolePlayer
	rolePriority := domain.RolePriorityPlayer

	for _, permission := range permissions {
		permissionParts := strings.Split(permission.Permission, ".")
		if len(permissionParts) < 2 {
			continue
		}

		rolePart := permissionParts[1]
		priority := domain.GetRolePriority(domain.Role(rolePart))
		if priority > rolePriority {
			role = domain.Role(rolePart)
			rolePriority = priority
		}
	}

	_ = s.cache.SetKey(ctx, fmt.Sprintf("profile_role:%s", mcUUID.String()), string(role), 1*time.Hour)
	return role, nil
}
