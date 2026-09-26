package services

import (
	"context"
	stdsql "database/sql"
	"errors"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type SeasonWorldService interface {
	// GetSeasonWorlds returns the worlds of the season in display order, checking first that the season exists.
	GetSeasonWorlds(ctx context.Context, seasonID uuid.UUID) ([]*domain.SeasonWorld, error)
	GetSeasonWorld(ctx context.Context, worldID uuid.UUID) (*domain.SeasonWorld, error)
	CreateSeasonWorld(ctx context.Context, cmd *commands.SaveSeasonWorldCommand) (*domain.SeasonWorld, error)
	UpdateSeasonWorld(ctx context.Context, cmd *commands.SaveSeasonWorldCommand) (*domain.SeasonWorld, error)
	// DeleteSeasonWorld refuses a world that ever had claims: their history would go with it.
	DeleteSeasonWorld(ctx context.Context, worldID uuid.UUID) error
}

type seasonWorldService struct {
	storage       storage.MainStorage
	seasonService SeasonService
}

func NewSeasonWorldService(storage storage.MainStorage, seasonService SeasonService) SeasonWorldService {
	return &seasonWorldService{storage: storage, seasonService: seasonService}
}

func (s *seasonWorldService) GetSeasonWorlds(ctx context.Context, seasonID uuid.UUID) ([]*domain.SeasonWorld, error) {
	if _, err := s.seasonService.GetSeasonByID(ctx, seasonID); err != nil {
		return nil, err
	}
	worlds, err := s.storage.Queries().FindSeasonWorlds(ctx, seasonID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get season worlds", err)
	}
	return worlds, nil
}

func (s *seasonWorldService) GetSeasonWorld(ctx context.Context, worldID uuid.UUID) (*domain.SeasonWorld, error) {
	world, err := s.storage.Queries().FindSeasonWorldByID(ctx, worldID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return nil, utils.NewNotFoundError("world not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get season world", err)
	}
	return world, nil
}

func saveSeasonWorldParams(id uuid.UUID, cmd *commands.SaveSeasonWorldCommand) sql.SaveSeasonWorldParams {
	return sql.SaveSeasonWorldParams{
		ID: id, SeasonID: cmd.SeasonID, Slug: cmd.Slug, Name: cmd.Name, PreviewImage: cmd.PreviewImage,
		MapURL: cmd.MapURL, ClaimLimit: cmd.ClaimLimit, ClaimDimensions: cmd.ClaimDimensions,
		PlanServer: cmd.PlanServer, ClaimMinPlaytimeHours: cmd.ClaimMinPlaytimeHours, Position: cmd.Position,
	}
}

func (s *seasonWorldService) CreateSeasonWorld(ctx context.Context, cmd *commands.SaveSeasonWorldCommand) (*domain.SeasonWorld, error) {
	if _, err := s.seasonService.GetSeasonByID(ctx, cmd.SeasonID); err != nil {
		return nil, err
	}
	id := uuid.New()
	if err := s.storage.Queries().CreateSeasonWorld(ctx, saveSeasonWorldParams(id, cmd)); err != nil {
		return nil, seasonWorldWriteError("failed to create season world", err)
	}
	return s.GetSeasonWorld(ctx, id)
}

func (s *seasonWorldService) UpdateSeasonWorld(ctx context.Context, cmd *commands.SaveSeasonWorldCommand) (*domain.SeasonWorld, error) {
	world, err := s.GetSeasonWorld(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	// A world stays in its season: its claims count against that season's access.
	cmd.SeasonID = world.SeasonID
	found, err := s.storage.Queries().UpdateSeasonWorld(ctx, saveSeasonWorldParams(cmd.ID, cmd))
	if err != nil {
		return nil, seasonWorldWriteError("failed to update season world", err)
	}
	if !found {
		return nil, utils.NewNotFoundError("world not found", nil)
	}
	return s.GetSeasonWorld(ctx, cmd.ID)
}

func (s *seasonWorldService) DeleteSeasonWorld(ctx context.Context, worldID uuid.UUID) error {
	claims, err := s.storage.Queries().CountChunkClaimsInWorld(ctx, worldID)
	if err != nil {
		return utils.NewInternalServerError("failed to count world claims", err)
	}
	if claims > 0 {
		return utils.NewConflictError("a world with claims cannot be deleted, remove its map instead", nil)
	}
	deleted, err := s.storage.Queries().DeleteSeasonWorld(ctx, worldID)
	if err != nil {
		return utils.NewInternalServerError("failed to delete season world", err)
	}
	if !deleted {
		return utils.NewNotFoundError("world not found", nil)
	}
	return nil
}

// seasonWorldWriteError maps a slug taken in the season to a conflict.
func seasonWorldWriteError(msg string, err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return utils.NewConflictError("the season already has a world with this slug", err)
	}
	return utils.NewInternalServerError(msg, err)
}
