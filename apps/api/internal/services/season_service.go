package services

import (
	"context"
	stdsql "database/sql"
	"errors"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type SeasonService interface {
	GetSeasonByID(ctx context.Context, seasonID uuid.UUID) (*domain.Season, error)
	// GetPrimarySeasonID returns the season that requests fall back to when they name none.
	GetPrimarySeasonID(ctx context.Context) (uuid.UUID, error)
	GetPublicSeasons(ctx context.Context) ([]*domain.Season, error)
	// GetSeasons returns every season, the newest start first.
	GetSeasons(ctx context.Context) ([]*domain.Season, error)
	CreateSeason(ctx context.Context, cmd *commands.SaveSeasonCommand) (*domain.Season, error)
	UpdateSeason(ctx context.Context, cmd *commands.SaveSeasonCommand) (*domain.Season, error)
	DeleteSeason(ctx context.Context, seasonID uuid.UUID) error
	InitializePrimarySeason(ctx context.Context) error

	// GetSeasonScreenshots returns the feed of the season, checking first that the season exists.
	GetSeasonScreenshots(ctx context.Context, seasonID uuid.UUID) ([]*domain.SeasonScreenshot, error)
	CreateSeasonScreenshot(ctx context.Context, cmd *commands.SaveSeasonScreenshotCommand) (*domain.SeasonScreenshot, error)
	UpdateSeasonScreenshot(ctx context.Context, cmd *commands.SaveSeasonScreenshotCommand) (*domain.SeasonScreenshot, error)
	DeleteSeasonScreenshot(ctx context.Context, seasonID, screenshotID uuid.UUID) error
}

// The cache is dropped whenever the primary season changes, the TTL only covers a missed drop.
const (
	primarySeasonCacheKey = "primary_season_id"
	primarySeasonCacheTTL = 5 * time.Minute
)

type seasonService struct {
	storage storage.MainStorage
	cache   storage.CacheStorage
}

func NewSeasonService(storage storage.MainStorage, cache storage.CacheStorage) SeasonService {
	return &seasonService{storage: storage, cache: cache}
}

func (s *seasonService) GetSeasonByID(ctx context.Context, seasonID uuid.UUID) (*domain.Season, error) {
	season, err := s.storage.Queries().FindSeasonByID(ctx, seasonID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return nil, utils.NewNotFoundError("season not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get season by id", err)
	}
	return season, nil
}

// GetPrimarySeasonID reads through the cache, a cache failure only costs a database query.
func (s *seasonService) GetPrimarySeasonID(ctx context.Context) (uuid.UUID, error) {
	if cached, err := s.cache.GetKey(ctx, primarySeasonCacheKey); err == nil {
		if id, err := uuid.Parse(cached); err == nil {
			return id, nil
		}
	}

	id, err := s.storage.Queries().FindPrimarySeasonID(ctx)
	if errors.Is(err, stdsql.ErrNoRows) {
		return uuid.Nil, utils.NewInternalServerError("primary season is not set", err)
	} else if err != nil {
		return uuid.Nil, utils.NewInternalServerError("failed to get primary season", err)
	}

	if err := s.cache.SetKey(ctx, primarySeasonCacheKey, id.String(), primarySeasonCacheTTL); err != nil {
		logger.Debugf(ctx, "[SEASON] Error caching primary season: %v", err)
	}
	return id, nil
}

func (s *seasonService) dropPrimarySeasonCache(ctx context.Context) {
	if err := s.cache.DeleteKey(ctx, primarySeasonCacheKey); err != nil {
		logger.Errorf(ctx, "[SEASON] Error dropping primary season cache: %v", err)
	}
}

func (s *seasonService) GetPublicSeasons(ctx context.Context) ([]*domain.Season, error) {
	seasons, err := s.storage.Queries().FindPublicSeasons(ctx)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get public seasons", err)
	}
	return seasons, nil
}

func (s *seasonService) GetSeasons(ctx context.Context) ([]*domain.Season, error) {
	seasons, err := s.storage.Queries().FindSeasons(ctx)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get seasons", err)
	}
	return seasons, nil
}

func insertSeasonParams(id uuid.UUID, cmd *commands.SaveSeasonCommand) sql.InsertSeasonParams {
	return sql.InsertSeasonParams{
		ID: id, Name: cmd.Name, PreviewImage: cmd.PreviewImage,
		StartDate: cmd.StartDate, EndDate: cmd.EndDate,
		PublicAddress: cmd.PublicAddress,
		ShellAddress:  cmd.ShellAddress,
		PlanURL:       cmd.PlanURL,
		IsActive:      cmd.IsActive, Preregistration: cmd.Preregistration,
		FreeRegistration: cmd.FreeRegistration,
		GameVersion:      cmd.GameVersion, WorldURL: cmd.WorldURL,
	}
}

