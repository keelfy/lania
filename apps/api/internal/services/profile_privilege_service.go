package services

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

// ProfilePrivilegeService keeps the privileges of profiles, and the permission nodes they stand for on the season servers.
type ProfilePrivilegeService interface {
	// AddProfilePrivilege gives the privilege to the profile for the season and writes its node to the season server.
	// The database is the source of truth, so a server that cannot be reached is only logged:
	// a resync writes the node later. orderItemID is nil when the privilege is not from an order.
	// The admin of the request, if any, is recorded as the one who gave it.
	AddProfilePrivilege(ctx context.Context, queries sql.Queries, profile *domain.Profile, privilegeID, seasonID uuid.UUID, orderItemID *uuid.UUID) error
	GetProfilePrivileges(ctx context.Context, profileID, seasonID uuid.UUID) ([]*domain.ProfilePrivilege, error)
	GetProfilePrivilegesByOwnerUserID(ctx context.Context, ownerUserID, seasonID uuid.UUID) ([]*domain.ProfilePrivilege, error)
	// SyncProfileInSeason makes the server of the season match the database: the node of every privilege the
	// profile has is written, and the node of every other privilege of the catalog is removed.
	SyncProfileInSeason(ctx context.Context, profile *domain.Profile, seasonID uuid.UUID) error
	// ClearPlayerInSeason removes the node of every privilege of the catalog from the player, who is the
	// minecraftUUID a profile left, so the old UUID keeps nothing the profile bought.
	ClearPlayerInSeason(ctx context.Context, mcUUID, seasonID uuid.UUID) error
}

type profilePrivilegeService struct {
	storage          storage.MainStorage
	minecraftService MinecraftService
}

func NewProfilePrivilegeService(storage storage.MainStorage, minecraftService MinecraftService) ProfilePrivilegeService {
	return &profilePrivilegeService{storage: storage, minecraftService: minecraftService}
}

func (s *profilePrivilegeService) AddProfilePrivilege(ctx context.Context, queries sql.Queries, profile *domain.Profile, privilegeID, seasonID uuid.UUID, orderItemID *uuid.UUID) error {
	privilege, err := queries.FindPrivilegeByID(ctx, privilegeID)
	if err != nil {
		return utils.NewInternalServerError("failed to find privilege", err)
	}
	err = queries.InsertProfilePrivilege(ctx, sql.InsertProfilePrivilegeParams{
		ProfileID:   profile.ID,
		PrivilegeID: privilegeID,
		SeasonID:    seasonID,
		OrderItemID: orderItemID,
		CreatedBy:   utils.GetUserIDFromContextOrNil(ctx),
	})
	if err != nil {
		return utils.NewInternalServerError("failed to add profile privilege", err)
	}

	if err := s.minecraftService.SetPrivilegesInSeason(ctx, seasonID, profile.MinecraftUUID, []string{privilege.Permission}, nil); err != nil {
		logger.Errorf(ctx, "[PRIVILEGE] Failed to write %s of profile %s in season %s, a resync writes it later: %v", privilege.Permission, profile.ID, seasonID, err)
	}
	return nil
}

func (s *profilePrivilegeService) GetProfilePrivileges(ctx context.Context, profileID, seasonID uuid.UUID) ([]*domain.ProfilePrivilege, error) {
	privileges, err := s.storage.Queries().FindProfilePrivileges(ctx, profileID, seasonID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to find profile privileges", err)
	}
	return privileges, nil
}

func (s *profilePrivilegeService) GetProfilePrivilegesByOwnerUserID(ctx context.Context, ownerUserID, seasonID uuid.UUID) ([]*domain.ProfilePrivilege, error) {
	privileges, err := s.storage.Queries().FindProfilePrivilegesByOwnerUserID(ctx, ownerUserID, seasonID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to find privileges by owner user id", err)
	}
	return privileges, nil
}

func (s *profilePrivilegeService) SyncProfileInSeason(ctx context.Context, profile *domain.Profile, seasonID uuid.UUID) error {
	catalog, err := s.storage.Queries().FindPrivileges(ctx)
	if err != nil {
		return utils.NewInternalServerError("failed to find privileges", err)
	}
	owned, err := s.GetProfilePrivileges(ctx, profile.ID, seasonID)
	if err != nil {
		return err
	}

	add := make([]string, 0, len(owned))
	for _, privilege := range owned {
		add = append(add, privilege.Permission)
	}
	remove := make([]string, 0, len(catalog))
	for _, privilege := range catalog {
		if !slices.Contains(add, privilege.Permission) {
			remove = append(remove, privilege.Permission)
		}
	}
	if len(add) == 0 && len(remove) == 0 {
		return nil
	}
	return s.minecraftService.SetPrivilegesInSeason(ctx, seasonID, profile.MinecraftUUID, add, remove)
}

func (s *profilePrivilegeService) ClearPlayerInSeason(ctx context.Context, mcUUID, seasonID uuid.UUID) error {
	catalog, err := s.storage.Queries().FindPrivileges(ctx)
	if err != nil {
		return utils.NewInternalServerError("failed to find privileges", err)
	}
	if len(catalog) == 0 {
		return nil
	}
	remove := make([]string, len(catalog))
	for i, privilege := range catalog {
		remove[i] = privilege.Permission
	}
	return s.minecraftService.SetPrivilegesInSeason(ctx, seasonID, mcUUID, nil, remove)
}
