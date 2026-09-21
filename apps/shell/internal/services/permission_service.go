package services

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/clients"
	"github.com/lania-smp/shell/internal/config"
	"github.com/lania-smp/shell/internal/logger"
	"github.com/lania-smp/shell/internal/storage"
)

type PermissionService interface {
	// GetPlayerGroups returns an entry for every requested player.
	GetPlayerGroups(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID][]string, error)
	// ListPlayersByGroups returns players that belong to at least one of the groups.
	ListPlayersByGroups(ctx context.Context, groups []string) (uuid.UUIDs, error)
	// SetPlayerPrefix replaces the player's chat prefix with a console command. An empty prefix removes it.
	// It fails when the server cannot be reached, because the console is the only way the prefix is written.
	SetPlayerPrefix(ctx context.Context, mcUUID uuid.UUID, prefix string) error
	// SetPlayerRoles makes every player of roles belong to exactly the role group it maps to, among roleGroups.
	// An empty group leaves the player in no role group. A group that is not in roleGroups is an error.
	SetPlayerRoles(ctx context.Context, roleGroups []string, roles map[uuid.UUID]string) error
}

// ErrInvalidPrefix means the prefix cannot be passed to the console as one quoted argument.
var ErrInvalidPrefix = errors.New("invalid prefix")

// ErrUnknownRoleGroup means a role maps to a group that is not one of the role groups.
var ErrUnknownRoleGroup = errors.New("group is not a role group")

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
	groupNodePrefix = "group."
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

func (s *permissionService) ListPlayersByGroups(ctx context.Context, groups []string) (uuid.UUIDs, error) {
	nodes := make([]string, len(groups))
	for i, group := range groups {
		nodes[i] = groupNodePrefix + group
	}
	return s.luckpermsStorage.FindPlayersWithPermissions(ctx, nodes)
}

func (s *permissionService) SetPlayerPrefix(ctx context.Context, mcUUID uuid.UUID, prefix string) error {
	// The prefix is quoted into the command, so a quote, a backslash or a control character would alter it.
	if strings.ContainsFunc(prefix, func(r rune) bool { return r == '"' || r == '\\' || unicode.IsControl(r) }) {
		return fmt.Errorf("%w: %q", ErrInvalidPrefix, prefix)
	}

	commands := []string{fmt.Sprintf("lp user %s meta clear prefix", mcUUID)}
	if prefix != "" {
		commands = append(commands, fmt.Sprintf("lp user %s meta addprefix %s \"%s\"", mcUUID, prefixPriority, prefix))
	}
	for _, command := range commands {
		output, err := s.console.Execute(ctx, command)
		if err != nil {
			return fmt.Errorf("failed to run %q: %w", command, err)
		}
		logger.Debugf(ctx, "%q: %s", command, output)
	}
	return nil
}

func (s *permissionService) SetPlayerRoles(ctx context.Context, roleGroups []string, roles map[uuid.UUID]string) error {
	if len(roles) == 0 {
		return nil
	}

	remove := make([]string, len(roleGroups))
	for i, group := range roleGroups {
		remove[i] = groupNodePrefix + group
	}

	replacements := make([]storage.PermissionReplacement, 0, len(roles))
	for mcUUID, group := range roles {
		replacement := storage.PermissionReplacement{MinecraftUUID: mcUUID, Remove: remove}
		if group != "" {
			if !slices.Contains(roleGroups, group) {
				return fmt.Errorf("%w: %q", ErrUnknownRoleGroup, group)
			}
			replacement.Add = []string{groupNodePrefix + group}
		}
		replacements = append(replacements, replacement)
	}

	if err := s.luckpermsStorage.ReplacePermissions(ctx, replacements); err != nil {
		return err
	}

	s.syncServer(ctx)
	return nil
}

// syncServer makes the running server pick up the database change. It is best
// effort: the database is the source of truth and the server loads it on start,
// so an unreachable server must not fail the write.
// Prefixes do not go through here: they are written by the server itself.
func (s *permissionService) syncServer(ctx context.Context) {
	command := config.GetLuckpermsSyncCommand()
	output, err := s.console.Execute(ctx, command)
	switch {
	case errors.Is(err, clients.ErrConsoleDisabled):
		logger.Debugf(ctx, "skipping %q: rcon is not configured", command)
	case err != nil:
		logger.Warnf(ctx, "failed to run %q, in-game permissions stay stale until the next sync: %v", command, err)
	default:
		logger.Debugf(ctx, "%q: %s", command, output)
	}
}
