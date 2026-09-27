package services

import (
	"context"
	stdsql "database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

const (
	premiumNameSyncInterval = time.Minute
	// premiumNameSyncBatchSize names are asked one by one, Mojang has no bulk lookup by UUID.
	premiumNameSyncBatchSize = 10
	premiumNameLookupDelay   = time.Second
	// premiumNameCheckInterval is how often the name of every licensed profile is compared with Mojang's.
	premiumNameCheckInterval = 24 * time.Hour
)

// ProfileRenameService changes the nickname of a profile. An unlicensed profile gets the offline UUID of the new
// nickname: the site moves everything it keeps, while the progress in game stays under the old UUID. A licensed
// profile keeps its Mojang UUID, and the site follows the name the account has on Mojang.
//
// A UUID the profile leaves stays reserved for it: the player sync goes on summing the playtime Plan keeps under it,
// and nobody else can take the nickname back.
type ProfileRenameService interface {
	// ChangeOwnedUsername renames the profile for its owner. A licensed profile takes only the current name of its
	// Mojang account. An unlicensed one takes a nickname nobody holds and no Mojang account has, once per cooldown.
	// unlicensed makes a licensed profile take such a nickname as well, for an owner who does not own the account:
	// the Mojang UUID is left free for its real owner. A verified licensed profile cannot do it.
	ChangeOwnedUsername(ctx context.Context, userID, profileID uuid.UUID, username string, unlicensed bool) (*domain.Profile, error)
	// GetUsernameChangeAvailableAt returns when the owner can change the nickname again, for the profiles still in
	// the cooldown after the owner's last change. Other profiles are absent.
	GetUsernameChangeAvailableAt(ctx context.Context, profileIDs uuid.UUIDs) (map[uuid.UUID]time.Time, error)
	// GetUsernameChanges returns every nickname change of the profile, newest first.
	GetUsernameChanges(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileUsernameChange, error)
	// RunPremiumNameSync picks up new names of licensed accounts from Mojang until ctx is done.
	RunPremiumNameSync(ctx context.Context)
}

type profileRenameService struct {
	storage              storage.MainStorage
	mojangService        MojangService
	profileResyncService ProfileResyncService

	cooldown time.Duration
	now      func() time.Time
}

func NewProfileRenameService(
	storage storage.MainStorage,
	mojangService MojangService,
	profileResyncService ProfileResyncService,
) ProfileRenameService {
	return &profileRenameService{
		storage:              storage,
		mojangService:        mojangService,
		profileResyncService: profileResyncService,
		cooldown:             config.GetNicknameChangeCooldown(),
		now:                  time.Now,
	}
}

// profileRename is one nickname change of a profile.
type profileRename struct {
	// before is the profile as the change was decided on; the change fails if it moved meanwhile.
	before    *domain.Profile
	username  string
	newMcUUID uuid.UUID
	source    domain.ProfileUsernameChangeSource
	changedBy *uuid.UUID
	// leavesLicense is a licensed profile moving to an unlicensed nickname. The Mojang UUID it leaves is not its
	// own, so it is not kept as a former UUID: no playtime under it counts, and its owner can take the nickname.
	leavesLicense bool
}

// licensed tells whether the profile keeps its UUID, which only a licensed account does.
func (r *profileRename) licensed() bool {
	return r.newMcUUID == r.before.MinecraftUUID
}

func (s *profileRenameService) ChangeOwnedUsername(ctx context.Context, userID, profileID uuid.UUID, username string, unlicensed bool) (*domain.Profile, error) {
	username = strings.TrimSpace(username)
	if !mojangUsernameRegexp.MatchString(username) {
		return nil, utils.NewBadRequestError("a nickname is 3 to 16 latin letters, digits or underscores", nil)
	}

	profile, err := s.storage.Queries().FindProfileByID(ctx, profileID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return nil, utils.NewNotFoundError("profile not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get profile", err)
	}
	if profile.OwnerUserID == nil || *profile.OwnerUserID != userID {
		return nil, utils.NewForbiddenError("only owner can change the nickname of the profile", nil)
	}
	if username == profile.MinecraftUsername {
		return nil, utils.NewBadRequestError("the profile already has this nickname", nil)
	}

	mojangUUIDs, err := s.storage.Queries().FindProfileMojangUUIDsByMinecraftUUIDs(ctx, uuid.UUIDs{profile.MinecraftUUID})
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get the mojang uuid of the profile", err)
	}
	rename := &profileRename{before: profile, source: domain.ProfileUsernameChangeSourceOwner, changedBy: &userID}
	licensed := mojangUUIDs[profile.MinecraftUUID] == profile.MinecraftUUID
	// The owner proved in game the account is theirs, so they cannot say it is not.
	if licensed && unlicensed && profile.VerifiedAt != nil {
		return nil, utils.NewConflictError("the licensed account of the profile is verified as the owner's", nil)
	}

	if licensed && !unlicensed {
		// The name comes from Mojang, in the case Mojang has it; the owner only confirms it.
		name, err := s.mojangService.LookupUsernameByMojangUUID(ctx, profile.MinecraftUUID)
		if err != nil {
			return nil, utils.NewServiceUnavailableError("failed to check the nickname on Mojang, try again later", err)
		}
		if !strings.EqualFold(name, username) {
			return nil, utils.NewConflictError("the nickname is not the current name of the Minecraft account of the profile", nil)
		}
		if name == profile.MinecraftUsername {
			return nil, utils.NewBadRequestError("the profile already has this nickname", nil)
		}
		rename.username, rename.newMcUUID = name, profile.MinecraftUUID
		return s.rename(ctx, rename)
	}

	gameUUID, mojangUUID, err := s.mojangService.ResolveGameUUID(ctx, username)
	if err != nil {
		return nil, err
	}
	// Moving to a licensed nickname means moving to another player in game: an admin merges such profiles.
	if mojangUUID != nil {
		return nil, utils.NewConflictError("the nickname belongs to a licensed Minecraft account", nil)
	}
	rename.username, rename.newMcUUID, rename.leavesLicense = username, gameUUID, licensed
	return s.rename(ctx, rename)
}

