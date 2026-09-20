package services

import (
	"context"
	stdsql "database/sql"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils"
)

type SeasonService interface {
	GetSeasonByID(ctx context.Context, seasonID uuid.UUID) (*domain.Season, error)
	// GetSeasons returns every season, the newest number first.
	GetSeasons(ctx context.Context) ([]*domain.Season, error)
}

type seasonService struct {
	storage storage.MainStorage
}

func NewSeasonService(storage storage.MainStorage) SeasonService {
	return &seasonService{storage: storage}
}

func (s *seasonService) GetSeasonByID(ctx context.Context, seasonID uuid.UUID) (*domain.Season, error) {
	season, err := s.storage.Queries().FindSeasonByID(ctx, seasonID)
	if err == stdsql.ErrNoRows {
		return nil, utils.NewNotFoundError("season not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get season by id", err)
	}
	return season, nil
}

func (s *seasonService) GetSeasons(ctx context.Context) ([]*domain.Season, error) {
	seasons, err := s.storage.Queries().FindSeasons(ctx)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get seasons", err)
	}
	return seasons, nil
}
