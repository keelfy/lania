package services

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/utils"
)

// ownerResyncCooldown is the time an owner waits between two resyncs of the same profile.
// Every resync calls every active season server, so it is not free.
const ownerResyncCooldown = time.Minute

// ProfileResyncService writes everything the Minecraft servers keep about a profile again, outside the timers.
type ProfileResyncService interface {
	// ResyncProfile writes the role, the chat prefix, the whitelist and the privileges of the profile to the server of every
	// active season. Every part is tried on every server even when another one fails, and the report tells
	// per season which parts were written and which failed. It returns an error only when nothing could be tried.
	ResyncProfile(ctx context.Context, profileID uuid.UUID) (*domain.ProfileResync, error)
	// ResyncOwnedProfile is ResyncProfile for the owner of the profile. It fails with a forbidden error for
	// anybody else, and with a too many requests error while the cooldown of the profile runs.
	ResyncOwnedProfile(ctx context.Context, profileID, userID uuid.UUID) (*domain.ProfileResync, error)
	// MoveProfileOnServers makes the season servers follow a profile that got another UUID or nickname. old is the
	// profile before the move. A UUID the profile left loses its whitelist entry, role, prefix and privileges, LuckPerms learns
	// the current nickname, then the profile is resynced. It runs after the move is stored, so failures are only
	// logged: the profile is already moved, and a resync writes it again. The report is nil when nothing was tried.
	MoveProfileOnServers(ctx context.Context, profileID uuid.UUID, old *domain.Profile) *domain.ProfileResync
}

type profileResyncService struct {
	profileService          ProfileService
	profileCosmeticsService ProfileCosmeticsService
	profilePrivilegeService ProfilePrivilegeService
	accessService           AccessService
	seasonService           SeasonService
	minecraftService        MinecraftService

	now func() time.Time
	// lastOwnerResync holds the time of the last resync an owner asked for, per profile.
	// It lives in memory, so every API instance keeps its own cooldown.
	mu              sync.Mutex
	lastOwnerResync map[uuid.UUID]time.Time
}

func NewProfileResyncService(
	profileService ProfileService,
	profileCosmeticsService ProfileCosmeticsService,
	profilePrivilegeService ProfilePrivilegeService,
	accessService AccessService,
	seasonService SeasonService,
	minecraftService MinecraftService,
) ProfileResyncService {
	return &profileResyncService{
		profileService:          profileService,
		profileCosmeticsService: profileCosmeticsService,
		profilePrivilegeService: profilePrivilegeService,
		accessService:           accessService,
		seasonService:           seasonService,
		minecraftService:        minecraftService,
		now:                     time.Now,
		lastOwnerResync:         make(map[uuid.UUID]time.Time),
	}
}

func (s *profileResyncService) ResyncOwnedProfile(ctx context.Context, profileID, userID uuid.UUID) (*domain.ProfileResync, error) {
	profile, err := s.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if profile.OwnerUserID == nil || *profile.OwnerUserID != userID {
		return nil, utils.NewForbiddenError("only owner can resync the profile", nil)
	}
	if !s.startOwnerResync(profileID) {
		return nil, utils.NewTooManyRequestsError("the profile was resynced a moment ago, try again later", nil)
	}
	return s.ResyncProfile(ctx, profileID)
}

// startOwnerResync tells whether the cooldown of the profile is over, and starts it again when it is.
func (s *profileResyncService) startOwnerResync(profileID uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	for id, last := range s.lastOwnerResync {
		if now.Sub(last) >= ownerResyncCooldown {
			delete(s.lastOwnerResync, id)
		}
	}
	if _, waiting := s.lastOwnerResync[profileID]; waiting {
		return false
	}
	s.lastOwnerResync[profileID] = now
	return true
}

