package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils"
)

type FlectoneService interface {
	CountPlaytimeByMinecaftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error)
}

type flectoneService struct {
	storage storage.FlectoneStorage
}

func NewFlectoneService(storage storage.FlectoneStorage) FlectoneService {
	return &flectoneService{
		storage: storage,
	}
}

func (s *flectoneService) CountPlaytimeByMinecaftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	onlineMap, err := s.storage.Queries().FindOnlineByMinecraftUUIDs(ctx, mcUUIDs)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to count online by minecraft uuid", err)
	}
	return onlineMap, nil
}
