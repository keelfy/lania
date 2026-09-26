package services

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type ChunkClaimService interface {
	// GetActiveClaims returns every claim that holds in the dimension of the world, and the world.
	GetActiveClaims(ctx context.Context, worldID uuid.UUID, dimension string) (*domain.SeasonWorld, []*domain.ChunkClaim, error)
	// ClaimChunks claims every chunk for the profile of the user, or none of them.
	ClaimChunks(ctx context.Context, cmd *commands.ClaimChunksCommand) error
	// ReleaseChunks ends the claims of the profile of the user on every chunk, or on none of them.
	ReleaseChunks(ctx context.Context, cmd *commands.ClaimChunksCommand) error
	// AdminReleaseChunks ends whatever claims hold on the chunks, in any season and dimension.
	AdminReleaseChunks(ctx context.Context, cmd *commands.ClaimChunksCommand) error
}

type chunkClaimService struct {
	storage          storage.MainStorage
	seasonService    SeasonService
	worldService     SeasonWorldService
	minecraftService MinecraftService
}

func NewChunkClaimService(
	storage storage.MainStorage,
	seasonService SeasonService,
	worldService SeasonWorldService,
	minecraftService MinecraftService,
) ChunkClaimService {
	return &chunkClaimService{
		storage: storage, seasonService: seasonService, worldService: worldService, minecraftService: minecraftService,
	}
}

func (s *chunkClaimService) GetActiveClaims(ctx context.Context, worldID uuid.UUID, dimension string) (*domain.SeasonWorld, []*domain.ChunkClaim, error) {
	world, err := s.worldService.GetSeasonWorld(ctx, worldID)
	if err != nil {
		return nil, nil, err
	}
	claims, err := s.storage.Queries().FindActiveChunkClaims(ctx, worldID, dimension)
	if err != nil {
		return nil, nil, utils.NewInternalServerError("failed to get chunk claims", err)
	}
	return world, claims, nil
}

// claimableWorld returns the world and its season when players can claim and release chunks of the dimension:
// the season runs, the world has a map and claims are on in the dimension.
func (s *chunkClaimService) claimableWorld(ctx context.Context, worldID uuid.UUID, dimension string) (*domain.SeasonWorld, *domain.Season, error) {
	world, err := s.worldService.GetSeasonWorld(ctx, worldID)
	if err != nil {
		return nil, nil, err
	}
	season, err := s.seasonService.GetSeasonByID(ctx, world.SeasonID)
	if err != nil {
		return nil, nil, err
	}
	if !season.IsActive {
		return nil, nil, utils.NewBadRequestError("chunks can be claimed in an active season only", nil)
	}
	if !world.ClaimsIn(dimension) {
		return nil, nil, utils.NewBadRequestError("chunk claims are off in this dimension", nil)
	}
	return world, season, nil
}

// ownedProfile returns the profile when the user owns it.
func (s *chunkClaimService) ownedProfile(ctx context.Context, profileID, userID uuid.UUID) (*domain.Profile, error) {
	profile, err := s.storage.Queries().FindProfileByID(ctx, profileID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return nil, utils.NewNotFoundError("profile not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to find profile", err)
	}
	if profile.OwnerUserID == nil || *profile.OwnerUserID != userID {
		return nil, utils.NewForbiddenError("only the owner can manage the claims of a profile", nil)
	}
	return profile, nil
}

