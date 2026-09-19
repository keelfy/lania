package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/clients"
	"github.com/lania-smp/shell/internal/config"
	"github.com/lania-smp/shell/internal/logger"
	"github.com/lania-smp/shell/internal/storage"
)

type PermissionService interface {
	// GetPlayerGroups returns an entry for every requested player.
	GetPlayerGroups(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID][]string, error)
	SetPlayerPrefix(ctx context.Context, mcUUID uuid.UUID, prefix string) error
}

type permissionService struct {
	luckpermsStorage storage.LuckpermsStorage
	console          clients.Console
}

func NewPermissionService(luckpermsStorage storage.LuckpermsStorage, console clients.Console) PermissionService {
	return &permissionService{
		luckpermsStorage: luckpermsStorage,
		console:          console,
	}
}

const (
	groupNodePrefix  = "group."
	prefixNodePrefix = "prefix."
	// prefixPriority is the LuckPerms weight of prefixes set by the website.
	prefixPriority = "100"
)

func (s *permissionService) GetPlayerGroups(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID][]string, error) {
	nodes, err := s.luckpermsStorage.FindPermissionsWithPrefix(ctx, mcUUIDs, groupNodePrefix)
	if err != nil {
		return nil, err
	}

	groups := make(map[uuid.UUID][]string, len(mcUUIDs))
	for _, mcUUID := range mcUUIDs {
		names := make([]string, 0, len(nodes[mcUUID]))
		for _, node := range nodes[mcUUID] {
			names = append(names, strings.TrimPrefix(node, groupNodePrefix))
		}
		groups[mcUUID] = names
	}
	return groups, nil
}

func (s *permissionService) SetPlayerPrefix(ctx context.Context, mcUUID uuid.UUID, prefix string) error {
	node := prefixNodePrefix + prefixPriority + "." + prefix
	logger.Debugf(ctx, "setting prefix node for %s: %s", mcUUID.String(), node)
	if err := s.luckpermsStorage.ReplacePermissionsWithPrefix(ctx, mcUUID, prefixNodePrefix, node); err != nil {
		return err
	}

	s.syncServer(ctx)
	return nil
}

// syncServer makes the running server pick up the database change. It is best
// effort: the database is the source of truth and the server loads it on start,
// so an unreachable server must not fail the write.
func (s *permissionService) syncServer(ctx context.Context) {
	command := config.GetLuckpermsSyncCommand()
	output, err := s.console.Execute(ctx, command)
	switch {
	case errors.Is(err, clients.ErrConsoleDisabled):
		logger.Debugf(ctx, "skipping %q: rcon is not configured", command)
	case err != nil:
		logger.Warnf(ctx, "failed to run %q, in-game prefix stays stale until the next sync: %v", command, err)
	default:
		logger.Debugf(ctx, "%q: %s", command, output)
	}
}
