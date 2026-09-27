package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type fakeNavAuthStorage struct {
	required map[uuid.UUID]bool
	saved    map[uuid.UUID]string
}

func (s *fakeNavAuthStorage) FindCredentialsRequired(_ context.Context, mcUUID uuid.UUID) (bool, bool, error) {
	required, found := s.required[mcUUID]
	return required, found, nil
}

func (s *fakeNavAuthStorage) SavePassword(_ context.Context, mcUUID uuid.UUID, passwordBcrypt string) error {
	s.saved[mcUUID] = passwordBcrypt
	return nil
}

func TestAuthServiceSetPassword(t *testing.T) {
	const hash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	unlicensed, licensed, unknown := uuid.New(), uuid.New(), uuid.New()

	tests := []struct {
		name     string
		mcUUID   uuid.UUID
		username string
		hash     string
		wantErr  error
	}{
		{name: "unlicensed player gets the hash", mcUUID: unlicensed, username: "Steve", hash: hash},
		{name: "licensed player has no password", mcUUID: licensed, username: "Steve", hash: hash, wantErr: ErrPlayerLicensed},
		{name: "unknown player never registered", mcUUID: unknown, username: "Steve", hash: hash, wantErr: ErrPlayerNotRegistered},
		{name: "plain password is refused", mcUUID: unlicensed, username: "Steve", hash: "hunter2", wantErr: ErrInvalidPasswordHash},
		{name: "username that alters the command", mcUUID: unlicensed, username: "Steve; stop", hash: hash, wantErr: ErrInvalidUsername},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &fakeNavAuthStorage{
				required: map[uuid.UUID]bool{unlicensed: true, licensed: false},
				saved:    map[uuid.UUID]string{},
			}

			console := &outputConsole{}

			err := NewAuthService(storage, console).SetPassword(context.Background(), tt.mcUUID, tt.username, tt.hash)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if saved, ok := storage.saved[tt.mcUUID]; (tt.wantErr == nil) != ok || (ok && saved != tt.hash) {
				t.Errorf("saved = %q, %v", saved, ok)
			}
			wantCommands := 0
			if tt.wantErr == nil {
				wantCommands = 1
			}
			if len(console.commands) != wantCommands || (wantCommands == 1 && !strings.HasPrefix(console.commands[0], "kick Steve ")) {
				t.Errorf("commands = %v, want a kick only after a saved password", console.commands)
			}
		})
	}
}
