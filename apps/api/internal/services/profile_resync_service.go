package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	// ResyncProfile writes the role, the chat prefix and the whitelists of the profile to the active season servers.
	// Every part is tried even when another one fails, and the error names the parts that failed.
	ResyncProfile(ctx context.Context, profileID uuid.UUID) error
	// ResyncOwnedProfile is ResyncProfile for the owner of the profile. It fails with a forbidden error for
	// anybody else, and with a too many requests error while the cooldown of the profile runs.
	ResyncOwnedProfile(ctx context.Context, profileID, userID uuid.UUID) error
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

func (s *profileResyncService) ResyncOwnedProfile(ctx context.Context, profileID, userID uuid.UUID) error {
	profile, err := s.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		return err
	}
	if profile.OwnerUserID == nil || *profile.OwnerUserID != userID {
		return utils.NewForbiddenError("only owner can resync the profile", nil)
	}
	if !s.startOwnerResync(profileID) {
		return utils.NewTooManyRequestsError("the profile was resynced a moment ago, try again later", nil)
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

func (s *profileResyncService) ResyncProfile(ctx context.Context, profileID uuid.UUID) error {
	profile, err := s.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		return err
	}

	var failedParts []string
	var failures []error
	fail := func(part string, err error) {
		logger.Errorf(ctx, "[PROFILE RESYNC] Failed to resync %s of profile %s: %v", part, profile.ID, err)
		failedParts = append(failedParts, part)
		failures = append(failures, fmt.Errorf("%s: %w", part, err))
	}

	if err := s.resyncRole(ctx, profile); err != nil {
		fail("role", err)
	}
	if err := s.resyncPrefix(ctx, profile); err != nil {
		fail("cosmetics", err)
	}
	if err := s.resyncWhitelists(ctx, profile); err != nil {
		fail("access", err)
	}

	if len(failures) > 0 {
		return utils.NewInternalServerError("failed to resync "+strings.Join(failedParts, ", "), errors.Join(failures...))
	}
	return nil
}

func (s *profileResyncService) resyncRole(ctx context.Context, profile *domain.Profile) error {
	return s.minecraftService.SetPlayerRoles(ctx, map[uuid.UUID]domain.Role{profile.MinecraftUUID: profile.Role})
}

func (s *profileResyncService) resyncPrefix(ctx context.Context, profile *domain.Profile) error {
	prefix, err := s.profileCosmeticsService.GetProfileChatPrefix(ctx, profile)
	if err != nil {
		return err
	}
	return s.minecraftService.SetPrefixByMinecraftUUID(ctx, profile.MinecraftUUID, prefix)
}

// resyncWhitelists makes the whitelist of every active season match the access of the profile to it,
// so it also removes a player whose access was revoked while the server could not be reached.
func (s *profileResyncService) resyncWhitelists(ctx context.Context, profile *domain.Profile) error {
	seasons, err := s.seasonService.GetSeasons(ctx)
	if err != nil {
		return err
	}
	accesses, err := s.accessService.GetAccessesByMinecraftUUIDs(ctx, uuid.UUIDs{profile.MinecraftUUID})
	if err != nil {
		return err
	}
	accessed := make(map[uuid.UUID]bool, len(accesses[profile.MinecraftUUID]))
	for _, access := range accesses[profile.MinecraftUUID] {
		accessed[access.SeasonID] = true
	}

	var failures []error
	for _, season := range seasons {
		// A season without shell has no server to write to.
		if !season.IsActive || season.ShellAddress == nil {
			continue
		}

		if accessed[season.ID] {
			err = s.minecraftService.AddToWhitelist(ctx, season.ID, profile)
		} else {
			err = s.minecraftService.RemoveFromWhitelist(ctx, season.ID, profile)
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("season %s: %w", season.ID, err))
		}
	}
	return errors.Join(failures...)
}
