package services

import (
	"context"

	"github.com/google/uuid"
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
	seasonService        SeasonService
	accessService        AccessService
	minecraftService     MinecraftService
	profileResyncService ProfileResyncService
}

func NewPremiumUUIDService(
	storage storage.MainStorage,
	seasonService SeasonService,
	accessService AccessService,
	minecraftService MinecraftService,
	profileResyncService ProfileResyncService,
) PremiumUUIDService {
	return &premiumUUIDService{
		storage:              storage,
		seasonService:        seasonService,
		accessService:        accessService,
		minecraftService:     minecraftService,
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
	seasons, err := s.seasonService.GetSeasons(ctx)
	if err != nil {
		return nil, err
	}

	report := &domain.PremiumRekeyReport{Unchecked: unchecked}
	for _, target := range targets {
		// Only the offline UUID is moved. Any other one was set by hand, and NavAuth may keep it for the player.
		offlineUUID, err := utils.GetOfflinePlayerUUID(target.MinecraftUsername)
		if err != nil || offlineUUID != target.MinecraftUUID {
			report.Skipped = append(report.Skipped, target.MinecraftUsername)
			continue
		}

		if err := s.rekey(ctx, target, seasons); err != nil {
			logger.Errorf(ctx, "[PREMIUM UUID] Failed to rekey profile %s (%s): %v", target.ProfileID, target.MinecraftUsername, err)
			report.Failed = append(report.Failed, target.MinecraftUsername)
			continue
		}
		report.Rekeyed = append(report.Rekeyed, target.MinecraftUsername)
	}
	return report, nil
}

// rekey moves the profile, then fixes what the season servers keep under the old UUID. A server that cannot be
// written to is only logged: the profile is already moved, and the admin resync writes it again.
func (s *premiumUUIDService) rekey(ctx context.Context, target *domain.PremiumRekeyTarget, seasons []*domain.Season) error {
	accesses, err := s.accessService.GetAccessesByMinecraftUUIDs(ctx, uuid.UUIDs{target.MinecraftUUID})
	if err != nil {
		return err
	}
	moved, err := s.storage.Queries().RekeyProfile(ctx, target.ProfileID, target.MinecraftUUID, target.MojangUUID)
	if err != nil {
		return err
	}
	if !moved {
		return nil
	}
	logger.Infof(ctx, "[PREMIUM UUID] Profile %s (%s) moved from %s to %s", target.ProfileID, target.MinecraftUsername, target.MinecraftUUID, target.MojangUUID)

	accessed := make(map[uuid.UUID]bool, len(accesses[target.MinecraftUUID]))
	for _, access := range accesses[target.MinecraftUUID] {
		accessed[access.SeasonID] = true
	}
	old := &domain.Profile{MinecraftUUID: target.MinecraftUUID, MinecraftUsername: target.MinecraftUsername}
	rekeyed := &domain.Profile{MinecraftUUID: target.MojangUUID, MinecraftUsername: target.MinecraftUsername}
	for _, season := range seasons {
		if !season.IsActive || season.ShellAddress == nil {
			continue
		}
		if accessed[season.ID] {
			if err := s.minecraftService.RemoveFromWhitelist(ctx, season.ID, old); err != nil {
				logger.Errorf(ctx, "[PREMIUM UUID] Failed to remove old uuid of %s from whitelist in season %s: %v", target.MinecraftUsername, season.ID, err)
			}
		}
		// LuckPerms keeps the offline UUID under the name, so its user commands would still find the old one.
		if err := s.minecraftService.RegisterPlayerInSeason(ctx, season.ID, rekeyed); err != nil {
			logger.Errorf(ctx, "[PREMIUM UUID] Failed to register %s in season %s: %v", target.MinecraftUsername, season.ID, err)
		}
	}
	// Role, prefix and the whitelist entry of the new UUID; the report is logged by the resync itself.
	if _, err := s.profileResyncService.ResyncProfile(ctx, target.ProfileID); err != nil {
		logger.Errorf(ctx, "[PREMIUM UUID] Failed to resync profile %s: %v", target.ProfileID, err)
	}
	return nil
}
