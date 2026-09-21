package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/utils"
)

// errSeasonHasNoShell means the season has no shell address, so its server cannot be reached.
var errSeasonHasNoShell = errors.New("season has no shell address")

// MinecraftService reads and changes state of the season servers through their shell services.
//
// Online status, roles and chat prefixes belong to the player, not to a season. Online status is read from
// every active server. A role and a prefix are written to every active server. A whitelist and playtime
// belong to one season and use its own shell.
type MinecraftService interface {
	// GetOnlineStatusByMinecraftUUIDs marks a player online when the player is on at least one active server.
	GetOnlineStatusByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error)
	// ListOnlineMinecraftUUIDs returns players that are online on at least one active server.
	ListOnlineMinecraftUUIDs(ctx context.Context) (uuid.UUIDs, error)
	// SetPrefixByMinecraftUUID writes the prefix to every active server.
	SetPrefixByMinecraftUUID(ctx context.Context, mcUUID uuid.UUID, prefix string) error
	// SetPlayerRoles writes the roles to every active server. Every server is tried, even when one fails.
	SetPlayerRoles(ctx context.Context, roles map[uuid.UUID]domain.Role) error
	// SetPrefixInSeason writes the prefix to the server of one season. It does nothing when the season has
	// no shell address, because it has no server yet.
	SetPrefixInSeason(ctx context.Context, seasonID, mcUUID uuid.UUID, prefix string) error
	// SetPlayerRolesInSeason writes the roles to the server of one season. It does nothing when the season has
	// no shell address, because it has no server yet.
	SetPlayerRolesInSeason(ctx context.Context, seasonID uuid.UUID, roles map[uuid.UUID]domain.Role) error
	// AddToWhitelist does nothing when the season has no shell address, because it has no server yet.
	AddToWhitelist(ctx context.Context, seasonID uuid.UUID, profile *domain.Profile) error
	// RemoveFromWhitelist does nothing when the season has no shell address, because it has no server yet.
	RemoveFromWhitelist(ctx context.Context, seasonID uuid.UUID, profile *domain.Profile) error
	// ListChangedPlaytimes returns playtime of players whose last session in the season ended at or after sinceMs.
	ListChangedPlaytimes(ctx context.Context, seasonID uuid.UUID, sinceMs int64) (map[uuid.UUID]*domain.Playtime, error)
}

type minecraftService struct {
	seasonService SeasonService
	shellPool     clients.ShellPool
}

func NewMinecraftService(seasonService SeasonService, shellPool clients.ShellPool) MinecraftService {
	return &minecraftService{seasonService: seasonService, shellPool: shellPool}
}

// seasonShell returns the client of the shell service that serves the season.
func (s *minecraftService) seasonShell(ctx context.Context, seasonID uuid.UUID) (clients.ShellAPI, error) {
	season, err := s.seasonService.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return nil, err
	}
	if season.ShellAddress == nil {
		return nil, errSeasonHasNoShell
	}
	api, err := s.shellPool.Get(*season.ShellAddress)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to connect to season shell", err)
	}
	return api, nil
}

