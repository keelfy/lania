package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/domain"
	"github.com/lania-smp/shell/internal/storage"
)

type PlayerService interface {
	// GetOnlineStatus returns an entry for every requested player.
	GetOnlineStatus(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error)
	// GetPlaytime returns an entry for every requested player.
	GetPlaytime(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*domain.Playtime, error)
	// ListOnlinePlayers returns every player that is online right now.
	ListOnlinePlayers(ctx context.Context) (uuid.UUIDs, error)
	// ListChangedPlaytimes returns playtime of players whose last session ended at or after sinceMs.
	ListChangedPlaytimes(ctx context.Context, sinceMs int64) (map[uuid.UUID]*domain.Playtime, error)
}

type playerService struct {
	planStorage     storage.PlanStorage
	flectoneStorage storage.FlectoneStorage
}

func NewPlayerService(planStorage storage.PlanStorage, flectoneStorage storage.FlectoneStorage) PlayerService {
	return &playerService{
		planStorage:     planStorage,
		flectoneStorage: flectoneStorage,
	}
}

func (s *playerService) GetOnlineStatus(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	online, err := s.flectoneStorage.FindOnline(ctx, mcUUIDs)
	if err != nil {
		return nil, err
	}
	for _, mcUUID := range mcUUIDs {
		if _, ok := online[mcUUID]; !ok {
			online[mcUUID] = false
		}
	}
	return online, nil
}

func (s *playerService) GetPlaytime(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*domain.Playtime, error) {
	playtimes, err := s.planStorage.FindPlaytimes(ctx, mcUUIDs)
	if err != nil {
		return nil, err
	}
	for _, mcUUID := range mcUUIDs {
		if _, ok := playtimes[mcUUID]; !ok {
			playtimes[mcUUID] = &domain.Playtime{}
		}
	}
	return playtimes, nil
}

func (s *playerService) ListChangedPlaytimes(ctx context.Context, sinceMs int64) (map[uuid.UUID]*domain.Playtime, error) {
	return s.planStorage.FindPlaytimesChangedSince(ctx, sinceMs)
}

func (s *playerService) ListOnlinePlayers(ctx context.Context) (uuid.UUIDs, error) {
	return s.flectoneStorage.FindOnlineUUIDs(ctx)
}
