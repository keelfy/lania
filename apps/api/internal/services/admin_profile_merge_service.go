package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

// errProfileMergePreviewed forces the preview transaction to roll back: a preview must never persist anything,
// blocked or not.
var errProfileMergePreviewed = errors.New("profile merge preview")

// AdminProfileMergeService moves the site data of one game profile into another, for a player who changed
// nickname. The source profile is deleted once its data has moved. Out of scope: in-game data on the Minecraft
// servers (profiles are offline-mode UUIDs, so a new nickname is a new player there).
type AdminProfileMergeService interface {
	// PreviewMergeProfiles runs the merge inside a transaction that is always rolled back, and returns what
	// it would do, including any blocker that would stop the real merge.
	PreviewMergeProfiles(ctx context.Context, sourceProfileID, targetProfileID uuid.UUID) (*domain.ProfileMergeSummary, error)
	// MergeProfiles moves every site record of the source profile into the target profile and deletes the
	// source. It fails with a conflict when a blocker applies. The season servers are updated afterward,
	// best effort; their report never rolls the merge back.
	MergeProfiles(ctx context.Context, sourceProfileID, targetProfileID uuid.UUID) (*domain.ProfileMergeSummary, *domain.ProfileResync, error)
	// ListMergesInto returns every profile merged into the profile, newest first.
	ListMergesInto(ctx context.Context, targetProfileID uuid.UUID) ([]*domain.ProfileMerge, error)
}

type adminProfileMergeService struct {
	storage              storage.MainStorage
	seasonService        SeasonService
	minecraftService     MinecraftService
	profileResyncService ProfileResyncService
	notificationService  NotificationService
}

func NewAdminProfileMergeService(
	storage storage.MainStorage,
	seasonService SeasonService,
	minecraftService MinecraftService,
	profileResyncService ProfileResyncService,
	notificationService NotificationService,
) AdminProfileMergeService {
	return &adminProfileMergeService{
		storage:              storage,
		seasonService:        seasonService,
		minecraftService:     minecraftService,
		profileResyncService: profileResyncService,
		notificationService:  notificationService,
	}
}

// profileMergeRun is what one run of the merge logic produced, whether or not it was blocked.
type profileMergeRun struct {
	Summary *domain.ProfileMergeSummary
	// Source is the profile as it was before the merge. It stays usable after the merge deletes the row,
	// so the caller can clean the season servers of a player who no longer has a site profile.
	Source *domain.Profile
	// Target is the profile after the merge, nil when the merge was blocked.
	Target *domain.Profile
}

func (s *adminProfileMergeService) PreviewMergeProfiles(ctx context.Context, sourceProfileID, targetProfileID uuid.UUID) (*domain.ProfileMergeSummary, error) {
	var run *profileMergeRun
	err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		var runErr error
		run, runErr = s.run(ctx, queries, sourceProfileID, targetProfileID)
		if runErr != nil {
			return runErr
		}
		return errProfileMergePreviewed
	})
	if errors.Is(err, errProfileMergePreviewed) {
		return run.Summary, nil
	}
	if err != nil {
		return nil, err
	}
	return run.Summary, nil
}

