package services

import (
	"context"
	stdsql "database/sql"
	"errors"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type SeasonService interface {
	GetSeasonByID(ctx context.Context, seasonID uuid.UUID) (*domain.Season, error)
	GetPublicSeasons(ctx context.Context) ([]*domain.Season, error)
	// GetSeasons returns every season, the newest start first.
	GetSeasons(ctx context.Context) ([]*domain.Season, error)
	CreateSeason(ctx context.Context, cmd *commands.SaveSeasonCommand) (*domain.Season, error)
	UpdateSeason(ctx context.Context, cmd *commands.SaveSeasonCommand) (*domain.Season, error)
	DeleteSeason(ctx context.Context, seasonID uuid.UUID) error
	InitializeActiveSeason(ctx context.Context) error
}

type seasonService struct {
	storage storage.MainStorage
}

func NewSeasonService(storage storage.MainStorage) SeasonService {
	return &seasonService{storage: storage}
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
		ID: id, SeasonNumber: cmd.SeasonNumber, Name: cmd.Name, PreviewImage: cmd.PreviewImage,
		StartDate: cmd.StartDate, EndDate: cmd.EndDate, ServerIP: cmd.ServerIP, ServerPort: cmd.ServerPort,
		RCONPassword: cmd.RCONPassword,
	}
}

func updateSeasonParams(cmd *commands.SaveSeasonCommand) sql.UpdateSeasonParams {
	return sql.UpdateSeasonParams{
		ID: cmd.ID, SeasonNumber: cmd.SeasonNumber, Name: cmd.Name, PreviewImage: cmd.PreviewImage,
		StartDate: cmd.StartDate, EndDate: cmd.EndDate, ServerIP: cmd.ServerIP, ServerPort: cmd.ServerPort,
		RCONPassword: cmd.RCONPassword, SetRCONPassword: cmd.SetRCONPassword,
	}
}

func activateSeason(ctx context.Context, queries sql.Queries, seasonID uuid.UUID) error {
	if err := queries.ClearActiveSeasons(ctx); err != nil {
		return err
	}
	found, err := queries.SetSeasonActive(ctx, seasonID)
	if err != nil {
		return err
	}
	if !found {
		return stdsql.ErrNoRows
	}
	return nil
}

func (s *seasonService) CreateSeason(ctx context.Context, cmd *commands.SaveSeasonCommand) (*domain.Season, error) {
	id := uuid.New()
	err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if err := queries.InsertSeason(ctx, insertSeasonParams(id, cmd)); err != nil {
			return err
		}
		if cmd.IsActive {
			return activateSeason(ctx, queries, id)
		}
		return nil
	})
	if err != nil {
		return nil, utils.NewInternalServerError("failed to create season", err)
	}
	if cmd.IsActive {
		config.SetActiveSeasonID(id)
	}
	return s.GetSeasonByID(ctx, id)
}

func (s *seasonService) UpdateSeason(ctx context.Context, cmd *commands.SaveSeasonCommand) (*domain.Season, error) {
	current, err := s.GetSeasonByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if current.IsActive && !cmd.IsActive {
		return nil, utils.NewConflictError("activate another season before deactivating this one", nil)
	}

	err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if err := queries.UpdateSeason(ctx, updateSeasonParams(cmd)); err != nil {
			return err
		}
		if cmd.IsActive && !current.IsActive {
			return activateSeason(ctx, queries, cmd.ID)
		}
		return nil
	})
	if err != nil {
		return nil, utils.NewInternalServerError("failed to update season", err)
	}
	if cmd.IsActive {
		config.SetActiveSeasonID(cmd.ID)
	}
	return s.GetSeasonByID(ctx, cmd.ID)
}

func (s *seasonService) DeleteSeason(ctx context.Context, seasonID uuid.UUID) error {
	season, err := s.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return err
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

// InitializeActiveSeason makes the database setting authoritative while preserving old deployments.
func (s *seasonService) InitializeActiveSeason(ctx context.Context) error {
	seasons, err := s.GetSeasons(ctx)
	if err != nil {
		return err
	}
	for _, season := range seasons {
		if season.IsActive {
			config.SetActiveSeasonID(season.ID)
			return nil
		}
	}

	fallback := config.GetActiveSeasonID()
	if err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		return activateSeason(ctx, queries, fallback)
	}); err != nil {
		return utils.NewInternalServerError("failed to initialize active season", err)
	}
	config.SetActiveSeasonID(fallback)
	return nil
}