// rename stores the change and then moves the profile on the season servers. The nickname must be free: a licensed
// account whose new name another profile holds marks that profile as a premium conflict for an admin.
func (s *profileRenameService) rename(ctx context.Context, rename *profileRename) (*domain.Profile, error) {
	profileID := rename.before.ID

	holder, err := s.storage.Queries().FindProfileByUsername(ctx, rename.username)
	if err != nil && !errors.Is(err, stdsql.ErrNoRows) {
		return nil, utils.NewInternalServerError("failed to find the profile with the nickname", err)
	}
	if err == nil && holder.ID != profileID {
		if !rename.licensed() {
			return nil, utils.NewConflictError("the nickname is taken by another profile", nil)
		}
		// Mojang gave the name to this account, so the other profile can no longer join with it.
		logger.Warnf(ctx, "[PROFILE RENAME] Nickname %s of licensed profile %s is held by profile %s, marked as premium conflict", rename.username, profileID, holder.ID)
		if err := s.storage.Queries().SetProfilePremiumConflict(ctx, holder.MinecraftUUID); err != nil {
			logger.Errorf(ctx, "[PROFILE RENAME] Failed to mark profile %s as premium conflict: %v", holder.ID, err)
		}
		return nil, utils.NewConflictError("the nickname is held by another profile, an admin will sort it out", nil)
	}

	var renamed *domain.Profile
	err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if err := queries.LockProfile(ctx, profileID); err != nil {
			return utils.NewInternalServerError("failed to lock profile", err)
		}
		current, err := queries.FindProfileByID(ctx, profileID)
		if err != nil {
			return utils.NewInternalServerError("failed to get profile", err)
		}
		if current.MinecraftUUID != rename.before.MinecraftUUID || current.MinecraftUsername != rename.before.MinecraftUsername {
			return utils.NewConflictError("the profile changed meanwhile, try again", nil)
		}

		if rename.source == domain.ProfileUsernameChangeSourceOwner && !rename.licensed() {
			availableAt, err := s.usernameChangeAvailableAt(ctx, queries, profileID)
			if err != nil {
				return err
			}
			if availableAt != nil {
				return utils.NewTooManyRequestsError("the nickname can be changed again on "+availableAt.UTC().Format(time.DateOnly), nil)
			}
		}

		if !rename.licensed() {
			holders, err := queries.FindProfileIDsHoldingMinecraftUUID(ctx, rename.newMcUUID)
			if err != nil {
				return utils.NewInternalServerError("failed to check the uuid of the nickname", err)
			}
			for _, holderID := range holders {
				if holderID != profileID {
					return utils.NewConflictError("the nickname is reserved by another profile", nil)
				}
			}
		}

		moved, err := queries.RenameProfile(ctx, profileID, rename.before.MinecraftUUID, rename.newMcUUID, rename.username, rename.changedBy)
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return utils.NewConflictError("the nickname is taken by another profile", err)
		} else if err != nil {
			return utils.NewInternalServerError("failed to rename profile", err)
		}
		if !moved {
			return utils.NewConflictError("the profile changed meanwhile, try again", nil)
		}

		if rename.licensed() {
			// The name was just compared with Mojang, so the name sync can wait a day again.
			if err := queries.UpsertProfileMojangUUID(ctx, rename.newMcUUID, &rename.newMcUUID); err != nil {
				return utils.NewInternalServerError("failed to store profile mojang uuid", err)
			}
		} else {
			// Back to an own old nickname: the UUID is the profile's again, not a former one.
			if err := queries.DeleteProfileFormerUUID(ctx, rename.newMcUUID, profileID); err != nil {
				return utils.NewInternalServerError("failed to take the former uuid back", err)
			}
			if !rename.leavesLicense {
				if err := queries.InsertProfileFormerUUID(ctx, rename.before.MinecraftUUID, profileID); err != nil {
					return utils.NewInternalServerError("failed to keep the former uuid", err)
				}
			}
			// The lookup row followed the UUID, and the nickname was just checked to have no Mojang account.
			if err := queries.UpsertProfileMojangUUID(ctx, rename.newMcUUID, nil); err != nil {
				return utils.NewInternalServerError("failed to store profile mojang uuid", err)
			}
		}

		err = queries.InsertProfileUsernameChange(ctx, sql.InsertProfileUsernameChangeParams{
			ProfileID:        profileID,
			OldUsername:      rename.before.MinecraftUsername,
			NewUsername:      rename.username,
			OldMinecraftUUID: rename.before.MinecraftUUID,
			NewMinecraftUUID: rename.newMcUUID,
			Source:           rename.source,
			ChangedBy:        rename.changedBy,
		})
		if err != nil {
			return utils.NewInternalServerError("failed to record the nickname change", err)
		}

		renamed = current
		renamed.MinecraftUUID = rename.newMcUUID
		renamed.MinecraftUsername = rename.username
		return nil
	})
	if err != nil {
		return nil, err
	}

	logger.Infof(ctx, "[PROFILE RENAME] Profile %s renamed from %s (%s) to %s (%s) by %s, left the license: %t", profileID, rename.before.MinecraftUsername, rename.before.MinecraftUUID, rename.username, rename.newMcUUID, rename.source, rename.leavesLicense)
	s.profileResyncService.MoveProfileOnServers(ctx, profileID, rename.before)
	return renamed, nil
}

