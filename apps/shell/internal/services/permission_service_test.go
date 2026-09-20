package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/clients"
	"github.com/lania-smp/shell/internal/storage"
)

type stubLuckpermsStorage struct {
	err          error
	written      bool
	replacements []storage.PermissionReplacement
}

func (s *stubLuckpermsStorage) FindPermissionsWithPrefix(context.Context, uuid.UUIDs, string) (map[uuid.UUID][]string, error) {
	return nil, nil
}

func (s *stubLuckpermsStorage) FindPlayersWithPermissions(context.Context, []string) (uuid.UUIDs, error) {
	return nil, nil
}

func (s *stubLuckpermsStorage) ReplacePermissionsWithPrefix(context.Context, uuid.UUID, string, string) error {
	if s.err != nil {
		return s.err
	}
	s.written = true
	return nil
}

func (s *stubLuckpermsStorage) ReplacePermissions(_ context.Context, replacements []storage.PermissionReplacement) error {
	if s.err != nil {
		return s.err
	}
	s.replacements = replacements
	return nil
}

type stubConsole struct {
	err      error
	commands []string
}

func (c *stubConsole) Execute(_ context.Context, command string) (string, error) {
	c.commands = append(c.commands, command)
	return "", c.err
}

func TestSetPlayerPrefixSync(t *testing.T) {
	tests := []struct {
		name         string
		storageErr   error
		consoleErr   error
		wantErr      bool
		wantSyncRuns int
	}{
		{name: "syncs after write", wantSyncRuns: 1},
		{name: "unreachable server does not fail the write", consoleErr: errors.New("connection refused"), wantSyncRuns: 1},
		{name: "disabled rcon does not fail the write", consoleErr: clients.ErrConsoleDisabled, wantSyncRuns: 1},
		{name: "failed write skips sync", storageErr: errors.New("db down"), wantErr: true, wantSyncRuns: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &stubLuckpermsStorage{err: tt.storageErr}
			console := &stubConsole{err: tt.consoleErr}
			service := NewPermissionService(store, console)

			err := service.SetPlayerPrefix(context.Background(), uuid.New(), "<red>[A]")

			if (err != nil) != tt.wantErr {
				t.Errorf("SetPlayerPrefix() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(console.commands) != tt.wantSyncRuns {
				t.Errorf("sync commands = %v, want %d", console.commands, tt.wantSyncRuns)
			}
			if !tt.wantErr && !store.written {
				t.Errorf("prefix must be written")
			}
		})
	}
}

func TestSetPlayerRoles(t *testing.T) {
	admin, player := uuid.New(), uuid.New()
	roleGroups := []string{"owner", "admin", "mod"}
	ctx := context.Background()

	t.Run("swaps role groups and syncs the server once", func(t *testing.T) {
		store := &stubLuckpermsStorage{}
		console := &stubConsole{}
		service := NewPermissionService(store, console)

		err := service.SetPlayerRoles(ctx, roleGroups, map[uuid.UUID]string{admin: "admin", player: ""})
		if err != nil {
			t.Fatalf("SetPlayerRoles() error = %v", err)
		}

		if len(store.replacements) != 2 || len(console.commands) != 1 {
			t.Fatalf("replacements = %v, sync commands = %v, want 2 replacements and 1 sync", store.replacements, console.commands)
		}
		for _, replacement := range store.replacements {
			if len(replacement.Remove) != 3 || replacement.Remove[0] != "group.owner" {
				t.Errorf("remove = %v, want every role group node", replacement.Remove)
			}
			switch replacement.MinecraftUUID {
			case admin:
				if len(replacement.Add) != 1 || replacement.Add[0] != "group.admin" {
					t.Errorf("admin add = %v, want group.admin", replacement.Add)
				}
			case player:
				if len(replacement.Add) != 0 {
					t.Errorf("player add = %v, want none", replacement.Add)
				}
			}
		}
	})

	t.Run("no roles touch nothing", func(t *testing.T) {
		store := &stubLuckpermsStorage{}
		console := &stubConsole{}
		if err := NewPermissionService(store, console).SetPlayerRoles(ctx, roleGroups, nil); err != nil {
			t.Fatalf("SetPlayerRoles() error = %v", err)
		}
		if store.replacements != nil || len(console.commands) != 0 {
			t.Errorf("replacements = %v, sync commands = %v, want nothing", store.replacements, console.commands)
		}
	})

	t.Run("a group outside the role groups is refused", func(t *testing.T) {
		store := &stubLuckpermsStorage{}
		err := NewPermissionService(store, &stubConsole{}).SetPlayerRoles(ctx, roleGroups, map[uuid.UUID]string{admin: "vip"})
		if !errors.Is(err, ErrUnknownRoleGroup) || store.replacements != nil {
			t.Errorf("error = %v, replacements = %v, want ErrUnknownRoleGroup and no write", err, store.replacements)
		}
	})

	t.Run("failed write skips sync", func(t *testing.T) {
		console := &stubConsole{}
		err := NewPermissionService(&stubLuckpermsStorage{err: errors.New("db down")}, console).SetPlayerRoles(ctx, roleGroups, map[uuid.UUID]string{admin: "admin"})
		if err == nil || len(console.commands) != 0 {
			t.Errorf("error = %v, sync commands = %v, want an error and no sync", err, console.commands)
		}
	})
}
