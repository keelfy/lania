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
	// ResyncProfile writes the role, the chat prefix and the whitelist of the profile to the server of every
	// active season. Every part is tried on every server even when another one fails, and the report tells
	// per season which parts were written and which failed. It returns an error only when nothing could be tried.
	ResyncProfile(ctx context.Context, profileID uuid.UUID) (*domain.ProfileResync, error)
	// ResyncOwnedProfile is ResyncProfile for the owner of the profile. It fails with a forbidden error for
	// anybody else, and with a too many requests error while the cooldown of the profile runs.
	ResyncOwnedProfile(ctx context.Context, profileID, userID uuid.UUID) (*domain.ProfileResync, error)
}

type profileResyncService struct {
	profileService          ProfileService
	profileCosmeticsService ProfileCosmeticsService
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
	accessService AccessService,
	seasonService SeasonService,
	minecraftService MinecraftService,
) ProfileResyncService {
	return &profileResyncService{
		profileService:          profileService,
		profileCosmeticsService: profileCosmeticsService,
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

// shellWrite is what one shell answered for the parts that do not depend on the season.
// Seasons that share a shell address share one write.
type shellWrite struct {
	role      error
	cosmetics error
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
	prefix, prefixErr := s.profileCosmeticsService.GetProfileChatPrefix(ctx, profile)
	accessed, accessErr := s.accessedSeasons(ctx, profile)

	report := &domain.ProfileResync{}
	writes := make(map[string]shellWrite)
	for _, season := range seasons {
		// A season without shell has no server to write to.
		if !season.IsActive || season.ShellAddress == nil {
			continue
		}

		write, written := writes[*season.ShellAddress]
		if !written {
			write.role = s.minecraftService.SetPlayerRolesInSeason(ctx, season.ID, map[uuid.UUID]domain.Role{profile.MinecraftUUID: profile.Role})
			write.cosmetics = prefixErr
			if prefixErr == nil {
				write.cosmetics = s.minecraftService.SetPrefixInSeason(ctx, season.ID, profile.MinecraftUUID, prefix)
			}
			writes[*season.ShellAddress] = write
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
				{Part: domain.ResyncPartRole, Err: write.role},
				{Part: domain.ResyncPartCosmetics, Err: write.cosmetics},
				{Part: domain.ResyncPartAccess, Err: whitelistErr},
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