// usernameChangeAvailableAt returns when the owner can change the nickname again, nil when the owner can now.
func (s *profileRenameService) usernameChangeAvailableAt(ctx context.Context, queries sql.Queries, profileID uuid.UUID) (*time.Time, error) {
	if s.cooldown <= 0 {
		return nil, nil
	}
	lastChangeAt, err := queries.FindLastProfileUsernameChangeAt(ctx, profileID, domain.ProfileUsernameChangeSourceOwner)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get the last nickname change", err)
	}
	if lastChangeAt == nil {
		return nil, nil
	}
	availableAt := lastChangeAt.Add(s.cooldown)
	if !availableAt.After(s.now()) {
		return nil, nil
	}
	return &availableAt, nil
}

func (s *profileRenameService) GetUsernameChangeAvailableAt(ctx context.Context, profileIDs uuid.UUIDs) (map[uuid.UUID]time.Time, error) {
	res := make(map[uuid.UUID]time.Time)
	if s.cooldown <= 0 || len(profileIDs) == 0 {
		return res, nil
	}
	lastChanges, err := s.storage.Queries().FindLastProfileUsernameChangeAtByProfileIDs(ctx, profileIDs, domain.ProfileUsernameChangeSourceOwner)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get the last nickname changes", err)
	}
	now := s.now()
	for profileID, lastChangeAt := range lastChanges {
		if availableAt := lastChangeAt.Add(s.cooldown); availableAt.After(now) {
			res[profileID] = availableAt
		}
	}
	return res, nil
}

func (s *profileRenameService) GetUsernameChanges(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileUsernameChange, error) {
	changes, err := s.storage.Queries().FindProfileUsernameChanges(ctx, profileID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get the nickname changes", err)
	}
	return changes, nil
}

func (s *profileRenameService) RunPremiumNameSync(ctx context.Context) {
	wait := premiumNameSyncInterval
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}

		wait = premiumNameSyncInterval
		err := s.syncPremiumNames(ctx)
		switch {
		case err == nil, errors.Is(err, context.Canceled):
		case errors.Is(err, errMojangRateLimited):
			logger.Warnf(ctx, "[PROFILE RENAME] Rate limited, premium name sync paused for %s", mojangSyncRateLimitBackoff)
			wait = mojangSyncRateLimitBackoff
		default:
			logger.Errorf(ctx, "[PROFILE RENAME] Failed to sync premium names: %v", err)
		}
	}
}

// syncPremiumNames compares the names of the licensed profiles checked longest ago with Mojang and follows a rename.
func (s *profileRenameService) syncPremiumNames(ctx context.Context) error {
	targets, err := s.storage.Queries().FindPremiumNameCheckTargets(ctx, s.now().Add(-premiumNameCheckInterval), premiumNameSyncBatchSize)
	if err != nil {
		return err
	}

	for i, target := range targets {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(premiumNameLookupDelay):
			}
		}

		name, err := s.mojangService.LookupUsernameByMojangUUID(ctx, target.MinecraftUUID)
		if err != nil {
			return err
		}
		// An empty name is an account Mojang no longer has; the profile keeps the name it had.
		if name != "" && name != target.MinecraftUsername {
			_, err := s.rename(ctx, &profileRename{
				before:    target,
				username:  name,
				newMcUUID: target.MinecraftUUID,
				source:    domain.ProfileUsernameChangeSourceMojang,
			})
			if err != nil {
				logger.Warnf(ctx, "[PROFILE RENAME] Failed to rename licensed profile %s from %s to %s: %v", target.ID, target.MinecraftUsername, name, err)
			}
		}
		if err := s.storage.Queries().UpsertProfileMojangUUID(ctx, target.MinecraftUUID, &target.MinecraftUUID); err != nil {
			return err
		}
	}
	return nil
}
