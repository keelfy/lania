package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils"
)

const roleSyncInterval = time.Minute

// RoleSyncService copies roles of profiles to the Minecraft servers. The profile is the source of truth.
type RoleSyncService interface {
	// RunRoleSync pushes recently changed roles to every active shell until ctx is done.
	RunRoleSync(ctx context.Context)
}

type roleSyncService struct {
	storage          storage.MainStorage
	minecraftService MinecraftService
}

func NewRoleSyncService(storage storage.MainStorage, minecraftService MinecraftService) RoleSyncService {
	return &roleSyncService{
		storage:          storage,
		minecraftService: minecraftService,
	}
}

func (s *roleSyncService) RunRoleSync(ctx context.Context) {
	wait := time.Duration(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}

		wait = roleSyncInterval
		if err := s.syncRoles(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf(ctx, "[ROLE SYNC] Failed to sync roles: %v", err)
		}
	}
}

// syncRoles pushes the roles that changed within the sync window. Pushing a role again is harmless, so a change
// stays in the window for several runs, and a shell that failed or was down gets it on the next run.
func (s *roleSyncService) syncRoles(ctx context.Context) error {
	// A window shorter than two runs could let a change fall between them.
	window := max(config.GetRoleSyncWindow(), 2*roleSyncInterval)
	changes, err := s.storage.Queries().FindProfileRolesChangedSince(ctx, window)
	if err != nil {
		return utils.NewInternalServerError("failed to find changed profile roles", err)
	}
	if len(changes) == 0 {
		return nil
	}

	roles := make(map[uuid.UUID]domain.Role, len(changes))
	for _, change := range changes {
		roles[change.MinecraftUUID] = change.Role
	}
	if err := s.minecraftService.SetPlayerRoles(ctx, roles); err != nil {
		return err
	}

	logger.Debugf(ctx, "[ROLE SYNC] Pushed %d roles changed within %s", len(roles), window)
	return nil
}
