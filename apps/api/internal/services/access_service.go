package services

import (
	"context"
	stdsql "database/sql"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type AccessService interface {
	GetAccessesByMinecraftUUIDs(ctx context.Context, minecraftUUIDs uuid.UUIDs) (map[uuid.UUID][]*domain.ProfileAccess, error)
	ObtainFreeAccessForProfile(ctx context.Context, seasonID uuid.UUID, profile *domain.Profile) error
	ObtainAccessForProfile(ctx context.Context, seasonID uuid.UUID, profile *domain.Profile, source domain.AccessSource, orderItemID *uuid.UUID) error
	CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx context.Context, mcUUID uuid.UUID, seasonID uuid.UUID) (bool, error)
	GetProfileIDsWithAccessBySeasonIDAndOwnerUserID(ctx context.Context, seasonID uuid.UUID, ownerUserID uuid.UUID) (uuid.UUIDs, error)
}

type accessService struct {
	storage storage.MainStorage
}

func NewAccessService(storage storage.MainStorage) AccessService {
	return &accessService{storage: storage}
}

func (s *accessService) GetAccessesByMinecraftUUIDs(ctx context.Context, minecraftUUIDs uuid.UUIDs) (map[uuid.UUID][]*domain.ProfileAccess, error) {
	accesses, err := s.storage.Queries().FindProfileAccessesByMinecraftUUIDs(ctx, minecraftUUIDs)
	if err != nil && err != stdsql.ErrNoRows {
		return nil, utils.NewInternalServerError("failed to get accesses by minecraft uuid", err)
	}

	profileAccessesMap := make(map[uuid.UUID][]*domain.ProfileAccess)
	for _, access := range accesses {
		profileAccessesMap[access.MinecraftUUID] = append(profileAccessesMap[access.MinecraftUUID], access)
	}

	return profileAccessesMap, nil
}

func (s *accessService) CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx context.Context, mcUUID uuid.UUID, seasonID uuid.UUID) (bool, error) {
	res, err := s.storage.Queries().CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx, mcUUID, seasonID)
	if err != nil && err != stdsql.ErrNoRows {
		return false, utils.NewInternalServerError("failed to check if profile has access", err)
	}
	return res, nil
}

func (s *accessService) ObtainFreeAccessForProfile(ctx context.Context, seasonID uuid.UUID, profile *domain.Profile) error {
	return s.ObtainAccessForProfile(ctx, seasonID, profile, domain.AccessSourceFree, nil)
}

func (s *accessService) ObtainAccessForProfile(ctx context.Context, seasonID uuid.UUID, profile *domain.Profile, source domain.AccessSource, orderItemID *uuid.UUID) error {
	authUserID := utils.GetUserIDFromContextOrNil(ctx)

	err := s.storage.Queries().InsertProfileAccess(ctx, sql.InsertProfileAccessParams{
		MinecraftUUID: profile.MinecraftUUID,
		SeasonID:      seasonID,
		Source:        string(source),
		OrderItemID:   orderItemID,
		UpdatedBy:     authUserID,
	})
	if err != nil {
		return utils.NewInternalServerError("failed to insert profile access", err)
	}
	return nil
}

func (s *accessService) GetProfileIDsWithAccessBySeasonIDAndOwnerUserID(ctx context.Context, seasonID uuid.UUID, ownerUserID uuid.UUID) (uuid.UUIDs, error) {
	profileIDs, err := s.storage.Queries().GetProfileAccessesBySeasonIDAndOwnerUserID(ctx, seasonID, ownerUserID)
	if err == stdsql.ErrNoRows {
		return uuid.UUIDs{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile accesses by season id and owner user id", err)
	}
	return profileIDs, nil
}