func (s *chunkClaimService) ClaimChunks(ctx context.Context, cmd *commands.ClaimChunksCommand) error {
	world, season, err := s.claimableWorld(ctx, cmd.WorldID, cmd.Dimension)
	if err != nil {
		return err
	}
	profile, err := s.ownedProfile(ctx, cmd.ProfileID, cmd.UserID)
	if err != nil {
		return err
	}
	hasAccess, err := s.storage.Queries().CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx, profile.MinecraftUUID, season.ID)
	if err != nil {
		return utils.NewInternalServerError("failed to check season access", err)
	}
	if !hasAccess {
		return utils.NewForbiddenError("the profile has no access to the season", nil)
	}
	if err := s.checkClaimPlaytime(ctx, world, profile); err != nil {
		return err
	}

	err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if err := queries.LockProfile(ctx, profile.ID); err != nil {
			return err
		}
		held, err := queries.CountActiveChunkClaimsByProfile(ctx, profile.ID, world.ID)
		if err != nil {
			return err
		}
		if held+len(cmd.Chunks) > world.ClaimLimit {
			return utils.NewConflictError(fmt.Sprintf("claim limit is %d chunks, the profile holds %d", world.ClaimLimit, held), nil)
		}
		taken, err := queries.CountActiveChunkClaimsAt(ctx, world.ID, cmd.Dimension, cmd.Chunks)
		if err != nil {
			return err
		}
		if taken > 0 {
			return utils.NewConflictError("some of the chunks are already claimed", nil)
		}
		return queries.InsertChunkClaims(ctx, world.ID, cmd.Dimension, profile.ID, cmd.Chunks)
	})
	return claimWriteError("failed to claim chunks", err)
}

// checkClaimPlaytime refuses a profile that has played less than the world asks on the world server. Playtime
// under the offline UUID of a rekeyed profile counts too.
func (s *chunkClaimService) checkClaimPlaytime(ctx context.Context, world *domain.SeasonWorld, profile *domain.Profile) error {
	if world.ClaimMinPlaytimeHours <= 0 {
		return nil
	}
	mcUUIDs := uuid.UUIDs{profile.MinecraftUUID}
	if profile.LegacyMinecraftUUID != nil {
		mcUUIDs = append(mcUUIDs, *profile.LegacyMinecraftUUID)
	}
	playtimes, err := s.minecraftService.GetPlaytimesInSeason(ctx, world.SeasonID, world.PlanServer, mcUUIDs)
	if err != nil {
		return utils.NewServiceUnavailableError("failed to get playtime on the world server", err)
	}
	var played time.Duration
	for _, playtime := range playtimes {
		played += time.Duration(playtime.TotalPlaytime) * time.Millisecond
	}
	if played < world.ClaimMinPlaytime() {
		return utils.NewForbiddenError(fmt.Sprintf("claims need %d hours played on the world server", world.ClaimMinPlaytimeHours), nil)
	}
	return nil
}

func (s *chunkClaimService) ReleaseChunks(ctx context.Context, cmd *commands.ClaimChunksCommand) error {
	if _, _, err := s.claimableWorld(ctx, cmd.WorldID, cmd.Dimension); err != nil {
		return err
	}
	if _, err := s.ownedProfile(ctx, cmd.ProfileID, cmd.UserID); err != nil {
		return err
	}
	return s.release(ctx, cmd, &cmd.ProfileID, "some of the chunks are not claimed by the profile")
}

func (s *chunkClaimService) AdminReleaseChunks(ctx context.Context, cmd *commands.ClaimChunksCommand) error {
	if _, err := s.worldService.GetSeasonWorld(ctx, cmd.WorldID); err != nil {
		return err
	}
	return s.release(ctx, cmd, nil, "some of the chunks are not claimed")
}

// release ends the claims on every chunk or on none: a request naming a chunk it cannot release changes nothing.
func (s *chunkClaimService) release(ctx context.Context, cmd *commands.ClaimChunksCommand, profileID *uuid.UUID, missing string) error {
	err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		released, err := queries.ReleaseChunkClaims(ctx, cmd.WorldID, cmd.Dimension, profileID, cmd.Chunks, cmd.UserID)
		if err != nil {
			return err
		}
		if released != int64(len(cmd.Chunks)) {
			return utils.NewConflictError(missing, nil)
		}
		return nil
	})
	return claimWriteError("failed to release chunks", err)
}

// claimWriteError keeps the errors the transaction raised on purpose and maps a claim that lost the race for
// a chunk to idx_chunk_claims_active to a conflict.
func claimWriteError(msg string, err error) error {
	if err == nil {
		return nil
	}
	var customErr *utils.CustomError
	if errors.As(err, &customErr) {
		return err
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return utils.NewConflictError("some of the chunks are already claimed", err)
	}
	return utils.NewInternalServerError(msg, err)
}
