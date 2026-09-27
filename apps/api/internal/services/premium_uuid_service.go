package services

import (
	"context"

	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils"
)

// PremiumUUIDService moves profiles of licensed nicknames to their Mojang UUID. NavAuth on the proxy gives such
// a player the Mojang UUID in game, while profiles made before it have the offline one.
type PremiumUUIDService interface {
	// RekeyPremiumProfiles moves every profile still at the offline UUID of a licensed nickname to the Mojang UUID,
	// then writes the LuckPerms name, role, prefix and whitelist of the new UUID to the season servers.
	// Premium conflicts are left to an admin. It is safe to repeat.
	RekeyPremiumProfiles(ctx context.Context) (*domain.PremiumRekeyReport, error)
}

type premiumUUIDService struct {
	storage              storage.MainStorage
	profileResyncService ProfileResyncService
}

func NewPremiumUUIDService(storage storage.MainStorage, profileResyncService ProfileResyncService) PremiumUUIDService {
	return &premiumUUIDService{
		storage:              storage,
		profileResyncService: profileResyncService,
	}
}

func (s *premiumUUIDService) RekeyPremiumProfiles(ctx context.Context) (*domain.PremiumRekeyReport, error) {
	targets, err := s.storage.Queries().FindPremiumRekeyTargets(ctx)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to find profiles to rekey", err)
	}
	unchecked, err := s.storage.Queries().CountUncheckedMojangProfiles(ctx)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to count unchecked profiles", err)
	}

	report := &domain.PremiumRekeyReport{Unchecked: unchecked}
	for _, target := range targets {
		// Only the offline UUID is moved. Any other one was set by hand, and NavAuth may keep it for the player.
		offlineUUID, err := utils.GetOfflinePlayerUUID(target.MinecraftUsername)
		if err != nil || offlineUUID != target.MinecraftUUID {
			report.Skipped = append(report.Skipped, target.MinecraftUsername)
			continue
		}

		if err := s.rekey(ctx, target); err != nil {
			logger.Errorf(ctx, "[PREMIUM UUID] Failed to rekey profile %s (%s): %v", target.ProfileID, target.MinecraftUsername, err)
			report.Failed = append(report.Failed, target.MinecraftUsername)
			continue
		}
		report.Rekeyed = append(report.Rekeyed, target.MinecraftUsername)
	}
	return report, nil
}

// rekey moves the profile, then fixes what the season servers keep under the old UUID, best effort: the profile
// is already moved, and the admin resync writes it again.
func (s *premiumUUIDService) rekey(ctx context.Context, target *domain.PremiumRekeyTarget) error {
	moved, err := s.storage.Queries().RekeyProfile(ctx, target.ProfileID, target.MinecraftUUID, target.MojangUUID)
	if err != nil {
		return err
	}
	if !moved {
		return nil
	}
	logger.Infof(ctx, "[PREMIUM UUID] Profile %s (%s) moved from %s to %s", target.ProfileID, target.MinecraftUsername, target.MinecraftUUID, target.MojangUUID)

	old := &domain.Profile{ID: target.ProfileID, MinecraftUUID: target.MinecraftUUID, MinecraftUsername: target.MinecraftUsername}
	s.profileResyncService.MoveProfileOnServers(ctx, target.ProfileID, old)
	return nil
}
