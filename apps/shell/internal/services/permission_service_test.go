package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/clients"
)

type stubLuckpermsStorage struct {
	err     error
	written bool
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