func updateSeasonParams(cmd *commands.SaveSeasonCommand) sql.UpdateSeasonParams {
	return sql.UpdateSeasonParams{
		ID: cmd.ID, Name: cmd.Name, PreviewImage: cmd.PreviewImage,
		StartDate: cmd.StartDate, EndDate: cmd.EndDate,
		PublicAddress:   cmd.PublicAddress,
		ShellAddress:    cmd.ShellAddress,
		PlanURL:         cmd.PlanURL,
		IsActive:        cmd.IsActive,
		Preregistration: cmd.Preregistration, FreeRegistration: cmd.FreeRegistration,
		GameVersion: cmd.GameVersion, WorldURL: cmd.WorldURL,
	}
}

func setPrimarySeason(ctx context.Context, queries sql.Queries, seasonID uuid.UUID) error {
	if err := queries.ClearPrimarySeasons(ctx); err != nil {
		return err
	}
	found, err := queries.SetSeasonPrimary(ctx, seasonID)
	if err != nil {
		return err
	}
	if !found {
		return stdsql.ErrNoRows
	}
	return nil
}

// ensureShellAddressFree fails with a conflict when another season already uses the shell address.
// The unique index has the last word, this check only names the season that holds the address.
func (s *seasonService) ensureShellAddressFree(ctx context.Context, address *string, seasonID uuid.UUID) error {
	if address == nil {
		return nil
	}
	seasons, err := s.GetSeasons(ctx)
	if err != nil {
		return err
	}
	if holder := seasonWithShellAddress(seasons, *address, seasonID); holder != nil {
		return utils.NewConflictError("shell address is used by season "+holder.Name, nil)
	}
	return nil
}

// seasonWithShellAddress returns the season other than exceptID that uses the shell address.
func seasonWithShellAddress(seasons []*domain.Season, address string, exceptID uuid.UUID) *domain.Season {
	for _, season := range seasons {
		if season.ID != exceptID && season.ShellAddress != nil && *season.ShellAddress == address {
			return season
		}
	}
	return nil
}

// seasonWriteError maps a write that hit the unique shell address index to a conflict.
func seasonWriteError(msg string, err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "idx_seasons_shell_address") {
		return utils.NewConflictError("shell address is used by another season", err)
	}
	return utils.NewInternalServerError(msg, err)
}

func (s *seasonService) CreateSeason(ctx context.Context, cmd *commands.SaveSeasonCommand) (*domain.Season, error) {
	id := uuid.New()
	if err := s.ensureShellAddressFree(ctx, cmd.ShellAddress, id); err != nil {
		return nil, err
	}
	err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if err := queries.InsertSeason(ctx, insertSeasonParams(id, cmd)); err != nil {
			return err
		}
		if cmd.IsPrimary {
			return setPrimarySeason(ctx, queries, id)
		}
		return nil
	})
	if err != nil {
		return nil, seasonWriteError("failed to create season", err)
	}
	if cmd.IsPrimary {
		s.dropPrimarySeasonCache(ctx)
	}
	return s.GetSeasonByID(ctx, id)
}

func (s *seasonService) UpdateSeason(ctx context.Context, cmd *commands.SaveSeasonCommand) (*domain.Season, error) {
	current, err := s.GetSeasonByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if current.IsPrimary && !cmd.IsPrimary {
		return nil, utils.NewConflictError("select another primary season before removing this one", nil)
	}
	if err := s.ensureShellAddressFree(ctx, cmd.ShellAddress, cmd.ID); err != nil {
		return nil, err
	}

	becomesPrimary := cmd.IsPrimary && !current.IsPrimary
	err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if err := queries.UpdateSeason(ctx, updateSeasonParams(cmd)); err != nil {
			return err
		}
		if becomesPrimary {
			return setPrimarySeason(ctx, queries, cmd.ID)
		}
		return nil
	})
	if err != nil {
		return nil, seasonWriteError("failed to update season", err)
	}
	if becomesPrimary {
		s.dropPrimarySeasonCache(ctx)
	}
	return s.GetSeasonByID(ctx, cmd.ID)
}

func (s *seasonService) DeleteSeason(ctx context.Context, seasonID uuid.UUID) error {
	season, err := s.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return err
	}
	if season.IsPrimary {
		return utils.NewConflictError("primary season cannot be deleted", nil)
	}
	if season.IsActive {
		return utils.NewConflictError("active season cannot be deleted", nil)
	}
	deleted, err := s.storage.Queries().DeleteSeason(ctx, seasonID)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1451 {
			return utils.NewConflictError("season is used by profiles, grants or orders", err)
		}
		return utils.NewInternalServerError("failed to delete season", err)
	}
	if !deleted {
		current, lookupErr := s.storage.Queries().FindSeasonByID(ctx, seasonID)
		if lookupErr == nil && current.IsPrimary {
			return utils.NewConflictError("primary season cannot be deleted", nil)
		}
		if lookupErr == nil && current.IsActive {
			return utils.NewConflictError("active season cannot be deleted", nil)
		}
		if errors.Is(lookupErr, stdsql.ErrNoRows) {
			return utils.NewNotFoundError("season not found", lookupErr)
		}
		return utils.NewInternalServerError("failed to verify season deletion", lookupErr)
	}
	return nil
}