func (s *profileResyncService) ResyncProfile(ctx context.Context, profileID uuid.UUID) (*domain.ProfileResync, error) {
	profile, err := s.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	seasons, err := s.seasonService.GetSeasons(ctx)
	if err != nil {
		return nil, err
	}

	// What cannot be read is a failure of its part on every server, not a reason to skip the servers.
	accessed, accessErr := s.accessedSeasons(ctx, profile)

	report := &domain.ProfileResync{}
	for _, season := range seasons {
		// A season without shell has no server to write to.
		if !season.IsActive || season.ShellAddress == nil {
			continue
		}

		roleErr := s.minecraftService.SetPlayerRolesInSeason(ctx, season.ID, map[uuid.UUID]domain.Role{profile.MinecraftUUID: profile.Role})
		// Every season keeps its own selection, so the prefix is built for each one.
		prefix, cosmeticsErr := s.profileCosmeticsService.GetProfileChatPrefix(ctx, profile.ID, season.ID)
		if cosmeticsErr == nil {
			cosmeticsErr = s.minecraftService.SetPrefixInSeason(ctx, season.ID, profile.MinecraftUUID, prefix)
		}

		// Match the whitelist to the access, so a player whose access was revoked while
		// the server could not be reached is removed too.
		whitelistErr := accessErr
		if accessErr == nil && accessed[season.ID] {
			whitelistErr = s.minecraftService.AddToWhitelist(ctx, season.ID, profile)
		} else if accessErr == nil {
			whitelistErr = s.minecraftService.RemoveFromWhitelist(ctx, season.ID, profile)
		}

		result := &domain.SeasonResync{
			SeasonID:   season.ID,
			SeasonName: season.Name,
			Parts: []domain.PartResync{
				{Part: domain.ResyncPartRole, Err: roleErr},
				{Part: domain.ResyncPartCosmetics, Err: cosmeticsErr},
				{Part: domain.ResyncPartAccess, Err: whitelistErr},
				{Part: domain.ResyncPartPrivileges, Err: s.profilePrivilegeService.SyncProfileInSeason(ctx, profile, season.ID)},
			},
		}
		for _, part := range result.Parts {
			if part.Err != nil {
				logger.Errorf(ctx, "[PROFILE RESYNC] Failed to resync %s of profile %s in season %s: %v", part.Part, profile.ID, season.ID, part.Err)
			}
		}
		report.Seasons = append(report.Seasons, result)
	}
	return report, nil
}

func (s *profileResyncService) MoveProfileOnServers(ctx context.Context, profileID uuid.UUID, old *domain.Profile) *domain.ProfileResync {
	profile, err := s.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE RESYNC] Failed to get moved profile %s: %v", profileID, err)
		return nil
	}
	seasons, err := s.seasonService.GetSeasons(ctx)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE RESYNC] Failed to list seasons to move profile %s: %v", profileID, err)
		return nil
	}

	for _, season := range seasons {
		if !season.IsActive || season.ShellAddress == nil {
			continue
		}
		if old.MinecraftUUID != profile.MinecraftUUID {
			if err := s.minecraftService.RemoveFromWhitelist(ctx, season.ID, old); err != nil {
				logger.Errorf(ctx, "[PROFILE RESYNC] Failed to remove old uuid %s of profile %s from the whitelist of season %s: %v", old.MinecraftUUID, profileID, season.ID, err)
			}
			if err := s.minecraftService.SetPlayerRolesInSeason(ctx, season.ID, map[uuid.UUID]domain.Role{old.MinecraftUUID: domain.RolePlayer}); err != nil {
				logger.Errorf(ctx, "[PROFILE RESYNC] Failed to reset the role of old uuid %s of profile %s in season %s: %v", old.MinecraftUUID, profileID, season.ID, err)
			}
			if err := s.profilePrivilegeService.ClearPlayerInSeason(ctx, old.MinecraftUUID, season.ID); err != nil {
				logger.Errorf(ctx, "[PROFILE RESYNC] Failed to clear the privileges of old uuid %s of profile %s in season %s: %v", old.MinecraftUUID, profileID, season.ID, err)
			}
			if err := s.minecraftService.SetPrefixInSeason(ctx, season.ID, old.MinecraftUUID, ""); err != nil {
				logger.Errorf(ctx, "[PROFILE RESYNC] Failed to clear the prefix of old uuid %s of profile %s in season %s: %v", old.MinecraftUUID, profileID, season.ID, err)
			}
		}
		// LuckPerms keeps the old UUID under the name, so its user commands would still find the old one.
		if err := s.minecraftService.RegisterPlayerInSeason(ctx, season.ID, profile); err != nil {
			logger.Errorf(ctx, "[PROFILE RESYNC] Failed to register %s in season %s: %v", profile.MinecraftUsername, season.ID, err)
		}
	}

	// Role, prefix, privileges and the whitelist entry of the current UUID; the report is logged by the resync itself.
	resync, err := s.ResyncProfile(ctx, profileID)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE RESYNC] Failed to resync moved profile %s: %v", profileID, err)
		return nil
	}
	return resync
}

// accessedSeasons returns the seasons the profile has access to.
func (s *profileResyncService) accessedSeasons(ctx context.Context, profile *domain.Profile) (map[uuid.UUID]bool, error) {
	accesses, err := s.accessService.GetAccessesByMinecraftUUIDs(ctx, uuid.UUIDs{profile.MinecraftUUID})
	if err != nil {
		return nil, err
	}
	accessed := make(map[uuid.UUID]bool, len(accesses[profile.MinecraftUUID]))
	for _, access := range accesses[profile.MinecraftUUID] {
		accessed[access.SeasonID] = true
	}
	return accessed, nil
}