// activeShells returns one client for every shell service that serves an active season.
// Seasons that share a shell address give it once.
func (s *minecraftService) activeShells(ctx context.Context) ([]clients.ShellAPI, error) {
	seasons, err := s.seasonService.GetSeasons(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	apis := make([]clients.ShellAPI, 0, len(seasons))
	for _, season := range seasons {
		if !season.IsActive || season.ShellAddress == nil {
			continue
		}
		if _, ok := seen[*season.ShellAddress]; ok {
			continue
		}
		seen[*season.ShellAddress] = struct{}{}

		api, err := s.shellPool.Get(*season.ShellAddress)
		if err != nil {
			return nil, utils.NewInternalServerError("failed to connect to season shell", err)
		}
		apis = append(apis, api)
	}
	if len(apis) == 0 {
		return nil, utils.NewInternalServerError("no active season has a shell address", nil)
	}
	return apis, nil
}

// readFromActiveShells runs read on every active shell. A shell that fails is skipped with a warning,
// so one server down does not hide the players of the others. It fails only when every shell fails.
func (s *minecraftService) readFromActiveShells(ctx context.Context, what string, read func(clients.ShellAPI) error) error {
	apis, err := s.activeShells(ctx)
	if err != nil {
		return err
	}

	var failures []error
	for _, api := range apis {
		if err := read(api); err != nil {
			logger.Warnf(ctx, "[SHELL] Failed to %s on one server: %v", what, err)
			failures = append(failures, err)
		}
	}
	if len(failures) == len(apis) {
		return utils.NewInternalServerError("failed to "+what, errors.Join(failures...))
	}
	return nil
}

func (s *minecraftService) GetOnlineStatusByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	online := make(map[uuid.UUID]bool, len(mcUUIDs))
	err := s.readFromActiveShells(ctx, "get online status by minecraft uuids", func(api clients.ShellAPI) error {
		status, err := api.GetOnlineStatus(ctx, mcUUIDs)
		for mcUUID, isOnline := range status {
			online[mcUUID] = online[mcUUID] || isOnline
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return online, nil
}

func (s *minecraftService) ListOnlineMinecraftUUIDs(ctx context.Context) (uuid.UUIDs, error) {
	seen := make(map[uuid.UUID]struct{})
	online := uuid.UUIDs{}
	err := s.readFromActiveShells(ctx, "list online minecraft uuids", func(api clients.ShellAPI) error {
		players, err := api.ListOnlinePlayers(ctx)
		for _, mcUUID := range players {
			if _, ok := seen[mcUUID]; !ok {
				seen[mcUUID] = struct{}{}
				online = append(online, mcUUID)
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return online, nil
}

func (s *minecraftService) SetPrefixByMinecraftUUID(ctx context.Context, mcUUID uuid.UUID, prefix string) error {
	apis, err := s.activeShells(ctx)
	if err != nil {
		return err
	}

	// Every server is tried, so one server down does not keep the others on the old prefix.
	var failures []error
	for _, api := range apis {
		if err := api.SetPlayerPrefix(ctx, mcUUID, prefix); err != nil {
			failures = append(failures, err)
		}
	}
	if len(failures) > 0 {
		return utils.NewInternalServerError("failed to set prefix by minecraft uuid", errors.Join(failures...))
	}
	return nil
}

// roleGroupsFor maps the roles to the role groups of the server. A staff role is a group on the server,
// a player has no group.
func roleGroupsFor(roles map[uuid.UUID]domain.Role) (roleGroups []string, groups map[uuid.UUID]string) {
	roleGroups = make([]string, len(domain.StaffRoles))
	for i, role := range domain.StaffRoles {
		roleGroups[i] = string(role)
	}
	groups = make(map[uuid.UUID]string, len(roles))
	for mcUUID, role := range roles {
		group := ""
		if role != domain.RolePlayer {
			group = string(role)
		}
		groups[mcUUID] = group
	}
	return roleGroups, groups
}

func (s *minecraftService) SetPlayerRoles(ctx context.Context, roles map[uuid.UUID]domain.Role) error {
	apis, err := s.activeShells(ctx)
	if err != nil {
		return err
	}

	roleGroups, groups := roleGroupsFor(roles)

	// Every server is tried, so one server down does not keep the others on the old roles.
	var failures []error
	for _, api := range apis {
		if err := api.SetPlayerRoles(ctx, roleGroups, groups); err != nil {
			failures = append(failures, err)
		}
	}
	if len(failures) > 0 {
		return utils.NewInternalServerError("failed to set player roles", errors.Join(failures...))
	}
	return nil
}

func (s *minecraftService) SetPrefixInSeason(ctx context.Context, seasonID, mcUUID uuid.UUID, prefix string) error {
	api, err := s.seasonShell(ctx, seasonID)
	if errors.Is(err, errSeasonHasNoShell) {
		return nil
	} else if err != nil {
		return err
	}
	if err := api.SetPlayerPrefix(ctx, mcUUID, prefix); err != nil {
		return utils.NewInternalServerError("failed to set prefix by minecraft uuid", err)
	}
	return nil
}

func (s *minecraftService) SetPlayerRolesInSeason(ctx context.Context, seasonID uuid.UUID, roles map[uuid.UUID]domain.Role) error {
	api, err := s.seasonShell(ctx, seasonID)
	if errors.Is(err, errSeasonHasNoShell) {
		return nil
	} else if err != nil {
		return err
	}
	roleGroups, groups := roleGroupsFor(roles)
	if err := api.SetPlayerRoles(ctx, roleGroups, groups); err != nil {
		return utils.NewInternalServerError("failed to set player roles", err)
	}
	return nil
}

func (s *minecraftService) AddToWhitelist(ctx context.Context, seasonID uuid.UUID, profile *domain.Profile) error {
	api, err := s.seasonShell(ctx, seasonID)
	if errors.Is(err, errSeasonHasNoShell) {
		return nil
	} else if err != nil {
		return err
	}
	if err := api.AddToWhitelist(ctx, profile.MinecraftUUID, profile.MinecraftUsername); err != nil {
		return utils.NewInternalServerError("failed to add profile to whitelist", err)
	}
	return nil
}

func (s *minecraftService) RemoveFromWhitelist(ctx context.Context, seasonID uuid.UUID, profile *domain.Profile) error {
	api, err := s.seasonShell(ctx, seasonID)
	if errors.Is(err, errSeasonHasNoShell) {
		return nil
	} else if err != nil {
		return err
	}
	if err := api.RemoveFromWhitelist(ctx, profile.MinecraftUUID, profile.MinecraftUsername); err != nil {
		return utils.NewInternalServerError("failed to remove profile from whitelist", err)
	}
	return nil
}

func (s *minecraftService) ListChangedPlaytimes(ctx context.Context, seasonID uuid.UUID, sinceMs int64) (map[uuid.UUID]*domain.Playtime, error) {
	api, err := s.seasonShell(ctx, seasonID)
	if err != nil {
		return nil, err
	}
	return api.ListChangedPlaytimes(ctx, sinceMs)
}