// InitializePrimarySeason moves the deployment-wide SHELL_ADDRESS into the primary season on old deployments.
func (s *seasonService) InitializePrimarySeason(ctx context.Context) error {
	seasons, err := s.GetSeasons(ctx)
	if err != nil {
		return err
	}
	for _, season := range seasons {
		if season.IsPrimary {
			return s.seedShellAddress(ctx, season, seasons)
		}
	}
	return nil
}

// seedShellAddress moves the deployment-wide SHELL_ADDRESS into a primary season that has none yet.
// It leaves the season alone when another season already serves the address, because a shell serves one season.
func (s *seasonService) seedShellAddress(ctx context.Context, season *domain.Season, seasons []*domain.Season) error {
	address := config.GetShellAddress()
	if season.ShellAddress != nil || address == "" {
		return nil
	}
	if holder := seasonWithShellAddress(seasons, address, season.ID); holder != nil {
		logger.Warnf(ctx, "[SEASON] Shell address %s is used by season %s, primary season keeps none", address, holder.Name)
		return nil
	}
	if err := s.storage.Queries().SetSeasonShellAddress(ctx, season.ID, address); err != nil {
		return utils.NewInternalServerError("failed to seed season shell address", err)
	}
	return nil
}

func (s *seasonService) GetSeasonScreenshots(ctx context.Context, seasonID uuid.UUID) ([]*domain.SeasonScreenshot, error) {
	if _, err := s.GetSeasonByID(ctx, seasonID); err != nil {
		return nil, err
	}
	screenshots, err := s.storage.Queries().FindSeasonScreenshots(ctx, seasonID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get season screenshots", err)
	}
	return screenshots, nil
}

// checkAuthorProfilesExist fails with a 400 naming nothing further: an admin cannot credit a profile id that
// does not exist, that is a mistake in the request, not a silently dropped credit.
func (s *seasonService) checkAuthorProfilesExist(ctx context.Context, profileIDs uuid.UUIDs) error {
	if len(profileIDs) == 0 {
		return nil
	}
	found, err := s.storage.Queries().FindProfilesByIDs(ctx, profileIDs)
	if err != nil {
		return utils.NewInternalServerError("failed to verify screenshot authors", err)
	}
	if len(found) != len(profileIDs) {
		return utils.NewBadRequestError("one or more author profiles do not exist", nil)
	}
	return nil
}

func (s *seasonService) CreateSeasonScreenshot(ctx context.Context, cmd *commands.SaveSeasonScreenshotCommand) (*domain.SeasonScreenshot, error) {
	if _, err := s.GetSeasonByID(ctx, cmd.SeasonID); err != nil {
		return nil, err
	}
	if err := s.checkAuthorProfilesExist(ctx, cmd.AuthorProfileIDs); err != nil {
		return nil, err
	}

	id := uuid.New()
	err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if err := queries.CreateSeasonScreenshot(ctx, sql.SaveSeasonScreenshotParams{
			ID: id, SeasonID: cmd.SeasonID, Image: cmd.Image, Title: cmd.Title, Position: cmd.Position,
		}); err != nil {
			return err
		}
		return queries.SetSeasonScreenshotAuthors(ctx, id, cmd.AuthorProfileIDs)
	})
	if err != nil {
		return nil, utils.NewInternalServerError("failed to create season screenshot", err)
	}
	return s.storage.Queries().FindSeasonScreenshotByID(ctx, id, cmd.SeasonID)
}

func (s *seasonService) UpdateSeasonScreenshot(ctx context.Context, cmd *commands.SaveSeasonScreenshotCommand) (*domain.SeasonScreenshot, error) {
	if _, err := s.GetSeasonByID(ctx, cmd.SeasonID); err != nil {
		return nil, err
	}
	if err := s.checkAuthorProfilesExist(ctx, cmd.AuthorProfileIDs); err != nil {
		return nil, err
	}

	var found bool
	err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		var err error
		found, err = queries.UpdateSeasonScreenshot(ctx, sql.SaveSeasonScreenshotParams{
			ID: cmd.ID, SeasonID: cmd.SeasonID, Image: cmd.Image, Title: cmd.Title, Position: cmd.Position,
		})
		if err != nil || !found {
			return err
		}
		return queries.SetSeasonScreenshotAuthors(ctx, cmd.ID, cmd.AuthorProfileIDs)
	})
	if err != nil {
		return nil, utils.NewInternalServerError("failed to update season screenshot", err)
	}
	if !found {
		return nil, utils.NewNotFoundError("screenshot not found", nil)
	}
	return s.storage.Queries().FindSeasonScreenshotByID(ctx, cmd.ID, cmd.SeasonID)
}

func (s *seasonService) DeleteSeasonScreenshot(ctx context.Context, seasonID, screenshotID uuid.UUID) error {
	deleted, err := s.storage.Queries().DeleteSeasonScreenshot(ctx, screenshotID, seasonID)
	if err != nil {
		return utils.NewInternalServerError("failed to delete season screenshot", err)
	}
	if !deleted {
		return utils.NewNotFoundError("screenshot not found", nil)
	}
	return nil
}
