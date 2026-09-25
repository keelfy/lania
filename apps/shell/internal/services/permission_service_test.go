package services

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/clients"
	"github.com/lania-smp/shell/internal/storage"
)

type stubLuckpermsStorage struct {
	err          error
	replacements []storage.PermissionReplacement
	savedPlayers map[uuid.UUID]string
}

func (s *stubLuckpermsStorage) SavePlayer(_ context.Context, mcUUID uuid.UUID, username string) error {
	if s.err != nil {
		return s.err
	}
	if s.savedPlayers == nil {
		s.savedPlayers = make(map[uuid.UUID]string)
	}
	s.savedPlayers[mcUUID] = username
	return nil
}

func (s *stubLuckpermsStorage) FindPermissionsWithPrefix(context.Context, uuid.UUIDs, string) (map[uuid.UUID][]string, error) {
	return nil, nil
}

func (s *stubLuckpermsStorage) FindPlayersWithPermissions(context.Context, []string) (uuid.UUIDs, error) {
	return nil, nil
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

func TestSetPlayerPrefix(t *testing.T) {
	mcUUID := uuid.MustParse("0f0c2a3e-5a4b-4d4c-9f6e-3b1a2c3d4e5f")
	clear := "lp user " + mcUUID.String() + " meta clear prefix"

	tests := []struct {
		name         string
		prefix       string
		consoleErr   error
		wantErr      error
		wantAnyErr   bool
		wantCommands []string
	}{
		{
			name:   "replaces the prefix",
			prefix: "<red>[A] ",
			wantCommands: []string{
				clear,
				"lp user " + mcUUID.String() + ` meta addprefix 100 "<red>[A] "`,
			},
		},
		{name: "empty prefix only clears", wantCommands: []string{clear}},
		{name: "unreachable server fails", prefix: "[A]", consoleErr: errors.New("connection refused"), wantAnyErr: true, wantCommands: []string{clear}},
		{name: "disabled rcon fails", prefix: "[A]", consoleErr: clients.ErrConsoleDisabled, wantErr: clients.ErrConsoleDisabled, wantCommands: []string{clear}},
		{name: "quote is rejected", prefix: `[A" meta clear`, wantErr: ErrInvalidPrefix},
		{name: "backslash is rejected", prefix: `[A\`, wantErr: ErrInvalidPrefix},
		{name: "line break is rejected", prefix: "[A]\nstop", wantErr: ErrInvalidPrefix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := &stubConsole{err: tt.consoleErr}
			service := NewPermissionService(&stubLuckpermsStorage{}, console)

			err := service.SetPlayerPrefix(context.Background(), mcUUID, tt.prefix)

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("SetPlayerPrefix() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantAnyErr && err == nil {
				t.Errorf("SetPlayerPrefix() must fail")
			}
			if tt.wantErr == nil && !tt.wantAnyErr && err != nil {
				t.Errorf("SetPlayerPrefix() unexpected error = %v", err)
			}
			if !slices.Equal(console.commands, tt.wantCommands) {
				t.Errorf("commands = %q, want %q", console.commands, tt.wantCommands)
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

func TestRegisterPlayer(t *testing.T) {
	ctx := context.Background()
	mcUUID := uuid.MustParse("0f0c2a3e-5a4b-4d4c-9f6e-3b1a2c3d4e5f")

	t.Run("saves the player", func(t *testing.T) {
		store := &stubLuckpermsStorage{}
		if err := NewPermissionService(store, &stubConsole{}).RegisterPlayer(ctx, mcUUID, "Steve_1"); err != nil {
			t.Fatalf("RegisterPlayer() error = %v", err)
		}
		if got := store.savedPlayers[mcUUID]; got != "Steve_1" {
			t.Errorf("saved username = %q, want %q", got, "Steve_1")
		}
	})

	t.Run("rejects an invalid username", func(t *testing.T) {
		store := &stubLuckpermsStorage{}
		err := NewPermissionService(store, &stubConsole{}).RegisterPlayer(ctx, mcUUID, "bad'name")
		if !errors.Is(err, ErrInvalidUsername) {
			t.Fatalf("RegisterPlayer() error = %v, want ErrInvalidUsername", err)
		}
		if len(store.savedPlayers) != 0 {
			t.Errorf("saved players = %v, want none", store.savedPlayers)
		}
	})
}
