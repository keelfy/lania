package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/storage"
)

type WhitelistService interface {
	AddPlayer(ctx context.Context, mcUUID uuid.UUID, username string) error
	RemovePlayer(ctx context.Context, mcUUID uuid.UUID) error
}

type whitelistService struct {
	whitelistStorage storage.WhitelistStorage
}

func NewWhitelistService(whitelistStorage storage.WhitelistStorage) WhitelistService {
	return &whitelistService{whitelistStorage: whitelistStorage}
}

func (s *whitelistService) AddPlayer(ctx context.Context, mcUUID uuid.UUID, username string) error {
	return s.whitelistStorage.Upsert(ctx, mcUUID, username)
}

func (s *whitelistService) RemovePlayer(ctx context.Context, mcUUID uuid.UUID) error {
	return s.whitelistStorage.Delete(ctx, mcUUID)
}