func (s *adminProfileMergeService) MergeProfiles(ctx context.Context, sourceProfileID, targetProfileID uuid.UUID) (*domain.ProfileMergeSummary, *domain.ProfileResync, error) {
	var run *profileMergeRun
	err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		var runErr error
		run, runErr = s.run(ctx, queries, sourceProfileID, targetProfileID)
		if runErr != nil {
			return runErr
		}
		if !run.Summary.CanMerge() {
			return utils.NewConflictError(blockersMessage(run.Summary.Blockers), nil)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	resync := s.resyncAfterMerge(ctx, run.Source, run.Target)
	return run.Summary, resync, nil
}

// run locks both profiles, checks for blockers and, when none apply, moves every table over, deletes the source
// profile and records the merge. It never returns a Go error for a blocker: that is reported through
// Summary.Blockers, so a preview can show it without the caller mistaking it for a failed request.
func (s *adminProfileMergeService) run(ctx context.Context, queries sql.Queries, sourceProfileID, targetProfileID uuid.UUID) (*profileMergeRun, error) {
	adminID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if sourceProfileID == targetProfileID {
		profile, err := queries.FindProfileByID(ctx, sourceProfileID)
		if err != nil {
			return nil, utils.NewNotFoundError("profile not found", err)
		}
		return &profileMergeRun{Summary: &domain.ProfileMergeSummary{
			SourceProfileID: profile.ID, SourceUsername: profile.MinecraftUsername,
			TargetProfileID: profile.ID, TargetUsername: profile.MinecraftUsername,
			RoleBefore: profile.Role, RoleAfter: profile.Role, OwnerUserID: profile.OwnerUserID,
			Blockers: []*domain.ProfileMergeBlocker{{Kind: domain.ProfileMergeBlockerSameProfile}},
		}, Source: profile}, nil
	}

	if err := queries.LockProfilesForMerge(ctx, sourceProfileID, targetProfileID); err != nil {
		return nil, utils.NewInternalServerError("failed to lock profiles for merge", err)
	}

	source, err := queries.FindProfileByID(ctx, sourceProfileID)
	if err != nil {
		return nil, utils.NewNotFoundError("source profile not found", err)
	}
	target, err := queries.FindProfileByID(ctx, targetProfileID)
	if err != nil {
		return nil, utils.NewNotFoundError("target profile not found", err)
	}

	summary := &domain.ProfileMergeSummary{
		SourceProfileID: source.ID, SourceUsername: source.MinecraftUsername,
		TargetProfileID: target.ID, TargetUsername: target.MinecraftUsername,
		RoleBefore: target.Role, RoleAfter: target.Role, OwnerUserID: target.OwnerUserID,
	}

	blockers, err := s.blockers(ctx, queries, source, target)
	if err != nil {
		return nil, err
	}
	summary.Blockers = blockers
	if len(blockers) > 0 {
		return &profileMergeRun{Summary: summary, Source: source}, nil
	}

	counts, err := queries.MergeProfileData(ctx, source.ID, source.MinecraftUUID, target.ID, target.MinecraftUUID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to merge profile data", err)
	}
	summary.Counts = *counts

	ownerUserID := target.OwnerUserID
	if ownerUserID == nil {
		ownerUserID = source.OwnerUserID
	}
	firstSeenAt := earlierTime(target.FirstSeenAt, source.FirstSeenAt)
	lastSeenAt := laterTime(target.LastSeenAt, source.LastSeenAt)
	if err := queries.UpdateProfileAfterMerge(ctx, target.ID, ownerUserID, firstSeenAt, lastSeenAt, adminID); err != nil {
		return nil, utils.NewInternalServerError("failed to update target profile", err)
	}

	roleAfter := target.Role
	if domain.GetRolePriority(source.Role) < domain.GetRolePriority(target.Role) {
		roleAfter = source.Role
	}
	// Written even when the role does not change: like SetProfileRole, writing it again stamps role_updated_at
	// so role sync pushes it to the servers once more.
	if err := queries.SetProfileRole(ctx, target.ID, roleAfter, &adminID); err != nil {
		return nil, utils.NewInternalServerError("failed to set target profile role", err)
	}

	if err := queries.DeleteProfile(ctx, source.ID); err != nil {
		return nil, utils.NewInternalServerError("failed to delete source profile", err)
	}

	if err := queries.InsertProfileMerge(ctx, sql.InsertProfileMergeParams{
		SourceProfileID: source.ID, SourceMinecraftUUID: source.MinecraftUUID, SourceUsername: source.MinecraftUsername,
		TargetProfileID: target.ID, MergedBy: adminID, Summary: *counts,
	}); err != nil {
		return nil, utils.NewInternalServerError("failed to record profile merge", err)
	}

	target.OwnerUserID = ownerUserID
	target.FirstSeenAt = firstSeenAt
	target.LastSeenAt = lastSeenAt
	target.Role = roleAfter
	summary.RoleAfter = roleAfter
	summary.OwnerUserID = ownerUserID

	s.notificationService.NotifyProfileMerged(ctx, queries, target, source.MinecraftUsername)

	return &profileMergeRun{Summary: summary, Source: source, Target: target}, nil
}

func (s *adminProfileMergeService) ListMergesInto(ctx context.Context, targetProfileID uuid.UUID) ([]*domain.ProfileMerge, error) {
	merges, err := s.storage.Queries().FindProfileMergesByTargetProfileID(ctx, targetProfileID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to find profile merges", err)
	}
	return merges, nil
}

func (s *adminProfileMergeService) blockers(ctx context.Context, queries sql.Queries, source, target *domain.Profile) ([]*domain.ProfileMergeBlocker, error) {
	blockers := make([]*domain.ProfileMergeBlocker, 0)

	if source.OwnerUserID != nil && target.OwnerUserID != nil && *source.OwnerUserID != *target.OwnerUserID {
		blockers = append(blockers, &domain.ProfileMergeBlocker{Kind: domain.ProfileMergeBlockerDifferentOwners})
	}

	liveSeasons, err := queries.FindLiveSyncedSeasonNamesWithPlaytime(ctx, source.MinecraftUUID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to check live season playtime", err)
	}
	if len(liveSeasons) > 0 {
		blockers = append(blockers, &domain.ProfileMergeBlocker{Kind: domain.ProfileMergeBlockerLiveSeasonPlaytime, SeasonNames: liveSeasons})
	}

	return blockers, nil
}

// resyncAfterMerge clears the source player from every season server, then does a full resync of the target,
// best effort. It runs after the merge transaction committed, so a shell failure here never rolls the merge back.
func (s *adminProfileMergeService) resyncAfterMerge(ctx context.Context, source, target *domain.Profile) *domain.ProfileResync {
	if target == nil {
		return nil
	}

	seasons, err := s.seasonService.GetSeasons(ctx)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE MERGE] Failed to list seasons after merging %s into %s: %v", source.ID, target.ID, err)
	}
	for _, season := range seasons {
		if !season.IsActive || season.ShellAddress == nil {
			continue
		}
		if err := s.minecraftService.RemoveFromWhitelist(ctx, season.ID, source); err != nil {
			logger.Errorf(ctx, "[PROFILE MERGE] Failed to remove %s from the whitelist of season %s: %v", source.ID, season.ID, err)
		}
		if err := s.minecraftService.SetPlayerRolesInSeason(ctx, season.ID, map[uuid.UUID]domain.Role{source.MinecraftUUID: domain.RolePlayer}); err != nil {
			logger.Errorf(ctx, "[PROFILE MERGE] Failed to reset the role of %s in season %s: %v", source.ID, season.ID, err)
		}
		if err := s.minecraftService.SetPrefixInSeason(ctx, season.ID, source.MinecraftUUID, ""); err != nil {
			logger.Errorf(ctx, "[PROFILE MERGE] Failed to clear the prefix of %s in season %s: %v", source.ID, season.ID, err)
		}
	}

	resync, err := s.profileResyncService.ResyncProfile(ctx, target.ID)
	if err != nil {
		logger.Errorf(ctx, "[PROFILE MERGE] Failed to resync target profile %s: %v", target.ID, err)
		return nil
	}
	return resync
}

func blockersMessage(blockers []*domain.ProfileMergeBlocker) string {
	messages := make([]string, 0, len(blockers))
	for _, blocker := range blockers {
		switch blocker.Kind {
		case domain.ProfileMergeBlockerSameProfile:
			messages = append(messages, "cannot merge a profile into itself")
		case domain.ProfileMergeBlockerDifferentOwners:
			messages = append(messages, "profiles belong to different owners")
		case domain.ProfileMergeBlockerLiveSeasonPlaytime:
			messages = append(messages, "source profile has playtime in a live season: "+strings.Join(blocker.SeasonNames, ", "))
		default:
			messages = append(messages, string(blocker.Kind))
		}
	}
	return strings.Join(messages, "; ")
}

func earlierTime(a, b *time.Time) *time.Time {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	case b.Before(*a):
		return b
	default:
		return a
	}
}

func laterTime(a, b *time.Time) *time.Time {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	case b.After(*a):
		return b
	default:
		return a
	}
}
