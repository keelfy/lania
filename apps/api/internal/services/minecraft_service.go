package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/utils"
)

// MinecraftService reads and changes state of the Minecraft server through shell.
type MinecraftService interface {
	GetOnlineStatusByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error)
	GetGroupsByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID][]string, error)
	SetPrefixByMinecraftUUID(ctx context.Context, mcUUID uuid.UUID, prefix string) error
	AddToWhitelist(ctx context.Context, profile *domain.Profile) error
}

type minecraftService struct {
	shellAPI clients.ShellAPI
}

func NewMinecraftService(shellAPI clients.ShellAPI) MinecraftService {
	return &minecraftService{shellAPI: shellAPI}
}

func (s *minecraftService) GetOnlineStatusByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	online, err := s.shellAPI.GetOnlineStatus(ctx, mcUUIDs)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get online status by minecraft uuids", err)
	}
	return online, nil
}

func (s *minecraftService) GetGroupsByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID][]string, error) {
	groups, err := s.shellAPI.GetPlayerGroups(ctx, mcUUIDs)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get groups by minecraft uuids", err)
	}
	return groups, nil
}

func (s *minecraftService) SetPrefixByMinecraftUUID(ctx context.Context, mcUUID uuid.UUID, prefix string) error {
	if err := s.shellAPI.SetPlayerPrefix(ctx, mcUUID, prefix); err != nil {
		return utils.NewInternalServerError("failed to set prefix by minecraft uuid", err)
	}
	return nil
}

func (s *minecraftService) AddToWhitelist(ctx context.Context, profile *domain.Profile) error {
	if err := s.shellAPI.AddToWhitelist(ctx, profile.MinecraftUUID, profile.MinecraftUsername); err != nil {
		return utils.NewInternalServerError("failed to add profile to whitelist", err)
	}
	return nil
}
